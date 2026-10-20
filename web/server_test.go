package web

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createSmallTestPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{100, 150, 200, 255}}, image.ZP, draw.Src)
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestEmbeddedStaticAssets(t *testing.T) {
	// Verify that embedded filesystem contains the required web assets
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		t.Fatalf("Failed to mount embedded static filesystem: %v", err)
	}

	requiredFiles := []string{"index.html", "style.css", "app.js", "sample.jpg"}
	for _, fname := range requiredFiles {
		f, err := sub.Open(fname)
		if err != nil {
			t.Errorf("Embedded static filesystem missing required file %q: %v", fname, err)
			continue
		}
		stat, err := f.Stat()
		if err != nil || stat.Size() == 0 {
			t.Errorf("Embedded file %q is empty or failed to stat", fname)
		}
		f.Close()
	}
}

func TestSampleEndpoint(t *testing.T) {
	server := NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/api/sample", nil)
	w := httptest.NewRecorder()

	server.handleLoadDefault(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK from /api/sample, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to parse JSON from /api/sample: %v", err)
	}

	if data["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", data["status"])
	}
	b64, ok := data["b64"].(string)
	if !ok || !strings.HasPrefix(b64, "data:image/jpeg;base64,") {
		t.Errorf("Expected data URI in b64 field, got %v", b64)
	}
	width, _ := data["width"].(float64)
	height, _ := data["height"].(float64)
	if width <= 0 || height <= 0 {
		t.Errorf("Expected positive dimensions, got %vx%v", width, height)
	}
}

func TestStaticServerRouting(t *testing.T) {
	server := NewServer(8080)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sample", server.handleLoadDefault)

	var fileHandler http.Handler
	dir, err := os.Getwd()
	if err == nil {
		localStaticDir := filepath.Join(dir, "web", "static")
		if _, err := os.Stat(filepath.Join(localStaticDir, "index.html")); err == nil {
			fileHandler = http.FileServer(http.Dir(localStaticDir))
		}
	}
	if fileHandler == nil {
		subFS, _ := fs.Sub(staticFS, "static")
		fileHandler = http.FileServer(http.FS(subFS))
	}
	mux.Handle("/", fileHandler)

	// Test GET /
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for GET /, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ART Studio") {
		t.Errorf("Expected 'ART Studio' in HTML response body")
	}
}

func TestMethodValidation(t *testing.T) {
	server := NewServer(8080)

	// handleUpload should reject GET
	reqUpload := httptest.NewRequest(http.MethodGet, "/api/upload", nil)
	wUpload := httptest.NewRecorder()
	server.handleUpload(wUpload, reqUpload)
	if wUpload.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed for GET /api/upload, got %d", wUpload.Result().StatusCode)
	}

	// handleStart should reject GET
	reqStart := httptest.NewRequest(http.MethodGet, "/api/start", nil)
	wStart := httptest.NewRecorder()
	server.handleStart(wStart, reqStart)
	if wStart.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed for GET /api/start, got %d", wStart.Result().StatusCode)
	}
}

func TestUploadSecurity_ValidImage(t *testing.T) {
	server := NewServer(8080)

	pngBytes := createSmallTestPNG(32, 32)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "photo.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write(pngBytes)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleUpload(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for valid PNG upload, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	if data["status"] != "ok" || data["width"].(float64) != 32 {
		t.Errorf("Unexpected upload response: %v", data)
	}
}

func TestUploadSecurity_RejectsNonImage(t *testing.T) {
	server := NewServer(8080)

	// Attempt to upload malicious script disguised as an image
	maliciousPayload := []byte("#!/bin/bash\necho 'malicious command execution attempt'\n")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("image", "exploit.sh")
	part.Write(maliciousPayload)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleUpload(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for non-image script upload, got %d", resp.StatusCode)
	}
}

func TestUploadSecurity_RejectsSVGScript(t *testing.T) {
	server := NewServer(8080)

	// Attempt to upload SVG containing script tag
	svgPayload := []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"><script>alert('xss')</script></svg>")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("image", "vector_xss.svg")
	part.Write(svgPayload)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleUpload(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for SVG script upload, got %d", resp.StatusCode)
	}
}

func TestSecurityHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	handler := securityHeadersMiddleware(mux)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Missing X-Content-Type-Options: nosniff header")
	}
	if resp.Header.Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("Missing X-Frame-Options: SAMEORIGIN header")
	}
	if !strings.Contains(resp.Header.Get("Content-Security-Policy"), "default-src 'self'") {
		t.Errorf("Missing or invalid Content-Security-Policy header")
	}
}

func TestValidateAndDecodeImage_DecompressionBomb(t *testing.T) {
	// Synthesize fake PNG header with huge dimensions (65536 x 65536 px)
	header := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR length
		0x49, 0x48, 0x44, 0x52, // "IHDR"
		0x00, 0x01, 0x00, 0x00, // Width = 65536 px (> 8192)
		0x00, 0x01, 0x00, 0x00, // Height = 65536 px (> 8192)
		0x08, 0x02, 0x00, 0x00, 0x00,
	}
	_, err := validateAndDecodeImage(header)
	if err == nil {
		t.Fatalf("Expected validateAndDecodeImage to reject oversized PNG dimensions")
	}
}
