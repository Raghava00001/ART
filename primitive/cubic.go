package primitive

import (
	"fmt"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/raster"
)

type Cubic struct {
	Worker *Worker
	X1, Y1 float64
	X2, Y2 float64
	X3, Y3 float64
	X4, Y4 float64
	Width  float64
}

func NewRandomCubic(worker *Worker) *Cubic {
	rnd := worker.Rnd
	x1, y1 := worker.SamplePositionFloat()
	scale := worker.FeatureScaleFactor(int(x1), int(y1))
	span := math.Max(8.0, 40.0*scale)
	half := span / 2
	x2 := x1 + rnd.Float64()*span - half
	y2 := y1 + rnd.Float64()*span - half
	x3 := x2 + rnd.Float64()*span - half
	y3 := y2 + rnd.Float64()*span - half
	x4 := x3 + rnd.Float64()*span - half
	y4 := y3 + rnd.Float64()*span - half
	width := math.Max(0.5, (rnd.Float64()*2.0+0.5)*scale)
	c := &Cubic{worker, x1, y1, x2, y2, x3, y3, x4, y4, width}
	c.Mutate()
	return c
}

func (c *Cubic) Draw(dc *gg.Context, scale float64) {
	dc.MoveTo(c.X1, c.Y1)
	dc.CubicTo(c.X2, c.Y2, c.X3, c.Y3, c.X4, c.Y4)
	dc.SetLineWidth(c.Width)
	dc.Stroke()
}

func (c *Cubic) SVG(attrs string) string {
	attrs = strings.Replace(attrs, "fill", "stroke", -1)
	return fmt.Sprintf(
		"<path %s fill=\"none\" d=\"M %f %f C %f %f, %f %f, %f %f\" stroke-width=\"%f\" />",
		attrs, c.X1, c.Y1, c.X2, c.Y2, c.X3, c.Y3, c.X4, c.Y4, c.Width)
}

func (c *Cubic) Copy() Shape {
	a := *c
	return &a
}

func (c *Cubic) Mutate() {
	const m = 16
	w := c.Worker.W
	h := c.Worker.H
	rnd := c.Worker.Rnd
	step := c.Worker.MutationStepFor(16, int(c.X1), int(c.Y1))
	wStep := math.Max(0.2, c.Worker.MutationStepFor(1.0, int(c.X1), int(c.Y1)))
	for {
		switch rnd.Intn(5) {
		case 0:
			c.X1 = clamp(c.X1+rnd.NormFloat64()*step, -m, float64(w-1+m))
			c.Y1 = clamp(c.Y1+rnd.NormFloat64()*step, -m, float64(h-1+m))
		case 1:
			c.X2 = clamp(c.X2+rnd.NormFloat64()*step, -m, float64(w-1+m))
			c.Y2 = clamp(c.Y2+rnd.NormFloat64()*step, -m, float64(h-1+m))
		case 2:
			c.X3 = clamp(c.X3+rnd.NormFloat64()*step, -m, float64(w-1+m))
			c.Y3 = clamp(c.Y3+rnd.NormFloat64()*step, -m, float64(h-1+m))
		case 3:
			c.X4 = clamp(c.X4+rnd.NormFloat64()*step, -m, float64(w-1+m))
			c.Y4 = clamp(c.Y4+rnd.NormFloat64()*step, -m, float64(h-1+m))
		case 4:
			c.Width = clamp(c.Width+rnd.NormFloat64()*wStep, 0.5, 16)
		}
		if c.Valid() {
			break
		}
	}
}

func (c *Cubic) Valid() bool {
	dx := c.X4 - c.X1
	dy := c.Y4 - c.Y1
	d14 := dx*dx + dy*dy
	return d14 > 4
}

func (c *Cubic) Rasterize() []Scanline {
	var path raster.Path
	p1 := fixp(c.X1, c.Y1)
	path.Start(p1)
	const n = 12
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		t1 := 1.0 - t
		x := t1*t1*t1*c.X1 + 3*t1*t1*t*c.X2 + 3*t1*t*t*c.X3 + t*t*t*c.X4
		y := t1*t1*t1*c.Y1 + 3*t1*t1*t*c.Y2 + 3*t1*t*t*c.Y3 + t*t*t*c.Y4
		path.Add1(fixp(x, y))
	}
	width := fix(c.Width)
	return strokePath(c.Worker, path, width, raster.RoundCapper, raster.RoundJoiner)
}
