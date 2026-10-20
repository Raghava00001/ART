package web

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"art/primitive"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

//go:embed static/*
var staticFS embed.FS

const (
	// MaxUploadBytes limits file size to 20MB, preventing memory and disk exhaustion DoS attacks
	MaxUploadBytes = 20 << 20
	// MaxImageDim limits maximum width or height to 8192px
	MaxImageDim = 8192
	// MaxPixelCount limits total resolution to ~33.5 megapixels, preventing decompression bombs (e.g. 50000x50000 px)
	MaxPixelCount = 33554432
)

var allowedMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// validateAndDecodeImage performs strict validation on file size, magic-byte MIME type,
// and dimensions (mitigating decompression/pixel bombs) before full raster decoding.
func validateAndDecodeImage(data []byte) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}
	if len(data) > MaxUploadBytes {
		return nil, fmt.Errorf("file size (%d bytes) exceeds maximum limit of 20MB", len(data))
	}

	// Sniff magic bytes (first 512 bytes) to verify genuine image content
	sniffLen := 512
	if len(data) < sniffLen {
		sniffLen = len(data)
	}
	mimeType := strings.TrimSpace(strings.Split(http.DetectContentType(data[:sniffLen]), ";")[0])
	if !allowedMIMEs[mimeType] {
		return nil, fmt.Errorf("unsupported or dangerous file type: %s. Only JPEG, PNG, GIF, and WebP are allowed", mimeType)
	}

	// Inspect dimensions BEFORE full pixel decompression to prevent decompression bombs
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("corrupt image header: %w", err)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", cfg.Width, cfg.Height)
	}

	if cfg.Width > MaxImageDim || cfg.Height > MaxImageDim {
		return nil, fmt.Errorf("image dimension (%dx%d) exceeds maximum allowed size (%dx%d px)", cfg.Width, cfg.Height, MaxImageDim, MaxImageDim)
	}

	if int64(cfg.Width)*int64(cfg.Height) > MaxPixelCount {
		return nil, fmt.Errorf("total pixel count (%d MP) exceeds maximum safety limit (33.5 MP)", (cfg.Width*cfg.Height)/1000000)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

type Server struct {
	mu          sync.Mutex
	activeModel *primitive.Model
	rawImage    image.Image
	running     bool
	cancelChan  chan struct{}
	clients     map[chan string]bool
	lastSVG     string
	port        int
}

func NewServer(port int) *Server {
	return &Server{
		clients: make(map[chan string]bool),
		port:    port,
	}
}

func (s *Server) broadcast(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

type StartRequest struct {
	Count      int    `json:"count"`
	Mode       int    `json:"mode"`
	Alpha      int    `json:"alpha"`
	Repeat     int    `json:"repeat"`
	Resolution int    `json:"resolution"`
	OutputSize int    `json:"outputSize"`
	Pyramid    bool   `json:"pyramid"`
	Continue   bool   `json:"continue,omitempty"`  // continue synthesis on existing model without restarting
	ImageData  string `json:"imageData,omitempty"` // optional base64 data uri
	MaskData   string `json:"maskData,omitempty"`  // optional base64 user priority mask
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Strictly limit request body to prevent network and memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadBytes)
	if err := r.ParseMultipartForm(MaxUploadBytes); err != nil {
		http.Error(w, "Upload exceeds maximum allowed size (20MB)", http.StatusRequestEntityTooLarge)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			r.MultipartForm.RemoveAll() // Clean up any temporary multipart files from disk immediately
		}
	}()

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, MaxUploadBytes+1))
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	if len(data) > MaxUploadBytes {
		http.Error(w, "Upload exceeds maximum allowed size (20MB)", http.StatusRequestEntityTooLarge)
		return
	}

	img, err := validateAndDecodeImage(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.rawImage = img
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"width":  img.Bounds().Dx(),
		"height": img.Bounds().Dy(),
		"status": "ok",
	})
}

func (s *Server) handleLoadDefault(w http.ResponseWriter, r *http.Request) {
	possibleDiskPaths := []string{
		"web/static/sample.jpg",
		"static/sample.jpg",
		"sample.jpg",
	}

	var data []byte
	for _, p := range possibleDiskPaths {
		if d, err := os.ReadFile(p); err == nil && len(d) > 0 {
			data = d
			break
		}
	}

	// Fallback to embedded filesystem
	if len(data) == 0 {
		if d, err := staticFS.ReadFile("static/sample.jpg"); err == nil && len(d) > 0 {
			data = d
		}
	}

	if len(data) == 0 {
		http.Error(w, "No sample image found", http.StatusNotFound)
		return
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		http.Error(w, "Failed to decode sample image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.rawImage = img
	s.mu.Unlock()

	b64 := base64.StdEncoding.EncodeToString(data)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"width":  img.Bounds().Dx(),
		"height": img.Bounds().Dy(),
		"b64":    "data:image/jpeg;base64," + b64,
	})
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Bound request body size to prevent JSON memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadBytes+5<<20)

	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload or size limit exceeded", http.StatusBadRequest)
		return
	}

	// Numerical parameter sanitization & bounds checking
	if req.Count <= 0 {
		req.Count = 100
	} else if req.Count > 5000 {
		req.Count = 5000
	}

	if req.Resolution <= 0 {
		req.Resolution = 384
	} else if req.Resolution < 32 {
		req.Resolution = 32
	} else if req.Resolution > 2048 {
		req.Resolution = 2048
	}

	if req.OutputSize <= 0 {
		req.OutputSize = 1024
	} else if req.OutputSize < 64 {
		req.OutputSize = 64
	} else if req.OutputSize > 8192 {
		req.OutputSize = 8192
	}

	if req.Mode < 0 || req.Mode > 9 {
		req.Mode = 1
	}

	if req.Alpha < 0 {
		req.Alpha = 0
	} else if req.Alpha > 255 {
		req.Alpha = 255
	}

	if req.Repeat < 0 {
		req.Repeat = 0
	} else if req.Repeat > 50 {
		req.Repeat = 50
	}

	s.mu.Lock()
	if req.ImageData != "" {
		// Limit base64 string length to avoid memory spikes
		if len(req.ImageData) <= (MaxUploadBytes * 4 / 3) {
			idx := strings.Index(req.ImageData, ",")
			raw := req.ImageData
			if idx >= 0 {
				raw = req.ImageData[idx+1:]
			}
			data, err := base64.StdEncoding.DecodeString(raw)
			if err == nil {
				img, err := validateAndDecodeImage(data)
				if err == nil {
					s.rawImage = img
				}
			}
		}
	}

	if s.rawImage == nil {
		s.mu.Unlock()
		http.Error(w, "No image uploaded yet", http.StatusBadRequest)
		return
	}

	// Stop any existing run
	if s.running && s.cancelChan != nil {
		close(s.cancelChan)
	}
	s.running = true
	s.cancelChan = make(chan struct{})
	cancel := s.cancelChan
	rawImg := s.rawImage
	s.mu.Unlock()

	go s.runWorkerTask(rawImg, req, cancel)

	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.running && s.cancelChan != nil {
		close(s.cancelChan)
		s.running = false
	}
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan string, 64)
	s.mu.Lock()
	s.clients[ch] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		close(ch)
		s.mu.Unlock()
	}()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	sizeParam := strings.ToLower(r.URL.Query().Get("size"))
	texture := strings.ToLower(r.URL.Query().Get("texture"))
	mat := r.URL.Query().Get("mat") == "true"

	s.mu.Lock()
	model := s.activeModel
	s.mu.Unlock()

	if model == nil {
		http.Error(w, "No completed model available", http.StatusBadRequest)
		return
	}

	if format == "svg" {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Disposition", "attachment; filename=\"art.svg\"")
		w.Write([]byte(model.SVG()))
		return
	}

	targetSize := 0
	switch sizeParam {
	case "2k":
		targetSize = 2048
	case "4k":
		targetSize = 3840
	case "5000", "print":
		targetSize = 5000
	case "8k":
		targetSize = 7680
	default:
		if sizeParam != "" {
			fmt.Sscanf(sizeParam, "%d", &targetSize)
			if targetSize > 8192 {
				targetSize = 8192
			} else if targetSize < 0 {
				targetSize = 0
			}
		}
	}

	var exportImg image.Image
	filename := "art.png"

	if targetSize > 0 {
		exportImg = model.RenderHighRes(targetSize, texture, mat)
		filename = fmt.Sprintf("art-%dpx-master.png", targetSize)
	} else {
		exportImg = model.Context.Image()
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	png.Encode(w, exportImg)
}

func (s *Server) runWorkerTask(rawInput image.Image, req StartRequest, cancel chan struct{}) {
	count := req.Count
	if count <= 0 {
		count = 100
	}
	inputSize := req.Resolution
	if inputSize <= 0 {
		inputSize = 384
	}
	outputSize := req.OutputSize
	if outputSize <= 0 {
		outputSize = 1024
	}
	mode := req.Mode
	alpha := req.Alpha
	if alpha <= 0 {
		alpha = 128
	}
	repeat := req.Repeat

	type pyramidStage struct {
		res   int
		limit int
	}
	var stages []pyramidStage
	if req.Pyramid && inputSize >= 64 && count >= 10 {
		if inputSize >= 256 {
			stages = []pyramidStage{
				{inputSize / 4, count * 20 / 100},
				{inputSize / 2, count * 50 / 100},
				{inputSize, count},
			}
		} else {
			stages = []pyramidStage{
				{inputSize / 2, count * 35 / 100},
				{inputSize, count},
			}
		}
	}

	currentStageIdx := 0

	s.mu.Lock()
	existingModel := s.activeModel
	s.mu.Unlock()

	var model *primitive.Model
	isContinued := req.Continue && existingModel != nil
	var bgHex string

	if isContinued {
		model = existingModel
		bgHex = fmt.Sprintf("#%02x%02x%02x", model.Background.R, model.Background.G, model.Background.B)
		fineTarget := resize.Thumbnail(uint(inputSize), uint(inputSize), rawInput, resize.Bilinear)
		model.SetTarget(fineTarget)
	} else {
		var initialInput image.Image
		if len(stages) > 0 {
			initialInput = resize.Thumbnail(uint(stages[0].res), uint(stages[0].res), rawInput, resize.Bilinear)
		} else {
			initialInput = resize.Thumbnail(uint(inputSize), uint(inputSize), rawInput, resize.Bilinear)
		}

		bg := primitive.MakeColor(primitive.AverageImageColor(initialInput))
		bgHex = fmt.Sprintf("#%02x%02x%02x", bg.R, bg.G, bg.B)
		workers := 4
		model = primitive.NewModel(initialInput, bg, outputSize, workers)
	}

	// Decode optional user focus brush mask
	if req.MaskData != "" {
		idx := strings.Index(req.MaskData, ",")
		rawMask := req.MaskData
		if idx >= 0 {
			rawMask = req.MaskData[idx+1:]
		}
		maskBytes, err := base64.StdEncoding.DecodeString(rawMask)
		if err == nil {
			mImg, _, err := image.Decode(bytes.NewReader(maskBytes))
			if err == nil {
				model.SetUserMask(mImg)
			}
		}
	} else if !isContinued {
		model.SetUserMask(nil)
	}

	s.mu.Lock()
	s.activeModel = model
	s.mu.Unlock()

	startShapes := len(model.Shapes)
	totalTarget := startShapes + count

	initEvent := map[string]interface{}{
		"type":      "init",
		"width":     model.Sw,
		"height":    model.Sh,
		"bg":        bgHex,
		"total":     totalTarget,
		"continued": isContinued,
		"start":     startShapes,
	}
	initData, _ := json.Marshal(initEvent)
	s.broadcast(string(initData))

	start := time.Now()
	for step := 1; step <= count; step++ {
		currentShapeNum := startShapes + step
		select {
		case <-cancel:
			s.broadcast(`{"type":"cancelled"}`)
			return
		default:
		}

		if !isContinued && len(stages) > 0 && currentStageIdx+1 < len(stages) && step > stages[currentStageIdx].limit {
			currentStageIdx++
			nextRes := stages[currentStageIdx].res
			nextTarget := resize.Thumbnail(uint(nextRes), uint(nextRes), rawInput, resize.Bilinear)
			model.SetTarget(nextTarget)
		}

		progress := float64(currentShapeNum) / float64(totalTarget)
		model.SetProgress(progress)

		t := time.Now()
		n := model.Step(primitive.ShapeType(mode), alpha, repeat)
		elapsed := time.Since(start).Seconds()
		nps := float64(n) / time.Since(t).Seconds()

		// Get the last added shape's SVG chunk
		lastIdx := len(model.Shapes) - 1
		lastShape := model.Shapes[lastIdx]
		lastColor := model.Colors[lastIdx]
		lastScale := model.Scales[lastIdx]
		attrs := fmt.Sprintf("fill=\"#%02x%02x%02x\" fill-opacity=\"%f\"",
			lastColor.R, lastColor.G, lastColor.B, float64(lastColor.A)/255)
		shapeSVG := fmt.Sprintf("<g transform=\"scale(%f) translate(0.5 0.5)\">%s</g>",
			lastScale, lastShape.SVG(attrs))

		event := map[string]interface{}{
			"type":       "shape",
			"frame":      currentShapeNum,
			"total":      totalTarget,
			"score":      model.Score,
			"elapsed":    elapsed,
			"nps":        nps,
			"svgChunk":   shapeSVG,
			"stage":      currentStageIdx,
			"resolution": model.Target.Bounds().Dx(),
		}
		data, _ := json.Marshal(event)
		s.broadcast(string(data))
	}

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	finalEvent := map[string]interface{}{
		"type":      "done",
		"total":     len(model.Shapes),
		"score":     model.Score,
		"duration":  time.Since(start).Seconds(),
		"fullSVG":   model.SVG(),
		"continued": isContinued,
	}
	data, _ := json.Marshal(finalEvent)
	s.broadcast(string(data))
}

// securityHeadersMiddleware injects defense-in-depth HTTP headers against MIME sniffing,
// clickjacking, unauthorized embedding, and malicious external script inclusion.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; connect-src 'self'; script-src 'self' 'unsafe-inline';")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func StartServer(port int) error {
	s := NewServer(port)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/upload", s.handleUpload)
	mux.HandleFunc("/api/sample", s.handleLoadDefault)
	mux.HandleFunc("/api/start", s.handleStart)
	mux.HandleFunc("/api/stop", s.handleStop)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/export", s.handleExport)

	// Determine static files: prefer local disk if present (for live dev edits),
	// otherwise fall back seamlessly to embedded assets.
	var fileHandler http.Handler
	dir, err := os.Getwd()
	if err == nil {
		localStaticDir := filepath.Join(dir, "web", "static")
		if _, err := os.Stat(filepath.Join(localStaticDir, "index.html")); err == nil {
			fileHandler = http.FileServer(http.Dir(localStaticDir))
		}
	}

	if fileHandler == nil {
		subFS, err := fs.Sub(staticFS, "static")
		if err != nil {
			return fmt.Errorf("failed to mount embedded static files: %w", err)
		}
		fileHandler = http.FileServer(http.FS(subFS))
	}

	mux.Handle("/", fileHandler)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("ART Studio listening at http://localhost:%d\n", port)
	return http.ListenAndServe(addr, securityHeadersMiddleware(mux))
}
