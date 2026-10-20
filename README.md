# ART — Autonomous Geometric Vector Art Engine

[![Go Report Card](https://goreportcard.com/badge/github.com/art/primitive)](https://goreportcard.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Security: Hardened](https://img.shields.io/badge/Security-Hardened-success.svg)](SECURITY.md)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](go.mod)
[![Architecture](https://img.shields.io/badge/Binary-Zero--Dependency%20Embedded-brightgreen.svg)](#installation--build)

**ART** is a high-performance vector synthesis engine that reproduces raster images using intelligent geometric and calligraphic primitives. Engineered with residual error heatmaps, multi-resolution annealing pyramids, and an interactive glassmorphic web atelier.

---

## Key Features

- **Residual-Weighted Importance Sampling**: Shapes spawn preferentially in high-error regions (facial contours, fine edges, subtle textures) using CDF-sampled residual error heatmaps rather than wasting compute cycles on uniform backgrounds.
- **Enhanced Bézier & Spline Engine**:
  - **Mode 6 (Variable Béziers)**: Quadratic curves with dynamic stroke widths (0.5–16px) for expressive, calligraphic brushwork.
  - **Mode 9 (Cubic Béziers)**: 4-control-point cubic curves enabling smooth S-curves, organic silhouettes, and fine strands.
- **Coarse-to-Fine Multi-Resolution Pyramid**: Accelerates macro convergence at low resolution (64px) before progressively refining midrange (128px) and high-frequency details (256–512px).
- **ART Studio (Interactive Web Atelier)**: A glassmorphic dark-mode web studio featuring real-time Server-Sent Events (SSE) streaming, interactive wipe comparison, manual priority brushing, and instant SVG/PNG/4K master exports.
- **Zero-Dependency Single Binary**: Using Go's `embed.FS`, all web assets, stylesheets, scripts, and sample artwork are compiled directly into a single portable binary.
- **Hardened Upload Security**: Comprehensive defense-in-depth security including decompression bomb mitigation, magic-byte MIME sniffing, strict 20MB request caps, parameter clamping, and zero disk persistence. Check out [SECURITY.md](SECURITY.md) for details.

---

## Architecture Overview

```text
  ┌─────────────────────────────────────────────────────────┐
  │                    Input Reference                      │
  └───────────────────────────┬─────────────────────────────┘
                              │
               ┌──────────────▼──────────────┐
               │ Multi-Resolution Pyramid    │
               │ (64px ➔ 128px ➔ 256/384px)  │
               └──────────────┬──────────────┘
                              │
       ┌──────────────────────▼──────────────────────┐
       │ Residual-Weighted Importance Map (CDF)       │
       │ + Edge Saliency Guidance                    │
       └──────────────────────┬──────────────────────┘
                              │
               ┌──────────────▼──────────────┐
               │ Concurrent Workers (Hill-   │
               │ Climbing Stochastic Search) │
               └──────────────┬──────────────┘
                              │
       ┌──────────────────────▼──────────────────────┐
       │ Primitives (Triangles, Beziers, Cubics...)  │
       └──────────────────────┬──────────────────────┘
                              │
       ┌──────────────────────▼──────────────────────┐
       │ Outputs: SVG Vector / PNG Raster / GIF /    │
       │ Real-Time SSE Web Studio Stream             │
       └─────────────────────────────────────────────┘
```

---

## Installation & Build

Build the single self-contained `art` binary:

```bash
go build -v -o art.exe .
```

*Note: On macOS and Linux, output binary is `./art`.*

To run the complete test suite:

```bash
go test -v ./...
```

---

## Usage

### 1. Interactive Web Studio

Launch the browser-based studio on any port (default `8080`):

```bash
art -gui -port 8080
```

Open `http://localhost:8080` in your browser:
- **Drag & Drop** any reference image or click the sample portrait to load instantly.
- **Select Primitives**: Choose from 10 geometric and calligraphic modes.
- **Priority Brush**: Paint over focal areas (eyes, mouth, signature lines) to concentrate detail.
- **A/B Split Wipe**: Slide the interactive divider to inspect raster vs. vector accuracy.
- **Export**: Download crisp scalable SVG vectors, transparent PNGs, or upscaled 4K master prints.

#### Web Studio Keyboard Shortcuts

| Shortcut | Action |
|:---|:---|
| <kbd>Enter</kbd> | Start / Stop synthesis |
| <kbd>Shift</kbd> + <kbd>Enter</kbd> | Continue / Refine current artwork |
| <kbd>Space</kbd> | Toggle Split Wipe / Vector Only view |
| <kbd>B</kbd> | Toggle Focal Priority Brush |
| <kbd>W</kbd> | Toggle Wireframe outline mode |
| <kbd>H</kbd> | Toggle Heatmap error visualization |
| <kbd>C</kbd> | Copy SVG markup to clipboard |
| <kbd>R</kbd> | Reload default reference artwork |
| <kbd>+</kbd> / <kbd>-</kbd> / <kbd>0</kbd> | Zoom In / Out / Fit to Canvas |
| <kbd>1</kbd>–<kbd>9</kbd>, <kbd>0</kbd> | Quick-switch primitive modes |
| <kbd>?</kbd> | Open Shortcuts modal dialog |

---

### 2. Command Line Interface

```bash
art -i input.jpg -o output.png -n 150 -m 1 -r 256 -pyramid
```

To generate both a vector SVG and a high-resolution PNG simultaneously:

```bash
art -i input.jpg -o output.svg -o output.png -n 200 -m 9 -s 2048 -pyramid
```

#### CLI Options Reference

| Flag | Default | Description |
|:---|:---|:---|
| `-i` | *(required)* | Path to input image |
| `-o` | *(required)* | Path to output image (`.png`, `.jpg`, `.svg`, `.gif`, can be repeated) |
| `-n` | *(required)* | Total number of primitives to generate |
| `-m` | `1` | Primitive mode (see table below) |
| `-pyramid` | `false` | Enable coarse-to-fine multi-resolution pyramid |
| `-gui` | `false` | Launch interactive web GUI studio |
| `-port` | `8080` | Web GUI server port |
| `-r` | `256` | Simulation resolution (higher values capture finer detail) |
| `-s` | `1024` | Output image render size |
| `-a` | `128` | Alpha opacity (0–255; `0` enables automatic per-shape alpha) |
| `-rep` | `0` | Extra shape refinement steps per iteration |
| `-nth` | `1` | Save every Nth frame when `%d` is in the output path |
| `-bg` | `""` | Background color hex (defaults to average image color) |
| `-j` | `0` | Number of parallel worker threads (default uses all CPU cores) |
| `-v` | `false` | Verbose progress logging |
| `-vv` | `false` | Extra verbose debugging output |

---

## Primitive Modes

| Mode | Shape Type | Description |
|:---:|:---|:---|
| **0** | Combo All | Dynamically mixes all available shapes |
| **1** | Triangles | Sharp geometric facets; ideal for low-poly portraits & architecture |
| **2** | Rectangles | Axis-aligned bounding boxes |
| **3** | Ellipses | Smooth elliptical splats |
| **4** | Circles | Circular pointillism / bokeh style |
| **5** | Rotated Rectangles | Oriented rectangular strokes |
| **6** | Variable Béziers | Quadratic curves with mutable 0.5–16px width for brushwork |
| **7** | Rotated Ellipses | Tilted ellipses capturing organic curvature |
| **8** | Polygons | Convex 4-sided polygonal shards |
| **9** | Cubic Béziers | 4-point cubic spline curves with calligraphic width for hair and contours |

---

## Output Formats

- **Vector SVG (`.svg`)**: Clean, scalable, lossless vector artwork with grouped scale layers.
- **Raster Image (`.png` / `.jpg`)**: Crisp high-resolution raster output up to 7680px (8K).
- **Frame Sequence (`frame_%04d.png`)**: Save progressive evolution frame by frame.
- **Animated GIF (`.gif`)**: Generate animated timelapses illustrating generation from scratch.

---

## Contributing

Contributions, bug reports, and suggestions are welcome!
1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/my-feature`).
3. Ensure all tests pass: `go test -v ./...`.
4. Commit your changes (`git commit -am 'Add new feature'`).
5. Push to the branch (`git push origin feature/my-feature`).
6. Open a Pull Request.

---

## License

This project is licensed under the [MIT License](LICENSE.md).  
Original core concept inspired by [Michael Fogleman](https://github.com/fogleman/primitive).
