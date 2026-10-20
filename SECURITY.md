# Security Policy

## Overview

**ART** is designed to process user-uploaded images safely. Because users can submit arbitrary image files via the Web Studio GUI and CLI, the engine implements defense-in-depth security controls to protect against Denial of Service (DoS), decompression bombs, malicious file execution, and cross-site scripting (XSS).

---

## Threat Model & Security Controls

### 1. In-Memory Processing & Zero Persistence
- Uploaded images are never written to executable directories or stored permanently on disk.
- All image data is decoded directly into ephemeral Go `image.Image` memory structures.
- Any temporary multipart disk buffers generated during upload transit are purged immediately via `r.MultipartForm.RemoveAll()`.

### 2. Decompression Bomb & Pixel Bomb Mitigation
- Image dimensions (width and height) are inspected in the header stream using `image.DecodeConfig` **before** full bitmap memory allocation.
- Hard safety limits are enforced:
  - **Maximum Dimensions**: 8192 × 8192 pixels.
  - **Maximum Total Pixels**: 33,554,432 pixels (~33.5 megapixels).
- Payloads exceeding these bounds are rejected immediately with `400 Bad Request`, preventing memory exhaustion or process OOM crashes.

### 3. Magic-Byte MIME Sniffing & Type Verification
- Rather than trusting user-controlled file extensions (`.png`, `.jpg`), the engine inspects the initial 512 bytes using `http.DetectContentType`.
- Only authentic raster formats (`image/jpeg`, `image/png`, `image/gif`, `image/webp`) are accepted.
- Shell scripts, executable binaries (ELF, PE, Mach-O), HTML files, and SVG files containing `<script>` tags are blocked.

### 4. Strict Request Quotas (DoS Protection)
- HTTP upload bodies are restricted to **20 MB** via `http.MaxBytesReader`.
- Numerical parameters for synthesis iterations, resolution, output dimensions, and opacity are validated and clamped against strict safety ranges:
  - Primitive Count: `1` to `5,000`
  - Simulation Resolution: `32` to `2,048` px
  - Master Export Resolution: up to `8,192` px

### 5. HTTP Security Headers
All web studio endpoints are protected with defense-in-depth HTTP response headers:
- `X-Content-Type-Options: nosniff`: Prevents MIME-confusion attacks.
- `X-Frame-Options: SAMEORIGIN`: Protects against clickjacking.
- `Referrer-Policy: strict-origin-when-cross-origin`: Minimizes referrer leakage.
- `Content-Security-Policy`: Restricts resource loading to trusted self-origins and designated font CDNs.

---

## Supported Versions

| Version | Supported |
|:---|:---|
| 2.x (Current) | :white_check_mark: |
| < 2.0 | :x: |

---

## Reporting a Vulnerability

If you discover a security vulnerability within this repository:

1. **Do not create a public issue.**
2. Report the vulnerability privately via GitHub Security Advisories or by contacting the maintainers directly.
3. Please include:
   - A description of the vulnerability.
   - Steps to reproduce or a proof of concept (PoC).
   - Potential impact.

We take security issues seriously and will respond promptly to validate and coordinate a fix.
