//go:build js && wasm
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"syscall/js"
	"time"

	"art/primitive"

	"github.com/nfnt/resize"
)

var (
	currentModel     *primitive.Model
	rawInputImage    image.Image
	currentMode      primitive.ShapeType
	currentAlpha     int
	currentRepeat    int
	totalTargetCount int
	completedShapes  int
	currentResolution int
	pyramidEnabled   bool
	stages           []pyramidStage
	currentStageIdx  int
	startTime        time.Time
	isEngineRunning  bool
)

type pyramidStage struct {
	res   int
	limit int
}

func initEngine(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return map[string]any{"error": "missing image data"}
	}

	dataURL := args[0].String()
	idx := strings.Index(dataURL, ",")
	if idx != -1 {
		dataURL = dataURL[idx+1:]
	}

	rawBytes, err := base64.StdEncoding.DecodeString(dataURL)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("base64 decode failed: %v", err)}
	}

	img, _, err := image.Decode(bytes.NewReader(rawBytes))
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("image decode failed: %v", err)}
	}
	rawInputImage = img

	mode := 1
	if len(args) > 1 {
		mode = args[1].Int()
	}
	currentMode = primitive.ShapeType(mode)

	count := 100
	if len(args) > 2 {
		count = args[2].Int()
	}
	totalTargetCount = count

	alpha := 128
	if len(args) > 3 {
		alpha = args[3].Int()
	}
	currentAlpha = alpha

	repeat := 1
	if len(args) > 4 {
		repeat = args[4].Int()
	}
	currentRepeat = repeat

	res := 256
	if len(args) > 5 {
		res = args[5].Int()
	}
	currentResolution = res

	pyramid := true
	if len(args) > 6 {
		pyramid = args[6].Bool()
	}
	pyramidEnabled = pyramid

	// Setup pyramid stages
	stages = nil
	currentStageIdx = 0
	var initialInput image.Image

	if pyramidEnabled && count >= 50 {
		stage0Res := res / 4
		if stage0Res < 64 {
			stage0Res = 64
		}
		stage1Res := res / 2
		if stage1Res < 96 {
			stage1Res = 96
		}
		stage2Res := res

		s0Limit := count / 3
		s1Limit := (count * 2) / 3
		stages = []pyramidStage{
			{res: stage0Res, limit: s0Limit},
			{res: stage1Res, limit: s1Limit},
			{res: stage2Res, limit: count},
		}
		initialInput = resize.Thumbnail(uint(stage0Res), uint(stage0Res), rawInputImage, resize.Bilinear)
	} else {
		initialInput = resize.Thumbnail(uint(res), uint(res), rawInputImage, resize.Bilinear)
	}

	bg := primitive.MakeColor(primitive.AverageImageColor(initialInput))
	outputSize := 1024

	// In WebAssembly single-threaded environment, 1 or 2 workers is ideal
	currentModel = primitive.NewModel(initialInput, bg, outputSize, 1)
	completedShapes = 0
	startTime = time.Now()
	isEngineRunning = true

	bgHex := fmt.Sprintf("#%02x%02x%02x", bg.R, bg.G, bg.B)

	return map[string]any{
		"width":  currentModel.Sw,
		"height": currentModel.Sh,
		"bg":     bgHex,
		"score":  currentModel.Score,
	}
}

func stepEngine(this js.Value, args []js.Value) any {
	if currentModel == nil || !isEngineRunning {
		return map[string]any{"done": true}
	}

	if completedShapes >= totalTargetCount {
		isEngineRunning = false
		return map[string]any{
			"done":     true,
			"total":    completedShapes,
			"score":    currentModel.Score,
			"duration": time.Since(startTime).Seconds(),
			"fullSvg":  currentModel.SVG(),
		}
	}

	// Advance pyramid stage if applicable
	if len(stages) > 0 && currentStageIdx+1 < len(stages) && completedShapes >= stages[currentStageIdx].limit {
		currentStageIdx++
		nextRes := stages[currentStageIdx].res
		nextTarget := resize.Thumbnail(uint(nextRes), uint(nextRes), rawInputImage, resize.Bilinear)
		currentModel.SetTarget(nextTarget)
	}

	completedShapes++
	currentModel.SetProgress(float64(completedShapes) / float64(totalTargetCount))

	stepStart := time.Now()
	npsCandidateCount := currentModel.Step(currentMode, currentAlpha, currentRepeat)
	stepDuration := time.Since(stepStart).Seconds()
	nps := float64(npsCandidateCount) / (stepDuration + 0.0001)

	// Extract the newly added shape SVG snippet
	idx := len(currentModel.Shapes) - 1
	shape := currentModel.Shapes[idx]
	c := currentModel.Colors[idx]
	scale := currentModel.Scale
	if idx < len(currentModel.Scales) {
		scale = currentModel.Scales[idx]
	}

	attrs := fmt.Sprintf("fill=\"#%02x%02x%02x\" fill-opacity=\"%f\"", c.R, c.G, c.B, float64(c.A)/255)
	shapeSvg := shape.SVG(attrs)
	svgChunk := fmt.Sprintf("<g transform=\"scale(%f) translate(0.5 0.5)\">%s</g>", scale, shapeSvg)

	elapsedSec := time.Since(startTime).Seconds()

	return map[string]any{
		"done":       false,
		"frame":      completedShapes,
		"total":      totalTargetCount,
		"score":      currentModel.Score,
		"elapsed":    elapsedSec,
		"nps":        nps,
		"svgChunk":   svgChunk,
		"stage":      currentStageIdx,
		"resolution": currentModel.Target.Bounds().Dx(),
	}
}

func getSvg(this js.Value, args []js.Value) any {
	if currentModel == nil {
		return ""
	}
	return currentModel.SVG()
}

func stopEngine(this js.Value, args []js.Value) any {
	isEngineRunning = false
	return true
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("primitiveInit", js.FuncOf(initEngine))
	js.Global().Set("primitiveStep", js.FuncOf(stepEngine))
	js.Global().Set("primitiveGetSvg", js.FuncOf(getSvg))
	js.Global().Set("primitiveStop", js.FuncOf(stopEngine))

	<-c
}
