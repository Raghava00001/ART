package primitive

import (
	"image"
	"math"
	"math/rand"
	"time"

	"github.com/golang/freetype/raster"
)

type Worker struct {
	W, H       int
	Target     *image.RGBA
	Current    *image.RGBA
	Buffer     *image.RGBA
	Rasterizer *raster.Rasterizer
	Lines      []Scanline
	Heatmap    *Heatmap
	Rnd        *rand.Rand
	Score      float64
	Counter    int
	ErrMap     *ErrorMap    // residual-weighted importance sampling map
	Progress   float64      // synthesis progress [0.0..1.0] for scale annealing
	Saliency   *SaliencyMap // edge gradient map for feature clarity
}

func NewWorker(target *image.RGBA) *Worker {
	w := target.Bounds().Size().X
	h := target.Bounds().Size().Y
	worker := Worker{}
	worker.W = w
	worker.H = h
	worker.Target = target
	worker.Buffer = image.NewRGBA(target.Bounds())
	worker.Rasterizer = raster.NewRasterizer(w, h)
	worker.Lines = make([]Scanline, 0, 4096) // TODO: based on height
	worker.Heatmap = NewHeatmap(w, h)
	worker.Rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	return &worker
}

func (worker *Worker) Init(current *image.RGBA, score float64) {
	worker.Current = current
	worker.Score = score
	worker.Counter = 0
	worker.Heatmap.Clear()
}

// ScaleFactor anneals from 1.0 (coarse broad strokes) down to ~0.20 (fine facial detail)
// as synthesis progress advances from 0% to 100%.
func (worker *Worker) ScaleFactor() float64 {
	p := worker.Progress
	if p <= 0 {
		return 1.0
	}
	if p >= 1.0 {
		return 0.20
	}
	// Smooth curve: high scale early, tapering into razor-sharp micro-primitives
	factor := 1.0 - 0.80*math.Pow(p, 1.2)
	if factor < 0.20 {
		return 0.20
	}
	return factor
}

// FeatureScaleFactor scales down shape radius when spawning directly over a high-saliency
// feature (such as eyes, nostrils, lips, or jawline) so the engine immediately adapts
// its brush size to delicate micro-strokes rather than averaging the region.
func (worker *Worker) FeatureScaleFactor(x, y int) float64 {
	base := worker.ScaleFactor()
	if worker.Saliency == nil {
		return base
	}
	s := worker.Saliency.SaliencyAt(x, y)
	if s > 0.15 {
		featureFactor := 1.0 - 0.80*math.Pow((s-0.15)/0.85, 0.7)
		if featureFactor < 0.20 {
			featureFactor = 0.20
		}
		return base * featureFactor
	}
	return base
}

// MutationStep returns the annealed Gaussian mutation magnitude for hill climbing.
func (worker *Worker) MutationStep(base float64) float64 {
	step := base * worker.ScaleFactor()
	if step < 2.0 {
		return 2.0
	}
	return step
}

// MutationStepFor returns the mutation step scaled to local feature detail.
func (worker *Worker) MutationStepFor(base float64, x, y int) float64 {
	step := base * worker.FeatureScaleFactor(x, y)
	if step < 1.0 {
		return 1.0
	}
	return step
}

// SamplePosition returns a spawn coordinate for a new shape.
// If an ErrorMap is available:
// - During the initial blocking phase (Progress <= 0.20), samples globally.
// - During detail refinement (Progress > 0.20), prioritizes prominent facial
//   feature cells (eyes, pupils, lips, jawline) 55% of the time.
func (worker *Worker) SamplePosition() (int, int) {
	if worker.ErrMap != nil {
		if worker.Progress > 0.20 && worker.ErrMap.HasFeatures && worker.Rnd.Float64() < 0.25 {
			return worker.ErrMap.SampleFeature(worker.Rnd)
		}
		return worker.ErrMap.Sample(worker.Rnd)
	}
	return worker.Rnd.Intn(worker.W), worker.Rnd.Intn(worker.H)
}

// SamplePositionFloat returns a spawn coordinate as float64.
func (worker *Worker) SamplePositionFloat() (float64, float64) {
	x, y := worker.SamplePosition()
	return float64(x), float64(y)
}

func (worker *Worker) Energy(shape Shape, alpha int) float64 {
	worker.Counter++
	lines := shape.Rasterize()
	// worker.Heatmap.Add(lines)
	color := computeColor(worker.Target, worker.Current, lines, alpha)
	copyLines(worker.Buffer, worker.Current, lines)
	drawLines(worker.Buffer, color, lines)
	return differencePartial(worker.Target, worker.Current, worker.Buffer, worker.Score, lines)
}

func (worker *Worker) BestHillClimbState(t ShapeType, a, n, age, m int) *State {
	var bestEnergy float64
	var bestState *State
	for i := 0; i < m; i++ {
		state := worker.BestRandomState(t, a, n)
		before := state.Energy()
		state = HillClimb(state, age).(*State)
		energy := state.Energy()
		vv("%dx random: %.6f -> %dx hill climb: %.6f\n", n, before, age, energy)
		if i == 0 || energy < bestEnergy {
			bestEnergy = energy
			bestState = state
		}
	}
	return bestState
}

func (worker *Worker) BestRandomState(t ShapeType, a, n int) *State {
	var bestEnergy float64
	var bestState *State
	for i := 0; i < n; i++ {
		state := worker.RandomState(t, a)
		energy := state.Energy()
		if i == 0 || energy < bestEnergy {
			bestEnergy = energy
			bestState = state
		}
	}
	return bestState
}

func (worker *Worker) RandomState(t ShapeType, a int) *State {
	switch t {
	default:
		if worker.Progress < 0.35 {
			// Early blocking: use solid planar primitives to establish curtains, shadows, and base tone
			planar := [...]ShapeType{
				ShapeTypeTriangle,
				ShapeTypePolygon,
				ShapeTypeRectangle,
				ShapeTypeRotatedRectangle,
				ShapeTypeEllipse,
				ShapeTypeRotatedEllipse,
			}
			return worker.RandomState(planar[worker.Rnd.Intn(len(planar))], a)
		}
		return worker.RandomState(ShapeType(worker.Rnd.Intn(9)+1), a)
	case ShapeTypeTriangle:
		return NewState(worker, NewRandomTriangle(worker), a)
	case ShapeTypeRectangle:
		return NewState(worker, NewRandomRectangle(worker), a)
	case ShapeTypeEllipse:
		return NewState(worker, NewRandomEllipse(worker), a)
	case ShapeTypeCircle:
		return NewState(worker, NewRandomCircle(worker), a)
	case ShapeTypeRotatedRectangle:
		return NewState(worker, NewRandomRotatedRectangle(worker), a)
	case ShapeTypeQuadratic:
		return NewState(worker, NewRandomQuadratic(worker), a)
	case ShapeTypeRotatedEllipse:
		return NewState(worker, NewRandomRotatedEllipse(worker), a)
	case ShapeTypePolygon:
		return NewState(worker, NewRandomPolygon(worker, 4, true), a)
	case ShapeTypeCubic:
		return NewState(worker, NewRandomCubic(worker), a)
	}
}
