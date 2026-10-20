# ART Engine — Technology Stack & Architecture

This document provides a comprehensive technical breakdown of the technologies used in **ART**, detailing **what** is used, **why** it was selected over alternatives, and **how** it is implemented within the system architecture.

---

## Architecture Summary

```text
  ┌─────────────────────────────────────────────────────────────────┐
  │                        Client Layer                             │
  │   - HTML5 Glassmorphic UI        - Server-Sent Events (SSE)     │
  │   - SVG Vector Viewport          - Canvas Priority Masking      │
  │   - Vanilla JS Motion Engine     - CSS Tokens (Dark/Light)      │
  └───────────────────────────────▲─────────────────────────────────┘
                                  │ Real-Time HTTP/SSE Streams
  ┌───────────────────────────────▼─────────────────────────────────┐
  │                     Go Web Server (web/)                        │
  │   - //go:embed Static Assets     - MaxBytesReader Quotas        │
  │   - Decompression Bomb Defense   - MIME Magic-Byte Sniffer      │
  │   - Event Broadcast Hub          - Security Headers Middleware  │
  └───────────────────────────────▲─────────────────────────────────┘
                                  │ Memory-Only Buffer Hand-Off
  ┌───────────────────────────────▼─────────────────────────────────┐
  │                 Primitive Core Engine (primitive/)              │
  │   - Residual Importance Map (CDF)- Edge Saliency Guidance       │
  │   - Multi-Resolution Pyramid     - Parallel Hill-Climbing Pool  │
  │   - Scanline Rasterizer          - 10 Geometric Shape Models    │
  │   - FreeType Fixed-Point Math    - SVG Generator / High-Res PNG │
  └─────────────────────────────────────────────────────────────────┘
```

---

## Technology Stack Deep Dive

### 1. Core Systems Language: Go (Golang 1.22+)

| Component | Usage in Project |
|:---|:---|
| **Why Go?** | - **Raw Concurrency**: Lightweight goroutines allow concurrent exploration of hundreds of thousands of stochastic shape mutations per second across all CPU cores without thread overhead.<br>- **Zero-Dependency Native Binaries**: Compiles into a single self-contained executable with no runtime (no Python interpreter, no Node.js runtime, no DLL dependencies).<br>- **Memory Safety**: High-speed memory management without the segmentation fault and buffer overflow risks of C/C++.<br>- **Built-in `embed.FS`**: Bundles the entire web GUI, styling, client scripts, and sample images directly into the executable binary. |
| **How It Is Used** | - Worker pools in [`primitive/worker.go`](primitive/worker.go) run hill-climbing search concurrently across `runtime.NumCPU()` goroutines.<br>- State synchronization is handled via mutex locks (`sync.Mutex`) and cancellation channels (`chan struct{}`).<br>- The HTTP web atelier in [`web/server.go`](web/server.go) runs directly on Go's standard `net/http` server. |

---

### 2. High-Performance Graphics & Rasterization

#### A. FreeType Fixed-Point Rasterizer (`github.com/golang/freetype/raster`)
- **Why**: Mathematical curves (quadratic and cubic Béziers) require antialiased vector rasterization down to sub-pixel coordinates. FreeType uses 26.6 fixed-point arithmetic (`x * 64`), providing blazing rendering speed without floating-point inaccuracies.
- **How**: Used in [`primitive/raster.go`](primitive/raster.go), [`primitive/quadratic.go`](primitive/quadratic.go), and [`primitive/cubic.go`](primitive/cubic.go). Paths are converted into horizontal scanline spans (`[]Scanline`), enabling difference evaluation against the target bitmap in microseconds.

#### B. 2D Vector Context (`github.com/fogleman/gg`)
- **Why**: Provides an idiomatic Cairo-like 2D drawing API for master rendering, supporting transformation matrices, anti-aliasing, linear gradients, and canvas clipping.
- **How**: Used during master high-resolution raster export (2K, 4K, 8K) in [`primitive/model.go`](primitive/model.go) and texturing effects.

#### C. High-Fidelity Downsampling (`github.com/nfnt/resize`)
- **Why**: High-quality Bilinear and Lanczos resampling preserves edge gradients during multi-stage resolution shifts.
- **How**: Powers the **Coarse-to-Fine Multi-Resolution Pyramid** in [`primitive/model.go`](primitive/model.go) and [`web/server.go`](web/server.go). The target is progressively upscaled from 64px (macro blocking) to 128px (mid-range forms) to 384px (fine facial features).

---

### 3. Web Atelier & Client Communication

#### A. Server-Sent Events (SSE) via HTTP/1.1 & HTTP/2
- **Why SSE over WebSockets?**
  - WebSockets require full-duplex protocol handshakes and custom framing.
  - ART synthesis is fundamentally **unidirectional streaming** (the server pushes generated shapes to the client at 30–60 fps).
  - SSE runs over standard HTTP, natively supports automatic connection recovery, works seamlessly behind corporate proxies and CDNs, and avoids WebSocket upgrade failures.
- **How**: In [`web/server.go`](web/server.go), `handleEvents` flushes JSON event payloads over `text/event-stream`. The client's native `EventSource` in [`web/static/app.js`](web/static/app.js) listens and appends vector chunks directly to the live SVG DOM without re-rendering the whole scene.

#### B. Pure Vanilla ES6+ JavaScript
- **Why Vanilla JS over React/Vue/Next.js?**
  - **Zero Build Step**: No `node_modules`, no Webpack/Vite bundle step, and zero npm vulnerabilities.
  - **Instant Startup**: The web page boots in < 50 milliseconds.
  - **Embedded Portability**: The single Go binary can serve the complete GUI without requiring Node.js on the host machine.
- **How**:
  - Manipulates the live SVG DOM tree directly for 60fps rendering.
  - Implements the interactive A/B split wipe comparison slider.
  - Extracts dynamic color palettes from reference photos using HTML5 Canvas `getImageData`.
  - Powers user-painted focal priority masking.

#### C. Glassmorphic Modern CSS Design System
- **Why**: Provides a responsive, aesthetically striking creative environment inspired by modern studio software (Figma, Linear, Apple macOS).
- **How**: [`web/static/style.css`](web/static/style.css) utilizes:
  - CSS Custom Properties (`--bg-primary`, `--accent-primary`, etc.) for seamless, persistent Dark/Light theme switching.
  - Hardware-accelerated `backdrop-filter: blur(24px)` for glass panels.
  - CSS Grid and Flexbox for responsive scaling on desktop and mobile.

---

### 4. Algorithmic Innovations

#### A. Residual-Weighted Importance Sampling (CDF Error Maps)
- **Problem**: Classic primitive generators pick random positions across the entire image uniformly. In portraits or landscapes, 70% of CPU time is wasted testing shapes on flat, uniform backgrounds where error is already minimal.
- **Solution in ART**:
  1. Computes the absolute RGB difference between the current canvas and the target: $\Delta = |T - C|$.
  2. Builds a 2D cumulative distribution function (CDF) heatmap across image cells in [`primitive/errormap.go`](primitive/errormap.go).
  3. Uses binary search over the CDF to spawn candidate shapes preferentially in high-error regions (eyes, lips, sharp contours).

#### B. Saliency-Guided Adaptive Feature Scaling
- **Problem**: Small shapes placed on large flat regions cause noisy artifacts; large shapes placed on delicate features blur out details.
- **Solution in ART**:
  1. [`primitive/saliency.go`](primitive/saliency.go) computes Sobel edge gradients across the image.
  2. Anneals shape radius dynamically based on local edge saliency: large strokes cover backgrounds, while delicate micro-strokes concentrate on eyes, hair strands, and edges.

#### C. Advanced Bézier & Cubic Spline Geometry
- **Mode 6 (Variable Béziers)**: Solves quadratic Bézier curves with mutable stroke widths (0.5–16px) for calligraphic brush strokes.
- **Mode 9 (Cubic Béziers)**: Implements 4-control-point cubic curves ($P_0, P_1, P_2, P_3$), enabling inflection points (S-curves) essential for flowing hair, drapery, and organic contours.

---

### 5. Security & Safety Architecture

| Threat | Solution Implemented |
|:---|:---|
| **Decompression / Pixel Bombs** | Pre-decompression header inspection via `image.DecodeConfig` rejects dimensions > 8192px or > 33.5 megapixels before bitmap allocation. |
| **Network & Memory DoS** | Request body size capped at 20MB using `http.MaxBytesReader`. |
| **MIME Spoofing** | Magic-byte inspection via `http.DetectContentType` rejects non-image files, scripts, and executable payloads. |
| **Disk Exhaustion** | Immediate cleanup with `defer r.MultipartForm.RemoveAll()`; zero permanent files written to disk. |
| **XSS & Clickjacking** | HTTP defense headers: `X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`, and strict `Content-Security-Policy`. |

---

## Project Structure

```text
ART/
├── .github/
│   └── workflows/
│       └── ci.yml               # Multi-platform CI (Ubuntu, macOS, Windows)
├── assets/                      # Sample visual output renders
│   ├── monalisa_triangles.png
│   ├── owl_beziers.png
│   └── pyramids_cubics.png
├── examples/                    # Sample reference input images
├── primitive/                   # Core geometric synthesis engine
│   ├── core.go                  # Color computation and line blitting
│   ├── cubic.go                 # Mode 9: Cubic Bézier 4-point splines
│   ├── quadratic.go             # Mode 6: Variable-width quadratic curves
│   ├── triangle.go              # Mode 1: Triangle rasterizer
│   ├── rectangle.go             # Modes 2, 5: Axis & rotated rectangles
│   ├── ellipse.go               # Modes 3, 4, 7: Ellipses and circles
│   ├── polygon.go               # Mode 8: Convex 4-point polygons
│   ├── errormap.go              # Residual error CDF importance sampling
│   ├── saliency.go              # Sobel edge gradient map
│   ├── raster.go                # Fixed-point FreeType scanline converter
│   ├── model.go                 # Multi-resolution pyramid and state manager
│   ├── worker.go                # Concurrent hill-climbing search workers
│   └── primitive_test.go        # Unit tests for rasterizers and models
├── web/                         # Interactive browser atelier & HTTP server
│   ├── server.go                # //go:embed server, SSE streaming, security
│   ├── server_test.go           # Server endpoint & security test suite
│   └── static/                  # Embedded client web assets
│       ├── index.html           # Glassmorphic UI markup & SVG viewport
│       ├── style.css            # CSS custom properties design system
│       ├── app.js               # EventSource consumer & canvas motion logic
│       └── sample.jpg           # Embedded classical reference portrait
├── Dockerfile                   # Production container definition for live cloud deploy
├── main.go                      # CLI flag parser and GUI launcher
├── main_test.go                 # Root package test suite
├── go.mod                       # Go module dependencies
├── LICENSE.md                   # MIT License
├── README.md                    # Comprehensive user guide & documentation
├── SECURITY.md                  # Security policies and threat model
└── TECH_STACK.md                # Technology stack & architectural rationale
```
