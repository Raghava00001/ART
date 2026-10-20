package primitive

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

type Model struct {
	Sw, Sh     int
	Scale      float64
	Background Color
	Target     *image.RGBA
	Current    *image.RGBA
	Context    *gg.Context
	Score      float64
	Shapes     []Shape
	Colors     []Color
	Scores     []float64
	Scales     []float64
	Workers    []*Worker
	ErrMap     *ErrorMap    // importance sampling error map
	Saliency   *SaliencyMap // edge gradient map for facial feature clarity
	Progress   float64      // synthesis progress [0.0..1.0] for scale annealing
	UserMask   image.Image  // optional user priority brush mask
}

func NewModel(target image.Image, background Color, size, numWorkers int) *Model {
	w := target.Bounds().Size().X
	h := target.Bounds().Size().Y
	aspect := float64(w) / float64(h)
	var sw, sh int
	var scale float64
	if aspect >= 1 {
		sw = size
		sh = int(float64(size) / aspect)
		scale = float64(size) / float64(w)
	} else {
		sw = int(float64(size) * aspect)
		sh = size
		scale = float64(size) / float64(h)
	}

	model := &Model{}
	model.Sw = sw
	model.Sh = sh
	model.Scale = scale
	model.Background = background
	model.Target = imageToRGBA(target)
	model.Current = uniformRGBA(target.Bounds(), background.NRGBA())
	model.Score = differenceFull(model.Target, model.Current)
	model.Context = model.newContext()

	// Compute Sobel saliency map for high-frequency edge & facial feature clarity
	model.Saliency = NewSaliencyMap(model.Target)

	// Initialize the error map with a 32x32 fine grid with edge weighting
	model.ErrMap = NewErrorMap(w, h, 32, model.Saliency)
	model.ErrMap.Update(model.Target, model.Current)

	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(model.Target)
		worker.Saliency = model.Saliency
		model.Workers = append(model.Workers, worker)
	}
	return model
}

// SetProgress updates the synthesis progress (0.0 to 1.0) and propagates to all workers
// for adaptive scale and mutation step annealing.
func (model *Model) SetProgress(p float64) {
	model.Progress = p
	for _, w := range model.Workers {
		w.Progress = p
	}
}

// SetUserMask incorporates a user-painted detail priority brush mask.
func (model *Model) SetUserMask(mask image.Image) {
	if mask == nil {
		return
	}
	model.UserMask = mask
	if model.Saliency != nil {
		model.Saliency.ApplyUserMask(mask)
		w := model.Target.Bounds().Dx()
		h := model.Target.Bounds().Dy()
		model.ErrMap = NewErrorMap(w, h, 32, model.Saliency)
		model.ErrMap.Update(model.Target, model.Current)
		for _, w := range model.Workers {
			w.Saliency = model.Saliency
			w.ErrMap = model.ErrMap
		}
	}
}

// SetTarget transitions the model to a new target image resolution (e.g. For coarse-to-fine pyramids).
// It resizes Current, adjusts Scale, rebuilds ErrorMap, and updates all workers.
func (model *Model) SetTarget(target image.Image) {
	w := target.Bounds().Size().X
	h := target.Bounds().Size().Y
	aspect := float64(w) / float64(h)
	if aspect >= 1 {
		model.Scale = float64(model.Sw) / float64(w)
	} else {
		model.Scale = float64(model.Sh) / float64(h)
	}

	model.Target = imageToRGBA(target)
	resizedCurrent := resize.Resize(uint(w), uint(h), model.Current, resize.Bilinear)
	model.Current = imageToRGBA(resizedCurrent)
	model.Score = differenceFull(model.Target, model.Current)

	model.Saliency = NewSaliencyMap(model.Target)
	if model.UserMask != nil {
		model.Saliency.ApplyUserMask(model.UserMask)
	}
	model.ErrMap = NewErrorMap(w, h, 32, model.Saliency)
	model.ErrMap.Update(model.Target, model.Current)

	numWorkers := len(model.Workers)
	model.Workers = model.Workers[:0]
	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(model.Target)
		worker.Saliency = model.Saliency
		worker.Progress = model.Progress
		model.Workers = append(model.Workers, worker)
	}
}

func (model *Model) newContext() *gg.Context {
	dc := gg.NewContext(model.Sw, model.Sh)
	dc.SetColor(model.Background.NRGBA())
	dc.Clear()
	return dc
}

func (model *Model) Frames(scoreDelta float64) []image.Image {
	var result []image.Image
	dc := model.newContext()
	result = append(result, imageToRGBA(dc.Image()))
	previous := 10.0
	for i, shape := range model.Shapes {
		c := model.Colors[i]
		scale := model.Scale
		if i < len(model.Scales) {
			scale = model.Scales[i]
		}
		dc.Push()
		dc.Scale(scale, scale)
		dc.Translate(0.5, 0.5)
		dc.SetRGBA255(c.R, c.G, c.B, c.A)
		shape.Draw(dc, scale)
		dc.Fill()
		dc.Pop()
		score := model.Scores[i]
		delta := previous - score
		if delta >= scoreDelta {
			previous = score
			result = append(result, imageToRGBA(dc.Image()))
		}
	}
	return result
}

func (model *Model) SVG() string {
	bg := model.Background
	var lines []string
	lines = append(lines, fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\">", model.Sw, model.Sh, model.Sw, model.Sh))
	lines = append(lines, fmt.Sprintf("<rect x=\"0\" y=\"0\" width=\"%d\" height=\"%d\" fill=\"#%02x%02x%02x\" />", model.Sw, model.Sh, bg.R, bg.G, bg.B))
	var lastScale float64 = -1
	for i, shape := range model.Shapes {
		scale := model.Scale
		if i < len(model.Scales) {
			scale = model.Scales[i]
		}
		if scale != lastScale {
			if lastScale >= 0 {
				lines = append(lines, "</g>")
			}
			lines = append(lines, fmt.Sprintf("<g transform=\"scale(%f) translate(0.5 0.5)\">", scale))
			lastScale = scale
		}
		c := model.Colors[i]
		attrs := "fill=\"#%02x%02x%02x\" fill-opacity=\"%f\""
		attrs = fmt.Sprintf(attrs, c.R, c.G, c.B, float64(c.A)/255)
		lines = append(lines, shape.SVG(attrs))
	}
	if lastScale >= 0 {
		lines = append(lines, "</g>")
	}
	lines = append(lines, "</svg>")
	return strings.Join(lines, "\n")
}

// RenderHighRes renders all synthesized vector shapes to an ultra-high-resolution print master (e.g. 3840px 4K, 5000px Poster, 7680px 8K).
func (model *Model) RenderHighRes(targetSize int, texture string, borderMat bool) *image.RGBA {
	if targetSize <= 0 {
		targetSize = 3840
	}
	aspect := float64(model.Sw) / float64(model.Sh)
	var tw, th int
	if aspect >= 1.0 {
		tw = targetSize
		th = int(math.Round(float64(targetSize) / aspect))
	} else {
		tw = int(math.Round(float64(targetSize) * aspect))
		th = targetSize
	}

	ratio := float64(tw) / float64(model.Sw)

	// If border mat is requested (gallery passe-partout), add a 4% minimalist margin
	margin := 0
	if borderMat {
		margin = maxInt(24, int(float64(tw)*0.04))
	}

	totalW := tw + margin*2
	totalH := th + margin*2

	dc := gg.NewContext(totalW, totalH)
	if borderMat {
		// Elegant deep obsidian gallery mat border
		dc.SetRGB255(14, 15, 18)
		dc.Clear()
	} else {
		dc.SetColor(model.Background.NRGBA())
		dc.Clear()
	}

	dc.Push()
	if margin > 0 {
		dc.Translate(float64(margin), float64(margin))
		// Fill artwork background inside margin
		dc.SetColor(model.Background.NRGBA())
		dc.DrawRectangle(0, 0, float64(tw), float64(th))
		dc.Fill()
	}

	for i, shape := range model.Shapes {
		c := model.Colors[i]
		scale := model.Scale
		if i < len(model.Scales) {
			scale = model.Scales[i]
		}
		dc.Push()
		dc.Scale(scale*ratio, scale*ratio)
		dc.Translate(0.5, 0.5)
		dc.SetRGBA255(c.R, c.G, c.B, c.A)
		shape.Draw(dc, scale)
		dc.Pop()
	}
	dc.Pop()

	img := dc.Image().(*image.RGBA)

	// Optional tactile texture simulation (canvas weave or fine paper grain)
	if texture == "paper" || texture == "canvas" {
		applyTactileTexture(img, totalW, totalH, texture)
	}

	return img
}

func applyTactileTexture(img *image.RGBA, w, h int, texture string) {
	stride := img.Stride
	pix := img.Pix
	for y := 0; y < h; y++ {
		row := y * stride
		for x := 0; x < w; x++ {
			idx := row + x*4
			var delta float64
			if texture == "canvas" {
				// Canvas weave: horizontal and vertical criss-cross thread pattern
				sx := math.Sin(float64(x) * 0.95)
				sy := math.Cos(float64(y) * 0.95)
				delta = (sx*sy)*2.4 + (float64((x*1973+y*9277)%11)-5.0)*0.25
			} else {
				// Fine art paper: subtle stippled organic tooth
				delta = float64((int(x*1973+y*9277)%17)-8) * 0.35
			}
			r := int(pix[idx]) + int(delta)
			g := int(pix[idx+1]) + int(delta)
			b := int(pix[idx+2]) + int(delta)
			if r < 0 {
				r = 0
			} else if r > 255 {
				r = 255
			}
			if g < 0 {
				g = 0
			} else if g > 255 {
				g = 255
			}
			if b < 0 {
				b = 0
			} else if b > 255 {
				b = 255
			}
			pix[idx] = uint8(r)
			pix[idx+1] = uint8(g)
			pix[idx+2] = uint8(b)
		}
	}
}

func (model *Model) Add(shape Shape, alpha int) {
	before := copyRGBA(model.Current)
	lines := shape.Rasterize()
	color := computeColor(model.Target, model.Current, lines, alpha)
	drawLines(model.Current, color, lines)
	score := differencePartial(model.Target, before, model.Current, model.Score, lines)

	model.Score = score
	model.Shapes = append(model.Shapes, shape)
	model.Colors = append(model.Colors, color)
	model.Scores = append(model.Scores, score)
	model.Scales = append(model.Scales, model.Scale)

	model.Context.Push()
	model.Context.Scale(model.Scale, model.Scale)
	model.Context.Translate(0.5, 0.5)
	model.Context.SetRGBA255(color.R, color.G, color.B, color.A)
	shape.Draw(model.Context, model.Scale)
	model.Context.Pop()

	// Incrementally update error map for the affected region
	if model.ErrMap != nil {
		model.ErrMap.UpdatePartial(model.Target, model.Current, lines)
	}
}

func (model *Model) Step(shapeType ShapeType, alpha, repeat int) int {
	state := model.runWorkers(shapeType, alpha, 1000, 100, 16)
	// state = HillClimb(state, 1000).(*State)
	model.Add(state.Shape, state.Alpha)

	for i := 0; i < repeat; i++ {
		state.Worker.Init(model.Current, model.Score)
		a := state.Energy()
		state = HillClimb(state, 100).(*State)
		b := state.Energy()
		if a == b {
			break
		}
		model.Add(state.Shape, state.Alpha)
	}

	// for _, w := range model.Workers[1:] {
	// 	model.Workers[0].Heatmap.AddHeatmap(w.Heatmap)
	// }
	// SavePNG("heatmap.png", model.Workers[0].Heatmap.Image(0.5))

	counter := 0
	for _, worker := range model.Workers {
		counter += worker.Counter
	}
	return counter
}

func (model *Model) runWorkers(t ShapeType, a, n, age, m int) *State {
	wn := len(model.Workers)
	ch := make(chan *State, wn)
	wm := m / wn
	if m%wn != 0 {
		wm++
	}
	for i := 0; i < wn; i++ {
		worker := model.Workers[i]
		worker.Init(model.Current, model.Score)
		// Share the error map, progress, and saliency map with each worker
		worker.ErrMap = model.ErrMap
		worker.Progress = model.Progress
		worker.Saliency = model.Saliency
		go model.runWorker(worker, t, a, n, age, wm, ch)
	}
	var bestEnergy float64
	var bestState *State
	for i := 0; i < wn; i++ {
		state := <-ch
		energy := state.Energy()
		if i == 0 || energy < bestEnergy {
			bestEnergy = energy
			bestState = state
		}
	}
	return bestState
}

func (model *Model) runWorker(worker *Worker, t ShapeType, a, n, age, m int, ch chan *State) {
	ch <- worker.BestHillClimbState(t, a, n, age, m)
}
