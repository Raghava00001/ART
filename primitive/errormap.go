package primitive

import (
	"image"
	"math"
	"math/rand"
)

// ErrorMap maintains a coarse-grid representation of per-pixel error between
// the target image and the current canvas. It provides importance-weighted
// coordinate sampling so that new shapes are preferentially spawned in
// high-error regions.
type ErrorMap struct {
	GridW, GridH   int          // number of cells in x and y
	ImgW, ImgH     int          // full image dimensions in pixels
	CellW, CellH   int          // pixels per cell
	Errors         []float64    // per-cell total squared error
	CDF            []float64    // cumulative distribution function for sampling
	TotalError     float64      // sum of all cell errors
	Saliency       *SaliencyMap // edge gradient feature map
	FeatureIndices []int        // cells with prominent facial/edge features
	FeatureCDF     []float64    // CDF specifically for feature cells
	TotalFeatError float64      // sum of feature cell errors
	HasFeatures    bool         // true if prominent feature cells exist
}

// NewErrorMap creates an error map with approximately gridSize cells along
// each axis. The actual grid dimensions are rounded to divide the image
// evenly.
func NewErrorMap(imgW, imgH, gridSize int, sm *SaliencyMap) *ErrorMap {
	if gridSize <= 0 {
		gridSize = 32
	}
	cellW := imgW / gridSize
	if cellW < 1 {
		cellW = 1
	}
	cellH := imgH / gridSize
	if cellH < 1 {
		cellH = 1
	}
	gw := (imgW + cellW - 1) / cellW
	gh := (imgH + cellH - 1) / cellH

	em := &ErrorMap{
		GridW:    gw,
		GridH:    gh,
		ImgW:     imgW,
		ImgH:     imgH,
		CellW:    cellW,
		CellH:    cellH,
		Errors:   make([]float64, gw*gh),
		CDF:      make([]float64, gw*gh),
		Saliency: sm,
	}
	return em
}

// Update recomputes all cell errors by comparing target and current images.
func (em *ErrorMap) Update(target, current *image.RGBA) {
	for i := range em.Errors {
		em.Errors[i] = 0
	}

	w := em.ImgW
	h := em.ImgH

	for y := 0; y < h; y++ {
		gy := y / em.CellH
		if gy >= em.GridH {
			gy = em.GridH - 1
		}
		i := target.PixOffset(0, y)
		for x := 0; x < w; x++ {
			gx := x / em.CellW
			if gx >= em.GridW {
				gx = em.GridW - 1
			}

			tr := int(target.Pix[i])
			tg := int(target.Pix[i+1])
			tb := int(target.Pix[i+2])
			cr := int(current.Pix[i])
			cg := int(current.Pix[i+1])
			cb := int(current.Pix[i+2])
			i += 4

			dr := tr - cr
			dg := tg - cg
			db := tb - cb
			em.Errors[gy*em.GridW+gx] += float64(dr*dr + dg*dg + db*db)
		}
	}

	// Apply saliency boost to high-feature cells (eyes, lips, jawline)
	if em.Saliency != nil {
		for gy := 0; gy < em.GridH; gy++ {
			startY := gy * em.CellH
			endY := startY + em.CellH
			for gx := 0; gx < em.GridW; gx++ {
				startX := gx * em.CellW
				endX := startX + em.CellW
				s := em.Saliency.CellAverage(startX, startY, endX, endY)
				idx := gy*em.GridW + gx
				if s > 0.05 {
					em.Errors[idx] *= (1.0 + 2.0*s)
				}
			}
		}
	}

	em.rebuildCDF()
}

// UpdatePartial incrementally updates only the cells touched by the given
// scanlines. This is much faster than a full Update when only a small
// region changed.
func (em *ErrorMap) UpdatePartial(target, current *image.RGBA, lines []Scanline) {
	// Determine which grid cells are affected
	touched := make(map[int]bool)
	for _, line := range lines {
		gy := line.Y / em.CellH
		if gy >= em.GridH {
			gy = em.GridH - 1
		}
		if gy < 0 {
			continue
		}
		for x := line.X1; x <= line.X2; x++ {
			gx := x / em.CellW
			if gx >= em.GridW {
				gx = em.GridW - 1
			}
			if gx < 0 {
				continue
			}
			touched[gy*em.GridW+gx] = true
		}
	}

	// Recompute error for each touched cell
	for cellIdx := range touched {
		gy := cellIdx / em.GridW
		gx := cellIdx % em.GridW

		startX := gx * em.CellW
		startY := gy * em.CellH
		endX := startX + em.CellW
		endY := startY + em.CellH
		if endX > em.ImgW {
			endX = em.ImgW
		}
		if endY > em.ImgH {
			endY = em.ImgH
		}

		var total float64
		for y := startY; y < endY; y++ {
			i := target.PixOffset(startX, y)
			for x := startX; x < endX; x++ {
				tr := int(target.Pix[i])
				tg := int(target.Pix[i+1])
				tb := int(target.Pix[i+2])
				cr := int(current.Pix[i])
				cg := int(current.Pix[i+1])
				cb := int(current.Pix[i+2])
				i += 4

				dr := tr - cr
				dg := tg - cg
				db := tb - cb
				total += float64(dr*dr + dg*dg + db*db)
			}
		}

		if em.Saliency != nil {
			s := em.Saliency.CellAverage(startX, startY, endX, endY)
			if s > 0.05 {
				total *= (1.0 + 2.0*s)
			}
		}
		em.Errors[cellIdx] = total
	}

	em.rebuildCDF()
}

// rebuildCDF constructs the cumulative distribution functions from cell errors.
// Both global error CDF and feature-focused CDF are updated so sampling can target fine features.
func (em *ErrorMap) rebuildCDF() {
	em.TotalError = 0
	em.FeatureIndices = em.FeatureIndices[:0]
	em.FeatureCDF = em.FeatureCDF[:0]
	em.TotalFeatError = 0

	for i, e := range em.Errors {
		// Use linear Euclidean perceptual color error (sqrt of squared difference sum) + floor
		w := math.Sqrt(e) + 1.0
		em.TotalError += w
		em.CDF[i] = em.TotalError

		if em.Saliency != nil {
			gy := i / em.GridW
			gx := i % em.GridW
			startX := gx * em.CellW
			startY := gy * em.CellH
			endX := startX + em.CellW
			endY := startY + em.CellH
			s := em.Saliency.CellAverage(startX, startY, endX, endY)
			if s >= 0.65 {
				em.FeatureIndices = append(em.FeatureIndices, i)
				fw := math.Sqrt(e) + 1.0
				em.TotalFeatError += fw
				em.FeatureCDF = append(em.FeatureCDF, em.TotalFeatError)
			}
		}
	}
	em.HasFeatures = len(em.FeatureIndices) > 0 && em.TotalFeatError > 0
}

// Sample returns a random pixel coordinate (x, y) weighted by error.
// High-error cells are sampled proportionally more often.
func (em *ErrorMap) Sample(rnd *rand.Rand) (int, int) {
	// Binary search the CDF for a random value
	target := rnd.Float64() * em.TotalError
	lo, hi := 0, len(em.CDF)-1
	for lo < hi {
		mid := (lo + hi) / 2
		if em.CDF[mid] < target {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	// Convert cell index back to grid coordinates
	cellIdx := lo
	gy := cellIdx / em.GridW
	gx := cellIdx % em.GridW

	// Random pixel within the cell
	startX := gx * em.CellW
	startY := gy * em.CellH
	endX := startX + em.CellW
	endY := startY + em.CellH
	if startX >= em.ImgW {
		startX = em.ImgW - 1
	}
	if startY >= em.ImgH {
		startY = em.ImgH - 1
	}
	if endX > em.ImgW {
		endX = em.ImgW
	}
	if endY > em.ImgH {
		endY = em.ImgH
	}
	if endX <= startX {
		endX = startX + 1
	}
	if endY <= startY {
		endY = startY + 1
	}

	x := startX + rnd.Intn(endX-startX)
	y := startY + rnd.Intn(endY-startY)
	return x, y
}

// SampleFloat returns a random pixel coordinate as float64, weighted by error.
func (em *ErrorMap) SampleFloat(rnd *rand.Rand) (float64, float64) {
	x, y := em.Sample(rnd)
	return float64(x), float64(y)
}

// SampleFeature returns a random pixel coordinate specifically sampled from
// prominent facial feature and contour cells (eyes, pupils, lips, jawline).
func (em *ErrorMap) SampleFeature(rnd *rand.Rand) (int, int) {
	if !em.HasFeatures || len(em.FeatureCDF) == 0 {
		return em.Sample(rnd)
	}
	target := rnd.Float64() * em.TotalFeatError
	lo, hi := 0, len(em.FeatureCDF)-1
	for lo < hi {
		mid := (lo + hi) / 2
		if em.FeatureCDF[mid] < target {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	cellIdx := em.FeatureIndices[lo]
	gy := cellIdx / em.GridW
	gx := cellIdx % em.GridW

	startX := gx * em.CellW
	startY := gy * em.CellH
	endX := startX + em.CellW
	endY := startY + em.CellH
	if startX >= em.ImgW {
		startX = em.ImgW - 1
	}
	if startY >= em.ImgH {
		startY = em.ImgH - 1
	}
	if endX > em.ImgW {
		endX = em.ImgW
	}
	if endY > em.ImgH {
		endY = em.ImgH
	}
	if endX <= startX {
		endX = startX + 1
	}
	if endY <= startY {
		endY = startY + 1
	}

	x := startX + rnd.Intn(endX-startX)
	y := startY + rnd.Intn(endY-startY)
	return x, y
}
