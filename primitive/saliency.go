package primitive

import (
	"image"
	"math"
	"sort"
)

// SaliencyMap computes and stores the feature saliency magnitude of an image.
// It combines noise-suppressed Sobel edge detection with a semantic portrait/skin chrominance
// prior so that delicate facial structures (eyes, pupils, eyelids, nose, lips, jawline)
// receive highest priority rather than getting drowned out by background foliage or noise.
type SaliencyMap struct {
	W, H   int
	Values []float32 // normalized [0, 1] per pixel
	IsSkin []bool    // semantic skin / portrait mask
}

// SaliencyAt returns the saliency value at (x, y) clamped to bounds.
func (sm *SaliencyMap) SaliencyAt(x, y int) float64 {
	if sm == nil || len(sm.Values) == 0 {
		return 0
	}
	if x < 0 {
		x = 0
	} else if x >= sm.W {
		x = sm.W - 1
	}
	if y < 0 {
		y = 0
	} else if y >= sm.H {
		y = sm.H - 1
	}
	return float64(sm.Values[y*sm.W+x])
}

// NewSaliencyMap computes a portrait-aware feature saliency map from the target image.
func NewSaliencyMap(img *image.RGBA) *SaliencyMap {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	sm := &SaliencyMap{
		W:      w,
		H:      h,
		Values: make([]float32, w*h),
		IsSkin: make([]bool, w*h),
	}

	if w < 5 || h < 5 {
		return sm
	}

	// 1. Compute grayscale luminance and YCbCr skin chrominance per pixel
	lum := make([]float32, w*h)
	rawSkin := make([]bool, w*h)
	skinCount := 0

	for y := 0; y < h; y++ {
		offset := img.PixOffset(0, y)
		row := y * w
		for x := 0; x < w; x++ {
			r := float32(img.Pix[offset])
			g := float32(img.Pix[offset+1])
			b := float32(img.Pix[offset+2])
			offset += 4

			// ITU-R BT.601 luminance
			lum[row+x] = 0.299*r + 0.587*g + 0.114*b

			// Fast Hue, Saturation, Value computation for strict melanin clustering
			max := r
			if g > max {
				max = g
			}
			if b > max {
				max = b
			}
			min := r
			if g < min {
				min = g
			}
			if b < min {
				min = b
			}
			delta := max - min
			hue := float32(0)
			if delta > 0 {
				if r == max {
					hue = (g - b) / delta
				} else if g == max {
					hue = 2.0 + (b-r)/delta
				} else {
					hue = 4.0 + (r-g)/delta
				}
				hue *= 60.0
				if hue < 0 {
					hue += 360.0
				}
			}
			sat := float32(0)
			if max > 0 {
				sat = delta / max
			}

			// Strict human melanin/hemoglobin cluster (Hue 12°-31°, Sat >= 0.32):
			// Separates true facial skin from golden autumn foliage (>35°) and yellowish architecture
			isSkin := hue >= 12.0 && hue <= 31.0 &&
				sat >= 0.32 && sat <= 0.88 &&
				r > 80.0 && g > 40.0 && b > 20.0 &&
				(r-g) > 25.0 && (r-b) > 45.0
			if isSkin {
				rawSkin[row+x] = true
				skinCount++
			}
		}
	}

	// 2. Morphological dilation of skin mask:
	// Eyes, pupils, eyelids, lips, teeth, and jawlines sit directly inside or on the
	// boundary of skin regions. Dilating the skin mask by 8 pixels ensures facial features
	// and jawline contours are fully encompassed in the portrait region.
	portraitMask := make([]bool, w*h)
	hasPortrait := skinCount > (w * h / 500) // Subject present

	if hasPortrait {
		const rad = 3
		for y := 0; y < h; y++ {
			yMin := y - rad
			if yMin < 0 {
				yMin = 0
			}
			yMax := y + rad
			if yMax >= h {
				yMax = h - 1
			}

			for x := 0; x < w; x++ {
				if !rawSkin[y*w+x] {
					continue
				}
				xMin := x - rad
				if xMin < 0 {
					xMin = 0
				}
				xMax := x + rad
				if xMax >= w {
					xMax = w - 1
				}

				for dy := yMin; dy <= yMax; dy++ {
					rowD := dy * w
					for dx := xMin; dx <= xMax; dx++ {
						portraitMask[rowD+dx] = true
					}
				}
			}
		}
	}
	sm.IsSkin = portraitMask

	// 3. Pre-smooth luminance with a 3x3 filter to eliminate single-pixel foliage/leaf noise
	smooth := make([]float32, w*h)
	for y := 0; y < h; y++ {
		ym1 := y - 1
		if ym1 < 0 {
			ym1 = 0
		}
		yp1 := y + 1
		if yp1 >= h {
			yp1 = h - 1
		}
		rowM1 := ym1 * w
		row0 := y * w
		rowP1 := yp1 * w

		for x := 0; x < w; x++ {
			xm1 := x - 1
			if xm1 < 0 {
				xm1 = 0
			}
			xp1 := x + 1
			if xp1 >= w {
				xp1 = w - 1
			}

			// 3x3 Gaussian-like smoothing
			val := lum[rowM1+xm1] + 2*lum[rowM1+x] + lum[rowM1+xp1] +
				2*lum[row0+xm1] + 4*lum[row0+x] + 2*lum[row0+xp1] +
				lum[rowP1+xm1] + 2*lum[rowP1+x] + lum[rowP1+xp1]
			smooth[row0+x] = val / 16.0
		}
	}

	// 4. Sobel edge gradient computation on smoothed luminance
	rawGrad := make([]float32, w*h)
	var gradSamples []float64

	// Sample up to 2000 points to determine the 95th percentile gradient magnitude
	sampleStep := (w * h) / 2000
	if sampleStep < 1 {
		sampleStep = 1
	}

	for y := 1; y < h-1; y++ {
		ym1 := (y - 1) * w
		y0 := y * w
		yp1 := (y + 1) * w

		for x := 1; x < w-1; x++ {
			xm1 := x - 1
			xp1 := x + 1

			// Horizontal Sobel Gx
			gx := (smooth[ym1+xp1] + 2*smooth[y0+xp1] + smooth[yp1+xp1]) -
				(smooth[ym1+xm1] + 2*smooth[y0+xm1] + smooth[yp1+xm1])

			// Vertical Sobel Gy
			gy := (smooth[yp1+xm1] + 2*smooth[yp1+x] + smooth[yp1+xp1]) -
				(smooth[ym1+xm1] + 2*smooth[ym1+x] + smooth[yp1+xm1])

			g := float32(math.Sqrt(float64(gx*gx + gy*gy)))
			rawGrad[y0+x] = g

			idx := y0 + x
			if idx%sampleStep == 0 {
				gradSamples = append(gradSamples, float64(g))
			}
		}
	}

	// Robust 95th percentile normalization (avoids single-pixel outlier glints)
	p95 := 25.0
	if len(gradSamples) > 20 {
		sort.Float64s(gradSamples)
		p95 = gradSamples[int(float64(len(gradSamples))*0.95)]
		if p95 < 12.0 {
			p95 = 12.0
		}
	}
	invP95 := float32(1.0 / p95)

	// 5. Final Saliency Synthesis with Portrait Priority
	for y := 0; y < h; y++ {
		row := y * w
		for x := 0; x < w; x++ {
			idx := row + x
			g := rawGrad[idx] * invP95
			if g > 1.0 {
				g = 1.0
			}

			// Apply sub-linear gamma curve (gamma 0.65) so soft facial contours
			// (jawline, eyelids, lips) are elevated rather than suppressed
			val := float32(math.Pow(float64(g), 0.65))

			if hasPortrait && portraitMask[idx] {
				// Gentle boost for delicate facial contours (eyes, lips, jawline)
				feat := float32(math.Pow(float64(g), 0.55)) * 1.5
				if feat > 1.0 {
					feat = 1.0
				}
				val = feat
			}

			sm.Values[idx] = val
		}
	}

	return sm
}

// CellAverage returns the mean saliency value within a rectangular subregion [x1..x2, y1..y2].
func (sm *SaliencyMap) CellAverage(x1, y1, x2, y2 int) float64 {
	if sm == nil || len(sm.Values) == 0 {
		return 0
	}
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > sm.W {
		x2 = sm.W
	}
	if y2 > sm.H {
		y2 = sm.H
	}
	if x2 <= x1 || y2 <= y1 {
		return 0
	}

	var sum float64
	count := (x2 - x1) * (y2 - y1)
	for y := y1; y < y2; y++ {
		row := y * sm.W
		for x := x1; x < x2; x++ {
			sum += float64(sm.Values[row+x])
		}
	}
	return sum / float64(count)
}

// ApplyUserMask incorporates user-painted priority brush masks into the saliency field.
// Areas with mask values > 0.05 receive high importance so the synthesis engine
// dedicates shapes and tighter mutation radii to preserve intricate user-designated details.
func (sm *SaliencyMap) ApplyUserMask(mask image.Image) {
	if sm == nil || len(sm.Values) == 0 || mask == nil {
		return
	}
	bounds := mask.Bounds()
	mw := bounds.Dx()
	mh := bounds.Dy()
	if mw <= 0 || mh <= 0 {
		return
	}

	for y := 0; y < sm.H; y++ {
		my := y * mh / sm.H
		row := y * sm.W
		for x := 0; x < sm.W; x++ {
			mx := x * mw / sm.W
			r, g, b, a := mask.At(bounds.Min.X+mx, bounds.Min.Y+my).RGBA()
			if a > 0 {
				intensity := (float64(r+g+b) / (3.0 * 65535.0)) * (float64(a) / 65535.0)
				if intensity > 0.05 {
					idx := row + x
					current := float64(sm.Values[idx])
					boosted := math.Max(current, 0.70+0.30*intensity)
					if boosted > 1.0 {
						boosted = 1.0
					}
					sm.Values[idx] = float32(boosted)
				}
			}
		}
	}
}
