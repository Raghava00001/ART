package primitive

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"testing"
)

// createTestImage creates a simple 64x64 synthetic RGBA test image
func createTestImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	// Fill background with dark slate
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{20, 24, 33, 255}}, image.ZP, draw.Src)
	// Draw a central bright square
	for y := 16; y < 48; y++ {
		for x := 16; x < 48; x++ {
			img.SetRGBA(x, y, color.RGBA{220, 80, 50, 255})
		}
	}
	return img
}

func TestShapeRasterization(t *testing.T) {
	target := createTestImage()
	worker := NewWorker(target)
	worker.Init(target, 100.0)

	shapeGenerators := []struct {
		name string
		gen  func() Shape
	}{
		{"Triangle", func() Shape { return NewRandomTriangle(worker) }},
		{"Rectangle", func() Shape { return NewRandomRectangle(worker) }},
		{"Ellipse", func() Shape { return NewRandomEllipse(worker) }},
		{"Circle", func() Shape { return NewRandomCircle(worker) }},
		{"RotatedRectangle", func() Shape { return NewRandomRotatedRectangle(worker) }},
		{"Quadratic", func() Shape { return NewRandomQuadratic(worker) }},
		{"RotatedEllipse", func() Shape { return NewRandomRotatedEllipse(worker) }},
		{"Polygon", func() Shape { return NewRandomPolygon(worker, 4, true) }},
		{"Cubic", func() Shape { return NewRandomCubic(worker) }},
	}

	for _, tc := range shapeGenerators {
		t.Run(tc.name, func(t *testing.T) {
			var shape Shape
			var lines []Scanline
			for attempts := 0; attempts < 10; attempts++ {
				shape = tc.gen()
				lines = shape.Rasterize()
				if len(lines) > 0 {
					break
				}
			}
			if len(lines) == 0 {
				t.Fatalf("Shape %s produced 0 scanlines after 10 attempts", tc.name)
			}
			svg := shape.SVG("fill=\"#ff0000\"")
			if len(svg) == 0 {
				t.Fatalf("Shape %s produced empty SVG", tc.name)
			}
			copyShape := shape.Copy()
			if copyShape == nil {
				t.Fatalf("Shape %s Copy() returned nil", tc.name)
			}
		})
	}
}

func TestModelStep(t *testing.T) {
	target := createTestImage()
	bg := Color{R: 20, G: 24, B: 33, A: 255}
	model := NewModel(target, bg, 64, 2)

	initialScore := model.Score
	if initialScore <= 0 {
		t.Fatalf("Expected initial score > 0, got %f", initialScore)
	}

	// Run step with Triangles
	model.Step(ShapeTypeTriangle, 128, 0)

	if len(model.Shapes) != 1 {
		t.Fatalf("Expected 1 shape, got %d", len(model.Shapes))
	}
	if model.Score > initialScore {
		t.Fatalf("Expected score to improve (decrease), got %f vs initial %f", model.Score, initialScore)
	}

	svg := model.SVG()
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Fatalf("Generated SVG does not contain valid SVG wrapper: %s", svg)
	}
}

func TestModelAllShapeModes(t *testing.T) {
	target := createTestImage()
	bg := Color{R: 20, G: 24, B: 33, A: 255}

	modes := []ShapeType{
		ShapeTypeAny,
		ShapeTypeTriangle,
		ShapeTypeRectangle,
		ShapeTypeEllipse,
		ShapeTypeCircle,
		ShapeTypeRotatedRectangle,
		ShapeTypeQuadratic,
		ShapeTypeRotatedEllipse,
		ShapeTypePolygon,
		ShapeTypeCubic,
	}

	for _, mode := range modes {
		t.Run(fmt.Sprintf("Mode_%d", int(mode)), func(t *testing.T) {
			m := NewModel(target, bg, 64, 1)
			m.Step(mode, 128, 0)
			if len(m.Shapes) != 1 {
				t.Errorf("Mode %d failed to add shape", int(mode))
			}
		})
	}
}

func TestSetTargetAndPyramid(t *testing.T) {
	target64 := createTestImage()
	bg := Color{R: 20, G: 24, B: 33, A: 255}
	model := NewModel(target64, bg, 64, 1)

	model.Step(ShapeTypeTriangle, 128, 0)
	score1 := model.Score

	// Refine to 128px target
	target128 := image.NewRGBA(image.Rect(0, 0, 128, 128))
	draw.Draw(target128, target128.Bounds(), &image.Uniform{color.RGBA{20, 24, 33, 255}}, image.ZP, draw.Src)
	model.SetTarget(target128)

	model.Step(ShapeTypeTriangle, 128, 0)
	if len(model.Shapes) != 2 {
		t.Fatalf("Expected 2 shapes after refinement, got %d", len(model.Shapes))
	}
	_ = score1
}
