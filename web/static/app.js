// ART Studio — Autonomous Geometric Vector Art Engine
// Antigravity Design Language & Motion Controller

document.addEventListener('DOMContentLoaded', () => {
  // ========================================================================
  // Ambient Interactive Particle Simulation (Antigravity Liftoff Atmosphere)
  // ========================================================================
  initAmbientParticles();

  // DOM References
  const dropZone = document.getElementById('drop-zone');
  const fileInput = document.getElementById('file-input');
  const dropContent = document.getElementById('drop-content');
  const sourcePreview = document.getElementById('source-preview');
  const btnLoadSample = document.getElementById('btn-load-sample');
  const sourceResBadge = document.getElementById('source-res-badge');
  const referenceCard = document.getElementById('reference-card');
  const referenceDims = document.getElementById('reference-dims');
  const referenceAspect = document.getElementById('reference-aspect');
  const paletteSwatches = document.getElementById('palette-swatches');
  const btnChangeImage = document.getElementById('btn-change-image');

  const presetPills = document.querySelectorAll('.preset-pill');
  const modeChips = document.querySelectorAll('.mode-chip');
  const sliders = document.querySelectorAll('.studio-slider');
  const inputCount = document.getElementById('input-count');
  const valCount = document.getElementById('val-count');
  const inputAlpha = document.getElementById('input-alpha');
  const valAlpha = document.getElementById('val-alpha');
  const inputRepeat = document.getElementById('input-repeat');
  const valRepeat = document.getElementById('val-repeat');
  const inputRes = document.getElementById('input-res');
  const valRes = document.getElementById('val-res');
  const togglePyramid = document.getElementById('toggle-pyramid');

  const headerModeDisplay = document.getElementById('header-mode-display');
  const headerResDisplay = document.getElementById('header-res-display');
  const headerPyramidDisplay = document.getElementById('header-pyramid-display');
  const engineStatusPill = document.getElementById('engine-status-pill');
  const engineStatusText = document.getElementById('engine-status-text');

  const btnStart = document.getElementById('btn-start');
  const btnStop = document.getElementById('btn-stop');
  const btnRefine = document.getElementById('btn-refine');
  const btnRefineText = document.getElementById('btn-refine-text');
  const btnCopySvg = document.getElementById('btn-copy-svg');
  const copyBtnText = document.getElementById('copy-btn-text');
  const btnExportPng = document.getElementById('btn-export-png');
  const btnExportSvg = document.getElementById('btn-export-svg');
  const btnExportAnim = document.getElementById('btn-export-anim');
  const btnExport4k = document.getElementById('btn-export-4k');
  const exportModal = document.getElementById('export-modal');
  const btnCloseExportModal = document.getElementById('btn-close-export-modal');
  const exportResCards = document.querySelectorAll('#export-res-options .export-option-card');
  const exportTexturePills = document.querySelectorAll('#export-texture-options .export-pill');
  const exportGalleryMat = document.getElementById('export-gallery-mat');
  const exportSpecDims = document.getElementById('export-spec-dims');
  const exportSpecPhysical = document.getElementById('export-spec-physical');
  const btnDoExportMaster = document.getElementById('btn-do-export-master');
  const exportSubmitText = document.getElementById('export-submit-text');

  const btnToggleBrush = document.getElementById('btn-toggle-brush');
  const btnToggleWireframe = document.getElementById('btn-toggle-wireframe');
  const btnToggleTexture = document.getElementById('btn-toggle-texture');
  const brushToolbar = document.getElementById('brush-toolbar');
  const brushSize = document.getElementById('brush-size');
  const brushSizeVal = document.getElementById('brush-size-val');
  const btnClearMask = document.getElementById('btn-clear-mask');
  const maskCanvas = document.getElementById('mask-canvas');
  const heatmapCanvas = document.getElementById('heatmap-canvas');
  const timelineScrubberWrap = document.getElementById('timeline-scrubber-wrap');
  const btnTimelinePlay = document.getElementById('btn-timeline-play');
  const timelineSlider = document.getElementById('timeline-slider');
  const timelineStepText = document.getElementById('timeline-step-text');

  const canvasStage = document.getElementById('canvas-stage');
  const canvasWrapper = document.getElementById('canvas-wrapper');
  const vectorCanvas = document.getElementById('vector-canvas');
  const wipeOverlay = document.getElementById('wipe-overlay');
  const splitImg = document.getElementById('split-img');
  const wipeHandle = document.getElementById('wipe-handle');
  const emptyState = document.getElementById('empty-state');
  const tagRef = document.getElementById('tag-ref');
  const tagVec = document.getElementById('tag-vec');
  const viewModeButtons = document.querySelectorAll('.segment-btn');

  const btnZoomIn = document.getElementById('btn-zoom-in');
  const btnZoomOut = document.getElementById('btn-zoom-out');
  const btnZoomFit = document.getElementById('btn-zoom-fit');
  const zoomLevelText = document.getElementById('zoom-level-text');
  const btnToggleGrid = document.getElementById('btn-toggle-grid');

  const progressBarFill = document.getElementById('progress-bar-fill');
  const metricShapes = document.getElementById('metric-shapes');
  const metricFidelity = document.getElementById('metric-fidelity');
  const metricScore = document.getElementById('metric-score');
  const metricSpeed = document.getElementById('metric-speed');
  const metricTime = document.getElementById('metric-time');
  const statusHint = document.getElementById('status-hint');

  const shortcutsModal = document.getElementById('shortcuts-modal');
  const btnShortcutsModal = document.getElementById('btn-shortcuts-modal');
  const btnCloseModal = document.getElementById('btn-close-modal');
  const toastShelf = document.getElementById('toast-shelf');

  // Application State
  let currentMode = 1;
  let currentImageData = null;
  let eventSource = null;
  let isRunning = false;
  let shapeCountTarget = 220;
  let currentDimensions = { width: 1024, height: 1024, bg: '#090a0d' };
  let wipeRatio = 0.5;
  let isDraggingWipe = false;
  let activeViewMode = 'split';
  let currentZoom = 1.0;
  let latestFullSvg = null;
  let currentTextureMode = 'none'; // 'none' | 'paper' | 'canvas'
  let isContinuing = false;
  let heatmapPending = false;

  // Mask & Detail Priority State
  let isBrushActive = false;
  let isDrawingMask = false;
  let hasUserMask = false;
  let currentBrushRadius = 35;
  let lastBrushX = 0;
  let lastBrushY = 0;
  let maskCtx = maskCanvas ? maskCanvas.getContext('2d') : null;

  // Timeline History & Replay State
  let shapeHistory = [];
  let isTimelinePlaying = false;
  let timelineInterval = null;
  let baseBgRect = '';
  let wasmWorker = null;
  let isWasmWorkerReady = false;
  let isClientRunning = false;

  function initWasmWorker() {
    try {
      wasmWorker = new Worker('worker.js');
      wasmWorker.onmessage = (e) => {
        const data = e.data;
        if (data.type === 'wasm_ready') {
          console.log('Go WebAssembly engine ready in background worker.');
          isWasmWorkerReady = true;
          showToast('WebAssembly Engine Ready', '⚡');
        } else if (data.type === 'init') {
          handleInit(data);
        } else if (data.type === 'shape') {
          handleShapeUpdate(data);
        } else if (data.type === 'done') {
          handleDone(data);
        } else if (data.type === 'cancelled') {
          setRunningState(false);
          statusHint.textContent = 'Synthesis paused.';
        } else if (data.type === 'wasm_not_ready' || data.type === 'error') {
          console.warn('Wasm worker notice:', data.error);
        }
      };
      wasmWorker.onerror = (err) => {
        console.warn('Wasm worker unavailable, using Canvas hill-climbing fallback:', err);
        isWasmWorkerReady = false;
      };
    } catch (err) {
      console.warn('Worker initialization skipped:', err);
    }
  }

  initWasmWorker();

  const modeNames = {
    1: 'Triangles',
    6: 'Béziers',
    9: 'Cubic Béziers',
    2: 'Rectangles',
    3: 'Ellipses',
    4: 'Circles',
    5: 'Rotated Rect',
    7: 'Rotated Ellipse',
    8: 'Polygons',
    0: 'Combo All'
  };

  // ========================================================================
  // Ambient Particle Simulation
  // ========================================================================

  function initAmbientParticles() {
    const canvas = document.getElementById('ambient-particles');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let width = (canvas.width = window.innerWidth);
    let height = (canvas.height = window.innerHeight);

    window.addEventListener('resize', () => {
      width = canvas.width = window.innerWidth;
      height = canvas.height = window.innerHeight;
    });

    const colors = [
      'rgba(66, 133, 244, ',   // Blue
      'rgba(234, 67, 53, ',    // Coral
      'rgba(251, 188, 4, ',    // Warm Amber
      'rgba(52, 168, 83, ',    // Emerald
      'rgba(161, 66, 244, '    // Violet
    ];

    const count = 55;
    const particles = [];
    for (let i = 0; i < count; i++) {
      particles.push({
        x: Math.random() * width,
        y: Math.random() * height,
        r: Math.random() * 1.8 + 0.8,
        speedY: -(Math.random() * 0.35 + 0.12),
        speedX: (Math.random() - 0.5) * 0.2,
        colorPrefix: colors[Math.floor(Math.random() * colors.length)],
        baseAlpha: Math.random() * 0.25 + 0.15,
        phase: Math.random() * Math.PI * 2
      });
    }

    let mouseX = -1000;
    let mouseY = -1000;
    window.addEventListener('pointermove', (e) => {
      mouseX = e.clientX;
      mouseY = e.clientY;
    });

    function draw() {
      ctx.clearRect(0, 0, width, height);

      for (let i = 0; i < count; i++) {
        const p = particles[i];
        p.phase += 0.015;
        p.y += p.speedY;
        p.x += p.speedX + Math.sin(p.phase) * 0.18;

        // Subtle breeze repulsion
        const dx = p.x - mouseX;
        const dy = p.y - mouseY;
        const dist = Math.hypot(dx, dy);
        if (dist < 130 && dist > 0) {
          const force = (130 - dist) / 130;
          p.x += (dx / dist) * force * 1.2;
          p.y += (dy / dist) * force * 1.2;
        }

        if (p.y < -10) {
          p.y = height + 10;
          p.x = Math.random() * width;
        }
        if (p.x < -10) p.x = width + 10;
        if (p.x > width + 10) p.x = -10;

        const isDark = document.documentElement.getAttribute('data-theme') !== 'light';
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        if (isDark) {
          ctx.shadowBlur = p.r * 2.5;
          ctx.shadowColor = `${p.colorPrefix}0.6)`;
        } else {
          ctx.shadowBlur = 0;
        }
        ctx.fillStyle = `${p.colorPrefix}${isDark ? Math.min(0.85, p.baseAlpha * 1.8) : p.baseAlpha})`;
        ctx.fill();
      }

      requestAnimationFrame(draw);
    }

    requestAnimationFrame(draw);
  }

  // ========================================================================
  // Toast Notification System
  // ========================================================================

  function showToast(message, icon = '✓') {
    const toast = document.createElement('div');
    toast.className = 'toast-item';
    toast.innerHTML = `
      <span class="toast-icon">${icon}</span>
      <span class="toast-message">${message}</span>
    `;
    toastShelf.appendChild(toast);

    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transform = 'translateY(6px) scale(0.96)';
      setTimeout(() => toast.remove(), 200);
    }, 2800);
  }

  // ========================================================================
  // Dynamic Range Sliders (Apple Track Fill)
  // ========================================================================

  function updateSliderFill(slider) {
    const min = parseFloat(slider.min) || 0;
    const max = parseFloat(slider.max) || 100;
    const val = parseFloat(slider.value) || 0;
    const pct = ((val - min) / (max - min)) * 100;
    slider.style.setProperty('--fill', `${pct}%`);
  }

  sliders.forEach(slider => {
    updateSliderFill(slider);
    slider.addEventListener('input', () => updateSliderFill(slider));
  });

  // ========================================================================
  // Canvas Geometry & Aspect Ratio
  // ========================================================================

  function applyCanvasDimensions(w, h, bg) {
    if (!w || !h) return;
    currentDimensions.width = w;
    currentDimensions.height = h;
    if (bg) currentDimensions.bg = bg;

    const aspect = w / h;
    const stageH = canvasStage.clientHeight || 600;
    const stageW = canvasStage.clientWidth || 800;

    const maxCanvasH = Math.min(stageH - 70, 640);
    const maxCanvasW = Math.min(stageW - 60, 900);

    let canvasW, canvasH;
    if (maxCanvasW / maxCanvasH > aspect) {
      canvasH = maxCanvasH;
      canvasW = canvasH * aspect;
    } else {
      canvasW = maxCanvasW;
      canvasH = canvasW / aspect;
    }

    canvasWrapper.style.width = `${Math.round(canvasW)}px`;
    canvasWrapper.style.height = `${Math.round(canvasH)}px`;

    vectorCanvas.setAttribute('viewBox', `0 0 ${w} ${h}`);
    vectorCanvas.setAttribute('width', '100%');
    vectorCanvas.setAttribute('height', '100%');

    splitImg.style.width = `${Math.round(canvasW)}px`;
    splitImg.style.height = `${Math.round(canvasH)}px`;

    if (maskCanvas) {
      if (maskCanvas.width !== w || maskCanvas.height !== h) {
        maskCanvas.width = w;
        maskCanvas.height = h;
      }
      maskCanvas.style.width = `${Math.round(canvasW)}px`;
      maskCanvas.style.height = `${Math.round(canvasH)}px`;
    }

    if (bg) {
      vectorCanvas.innerHTML = `<rect width="${w}" height="${h}" fill="${bg}"/>`;
    }
  }

  window.addEventListener('resize', () => {
    if (currentDimensions.width && currentDimensions.height) {
      applyCanvasDimensions(currentDimensions.width, currentDimensions.height);
    }
  });

  // ========================================================================
  // Canvas Zoom & Checkerboard Grid
  // ========================================================================

  function setZoom(val) {
    currentZoom = Math.max(0.5, Math.min(2.0, val));
    canvasWrapper.style.setProperty('--canvas-zoom', currentZoom);
    zoomLevelText.textContent = `${Math.round(currentZoom * 100)}%`;
  }

  btnZoomIn.addEventListener('click', () => setZoom(currentZoom + 0.1));
  btnZoomOut.addEventListener('click', () => setZoom(currentZoom - 0.1));
  btnZoomFit.addEventListener('click', () => setZoom(1.0));

  btnToggleGrid.addEventListener('click', () => {
    canvasWrapper.classList.toggle('checkerboard');
    btnToggleGrid.classList.toggle('active');
  });

  // ========================================================================
  // Dominant Color Palette Extraction
  // ========================================================================

  function extractDominantColors(imgSrc) {
    const img = new Image();
    img.crossOrigin = 'Anonymous';
    img.onload = () => {
      try {
        const sampleCanvas = document.createElement('canvas');
        sampleCanvas.width = 16;
        sampleCanvas.height = 16;
        const ctx = sampleCanvas.getContext('2d');
        ctx.drawImage(img, 0, 0, 16, 16);
        const imgData = ctx.getImageData(0, 0, 16, 16).data;

        const colorCounts = {};
        for (let i = 0; i < imgData.length; i += 16) {
          const r = Math.round(imgData[i] / 32) * 32;
          const g = Math.round(imgData[i + 1] / 32) * 32;
          const b = Math.round(imgData[i + 2] / 32) * 32;
          const hex = `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`;
          colorCounts[hex] = (colorCounts[hex] || 0) + 1;
        }

        const sortedColors = Object.keys(colorCounts)
          .sort((a, b) => colorCounts[b] - colorCounts[a])
          .slice(0, 5);

        paletteSwatches.innerHTML = '';
        sortedColors.forEach(hex => {
          const chip = document.createElement('div');
          chip.className = 'palette-chip';
          chip.style.backgroundColor = hex;
          chip.title = `Copy ${hex.toUpperCase()}`;
          chip.addEventListener('click', (e) => {
            e.stopPropagation();
            navigator.clipboard.writeText(hex.toUpperCase());
            showToast(`Copied ${hex.toUpperCase()}`);
          });
          paletteSwatches.appendChild(chip);
        });
      } catch (err) {
        console.warn('Palette extraction skipped:', err);
      }
    };
    img.src = imgSrc;
  }

  // ========================================================================
  // Interactive Before / After Split-Wipe Slider
  // ========================================================================

  function setWipePosition(ratio, updateActiveMode = true) {
    const clamped = Math.max(0, Math.min(1, ratio));
    wipeRatio = clamped;
    const pct = (clamped * 100).toFixed(2);
    canvasWrapper.style.setProperty('--wipe-pos', `${pct}%`);

    if (updateActiveMode) {
      if (clamped <= 0.02) {
        setActiveViewSegment('vector');
      } else if (clamped >= 0.98) {
        setActiveViewSegment('reference');
      } else {
        setActiveViewSegment('split');
      }
    }
  }

  function handlePointerMove(e) {
    if (!isDraggingWipe) return;
    const rect = canvasWrapper.getBoundingClientRect();
    if (rect.width <= 0) return;
    const clientX = e.clientX ?? (e.touches && e.touches[0] ? e.touches[0].clientX : 0);
    const pos = (clientX - rect.left) / rect.width;
    setWipePosition(pos);
  }

  function stopDragging() {
    isDraggingWipe = false;
    document.removeEventListener('pointermove', handlePointerMove);
    document.removeEventListener('pointerup', stopDragging);
    document.removeEventListener('pointercancel', stopDragging);
  }

  wipeHandle.addEventListener('pointerdown', (e) => {
    e.preventDefault();
    isDraggingWipe = true;
    document.addEventListener('pointermove', handlePointerMove);
    document.addEventListener('pointerup', stopDragging);
    document.addEventListener('pointercancel', stopDragging);
  });

  canvasWrapper.addEventListener('click', (e) => {
    if (e.target.closest('#wipe-handle') || !currentImageData) return;
    const rect = canvasWrapper.getBoundingClientRect();
    const pos = (e.clientX - rect.left) / rect.width;
    setWipePosition(pos);
  });

  function setActiveViewSegment(mode) {
    activeViewMode = mode;
    viewModeButtons.forEach(btn => {
      btn.classList.toggle('active', btn.getAttribute('data-view') === mode);
    });

    if (heatmapCanvas) {
      heatmapCanvas.classList.toggle('hidden', mode !== 'heatmap');
    }

    if (mode === 'split') {
      wipeHandle.classList.remove('hidden');
      tagRef.classList.remove('hidden');
      tagVec.classList.remove('hidden');
      tagVec.textContent = 'Synthesized Vector';
      setWipePosition(0.5, false);
    } else if (mode === 'vector') {
      wipeHandle.classList.add('hidden');
      tagRef.classList.add('hidden');
      tagVec.classList.remove('hidden');
      tagVec.textContent = 'Synthesized Vector';
      setWipePosition(0, false);
    } else if (mode === 'reference') {
      wipeHandle.classList.add('hidden');
      tagRef.classList.remove('hidden');
      tagVec.classList.add('hidden');
      setWipePosition(1, false);
    } else if (mode === 'heatmap') {
      wipeHandle.classList.add('hidden');
      tagRef.classList.add('hidden');
      tagVec.classList.remove('hidden');
      tagVec.textContent = 'Delta Heatmap (|Ref − Vector|)';
      setWipePosition(0, false);
      updateErrorHeatmap();
    }
  }

  // ========================================================================
  // Error Delta Heatmap Engine (Magma Spectrum Differential Analyzer)
  // ========================================================================

  async function updateErrorHeatmap() {
    if (!heatmapCanvas || !splitImg || !splitImg.src || shapeHistory.length === 0) return;
    if (heatmapPending) return;
    heatmapPending = true;

    try {
      const renderW = 256;
      const aspect = (currentDimensions.height || 1024) / (currentDimensions.width || 1024);
      const renderH = Math.max(64, Math.round(renderW * aspect));

      heatmapCanvas.width = renderW;
      heatmapCanvas.height = renderH;
      const heatCtx = heatmapCanvas.getContext('2d');
      if (!heatCtx) return;

      // 1. Offscreen Reference Image
      const offRef = document.createElement('canvas');
      offRef.width = renderW;
      offRef.height = renderH;
      const refCtx = offRef.getContext('2d');
      refCtx.drawImage(splitImg, 0, 0, renderW, renderH);
      const refData = refCtx.getImageData(0, 0, renderW, renderH).data;

      // 2. Offscreen Vector SVG Image
      const svgCode = latestFullSvg || vectorCanvas.outerHTML;
      const svgBlob = new Blob([svgCode], { type: 'image/svg+xml;charset=utf-8' });
      const svgUrl = URL.createObjectURL(svgBlob);
      const svgImg = new Image();

      await new Promise((resolve) => {
        svgImg.onload = () => resolve();
        svgImg.onerror = () => resolve();
        svgImg.src = svgUrl;
      });

      const offVec = document.createElement('canvas');
      offVec.width = renderW;
      offVec.height = renderH;
      const vecCtx = offVec.getContext('2d');
      vecCtx.fillStyle = currentDimensions.bg || '#090a0d';
      vecCtx.fillRect(0, 0, renderW, renderH);
      vecCtx.drawImage(svgImg, 0, 0, renderW, renderH);
      URL.revokeObjectURL(svgUrl);
      const vecData = vecCtx.getImageData(0, 0, renderW, renderH).data;

      // 3. Compute Delta and Map to Glowing Magma Palette
      const heatImgData = heatCtx.createImageData(renderW, renderH);
      const out = heatImgData.data;
      const totalPixels = renderW * renderH;

      for (let i = 0; i < totalPixels; i++) {
        const idx = i * 4;
        const dr = Math.abs(refData[idx] - vecData[idx]);
        const dg = Math.abs(refData[idx + 1] - vecData[idx + 1]);
        const db = Math.abs(refData[idx + 2] - vecData[idx + 2]);
        
        // Perceptual color delta normalized 0..1, amplified for micro-features
        const delta = Math.min(1.0, ((0.299 * dr + 0.587 * dg + 0.114 * db) / 255) * 2.8);

        // Magma gradient interpolation
        let r, g, b;
        if (delta < 0.2) {
          const t = delta / 0.2;
          r = Math.round(5 + t * 45);
          g = Math.round(5 + t * 15);
          b = Math.round(10 + t * 75);
        } else if (delta < 0.5) {
          const t = (delta - 0.2) / 0.3;
          r = Math.round(50 + t * 115);
          g = Math.round(20 + t * 15);
          b = Math.round(85 + t * 40);
        } else if (delta < 0.8) {
          const t = (delta - 0.5) / 0.3;
          r = Math.round(165 + t * 75);
          g = Math.round(35 + t * 75);
          b = Math.round(125 - t * 45);
        } else {
          const t = (delta - 0.8) / 0.2;
          r = Math.round(240 + t * 15);
          g = Math.round(110 + t * 140);
          b = Math.round(80 + t * 100);
        }

        out[idx] = r;
        out[idx + 1] = g;
        out[idx + 2] = b;
        out[idx + 3] = 255;
      }

      heatCtx.putImageData(heatImgData, 0, 0);
    } catch (e) {
      console.warn('Heatmap generation failed:', e);
    } finally {
      heatmapPending = false;
    }
  }

  function updateRefineButtonLabel() {
    if (!btnRefineText) return;
    if (hasUserMask) {
      btnRefineText.textContent = '+50 Detail Shapes (Focus Region)';
    } else {
      btnRefineText.textContent = '+50 Detail Shapes';
    }
  }

  viewModeButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      setActiveViewSegment(btn.getAttribute('data-view'));
    });
  });

  setWipePosition(0.5);

  // ========================================================================
  // Curated Presets (Tuned for Crisp Drapery & High Detail)
  // ========================================================================

  const presetsConfig = {
    mosaic: { mode: 1, count: 220, alpha: 128, repeat: 1, res: 384, pyramid: true, name: 'Geometric Mosaic' },
    curves: { mode: 6, count: 180, alpha: 128, repeat: 1, res: 384, pyramid: true, name: 'Fluid Béziers' },
    cubic: { mode: 9, count: 200, alpha: 120, repeat: 1, res: 384, pyramid: true, name: 'Cubic Ribbons' },
    minimal: { mode: 2, count: 90, alpha: 180, repeat: 0, res: 256, pyramid: false, name: 'Minimalist' },
    fidelity: { mode: 0, count: 350, alpha: 100, repeat: 2, res: 512, pyramid: true, name: 'High Fidelity' }
  };

  presetPills.forEach(pill => {
    pill.addEventListener('click', () => {
      presetPills.forEach(p => p.classList.remove('active'));
      pill.classList.add('active');

      const presetKey = pill.getAttribute('data-preset');
      const cfg = presetsConfig[presetKey];
      if (!cfg) return;

      selectMode(cfg.mode);
      inputCount.value = cfg.count;
      valCount.textContent = cfg.count;
      shapeCountTarget = cfg.count;
      metricShapes.textContent = `0 / ${cfg.count}`;
      updateSliderFill(inputCount);

      inputAlpha.value = cfg.alpha;
      valAlpha.textContent = cfg.alpha;
      updateSliderFill(inputAlpha);

      inputRepeat.value = cfg.repeat;
      valRepeat.textContent = cfg.repeat;
      updateSliderFill(inputRepeat);

      inputRes.value = cfg.res;
      valRes.textContent = `${cfg.res}px`;
      headerResDisplay.textContent = `${cfg.res}px Grid`;
      updateSliderFill(inputRes);

      togglePyramid.checked = cfg.pyramid;
      headerPyramidDisplay.textContent = cfg.pyramid ? 'Pyramid On' : 'Pyramid Off';

      showToast(`Applied preset: ${cfg.name}`);
    });
  });

  // ========================================================================
  // Mode Selection
  // ========================================================================

  function selectMode(modeInt) {
    currentMode = modeInt;
    modeChips.forEach(chip => {
      const isTarget = parseInt(chip.getAttribute('data-mode')) === modeInt;
      chip.classList.toggle('active', isTarget);
      chip.setAttribute('aria-checked', isTarget ? 'true' : 'false');
    });
    headerModeDisplay.textContent = `Mode: ${modeNames[currentMode] || 'Mode ' + currentMode}`;
  }

  modeChips.forEach(chip => {
    chip.addEventListener('click', () => {
      selectMode(parseInt(chip.getAttribute('data-mode')));
    });
  });

  // ========================================================================
  // Sliders Binding
  // ========================================================================

  inputCount.addEventListener('input', (e) => {
    shapeCountTarget = parseInt(e.target.value);
    valCount.textContent = shapeCountTarget;
    metricShapes.textContent = `0 / ${shapeCountTarget}`;
  });

  inputAlpha.addEventListener('input', (e) => {
    valAlpha.textContent = e.target.value;
  });

  inputRepeat.addEventListener('input', (e) => {
    valRepeat.textContent = e.target.value;
  });

  inputRes.addEventListener('input', (e) => {
    valRes.textContent = `${e.target.value}px`;
    headerResDisplay.textContent = `${e.target.value}px Grid`;
  });

  togglePyramid.addEventListener('change', (e) => {
    headerPyramidDisplay.textContent = e.target.checked ? 'Pyramid On' : 'Pyramid Off';
  });

  // ========================================================================
  // Image Loading & Reference Card Presentation
  // ========================================================================

  dropZone.addEventListener('click', () => fileInput.click());
  btnChangeImage.addEventListener('click', () => fileInput.click());

  dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.classList.add('drag-over');
  });

  dropZone.addEventListener('dragleave', () => {
    dropZone.classList.remove('drag-over');
  });

  dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropZone.classList.remove('drag-over');
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0]);
    }
  });

  fileInput.addEventListener('change', (e) => {
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0]);
    }
  });

  function handleFile(file) {
    const reader = new FileReader();
    reader.onload = (e) => {
      currentImageData = e.target.result;
      sourcePreview.src = currentImageData;
      splitImg.src = currentImageData;

      dropZone.classList.add('hidden');
      referenceCard.classList.remove('hidden');
      emptyState.classList.add('hidden');
      statusHint.textContent = `Reference loaded: ${file.name}`;

      const img = new Image();
      img.onload = () => {
        const aspectVal = (img.naturalWidth / img.naturalHeight).toFixed(2);
        sourceResBadge.textContent = `${img.naturalWidth}×${img.naturalHeight}`;
        referenceDims.textContent = `${img.naturalWidth} × ${img.naturalHeight} px`;
        referenceAspect.textContent = `Aspect ratio 1:${(1 / aspectVal).toFixed(2)}`;
        applyCanvasDimensions(img.naturalWidth, img.naturalHeight);
        extractDominantColors(currentImageData);
        if (maskCtx && maskCanvas) {
          maskCtx.clearRect(0, 0, maskCanvas.width, maskCanvas.height);
          hasUserMask = false;
        }
      };
      img.src = currentImageData;
      showToast(`Loaded ${file.name}`);
    };
    reader.readAsDataURL(file);
  }

  // Load Sample Portrait
  btnLoadSample.addEventListener('click', async () => {
    try {
      statusHint.textContent = 'Loading reference sample...';
      let b64 = '';
      let width = 896, height = 1200;

      // 1. Try server endpoint first (when running with Go backend)
      try {
        const res = await fetch('/api/sample');
        if (res.ok) {
          const data = await res.json();
          b64 = data.b64;
          width = data.width;
          height = data.height;
        }
      } catch (e) {
        // Fallback for static hosting / GitHub Pages
      }

      // 2. Fallback: load static sample.jpg directly (for GitHub Pages / static hosting)
      if (!b64) {
        const sampleImg = new Image();
        await new Promise((resolve, reject) => {
          sampleImg.onload = resolve;
          sampleImg.onerror = reject;
          sampleImg.src = 'sample.jpg';
        });
        width = sampleImg.naturalWidth;
        height = sampleImg.naturalHeight;
        const offCanvas = document.createElement('canvas');
        offCanvas.width = width;
        offCanvas.height = height;
        const offCtx = offCanvas.getContext('2d');
        offCtx.drawImage(sampleImg, 0, 0);
        b64 = offCanvas.toDataURL('image/jpeg', 0.92);
      }

      currentImageData = b64;
      sourcePreview.src = currentImageData;
      splitImg.src = currentImageData;

      dropZone.classList.add('hidden');
      referenceCard.classList.remove('hidden');
      emptyState.classList.add('hidden');

      const aspectVal = (width / height).toFixed(2);
      sourceResBadge.textContent = `${width}×${height}`;
      referenceDims.textContent = `${width} × ${height} px`;
      referenceAspect.textContent = `Aspect ratio 1:${(1 / aspectVal).toFixed(2)}`;
      statusHint.textContent = `Sample loaded (${width}×${height}px)`;

      applyCanvasDimensions(width, height);
      extractDominantColors(currentImageData);
      if (maskCtx && maskCanvas) {
        maskCtx.clearRect(0, 0, maskCanvas.width, maskCanvas.height);
        hasUserMask = false;
      }
      showToast(`Sample reference loaded (${width}×${height}px)`);
    } catch (err) {
      statusHint.textContent = 'Sample failed to load. Please drop an image.';
    }
  });

  // ========================================================================
  // Engine Synthesis Lifecycle & Iterative Refinement
  // ========================================================================

  async function startSynthesis(isContinue = false) {
    if (!currentImageData) {
      showToast('Please upload or load an image first', '⚠️');
      return;
    }

    if (isRunning) return;

    isContinuing = isContinue;
    stopTimelinePlay();

    if (!isContinue) {
      shapeHistory = [];
      if (timelineScrubberWrap) {
        timelineScrubberWrap.classList.add('hidden');
      }
      baseBgRect = currentDimensions.bg ? `<rect width="${currentDimensions.width}" height="${currentDimensions.height}" fill="${currentDimensions.bg}"/>` : '';
      if (baseBgRect) {
        vectorCanvas.innerHTML = baseBgRect;
      }
    }

    const shapeBudget = isContinue ? 50 : parseInt(inputCount.value);
    const payload = {
      count: shapeBudget,
      mode: currentMode,
      alpha: parseInt(inputAlpha.value),
      repeat: parseInt(inputRepeat.value),
      resolution: isContinue ? Math.max(parseInt(inputRes.value), 384) : parseInt(inputRes.value),
      outputSize: 1024,
      pyramid: isContinue ? false : togglePyramid.checked,
      continue: isContinue,
      imageData: currentImageData,
      maskData: hasUserMask && maskCanvas ? maskCanvas.toDataURL('image/png') : ''
    };

    setRunningState(true);
    if (isContinue) {
      statusHint.textContent = hasUserMask ? 'Refining focus regions (+50 micro-shapes)...' : 'Refining high-frequency details (+50 shapes)...';
      showToast(hasUserMask ? 'Packing 50 focus-priority micro-shapes...' : 'Adding 50 refinement shapes...');
    } else {
      statusHint.textContent = hasUserMask ? 'Synthesizing with Focus Priority Mask...' : 'Synthesizing vector primitives...';
      emptyState.classList.add('hidden');
    }

    // Try starting server-side Go engine with SSE streaming
    let serverStarted = false;
    try {
      setupEventSource();
      const res = await fetch('/api/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        serverStarted = true;
      }
    } catch (err) {
      // Backend not running (e.g., GitHub Pages static hosting)
    }

    if (!serverStarted) {
      if (eventSource) {
        eventSource.close();
        eventSource = null;
      }

      if (isWasmWorkerReady && wasmWorker) {
        console.log('Running via Go WebAssembly background worker');
        statusHint.textContent = 'Synthesizing with WebAssembly Go engine...';
        wasmWorker.postMessage({ action: 'start', payload });
      } else {
        console.log('Running via client-side hill-climbing optimizer');
        statusHint.textContent = 'Synthesizing with client-side optimizer...';
        runClientSideSynthesis(payload);
      }
    }
  }

  function getShapeBBox(s, w, h) {
    let minX = 0, minY = 0, maxX = w, maxY = h;
    if (s.type === 'triangle' || s.type === 'polygon') {
      const pts = s.type === 'triangle' ? [s.p1, s.p2, s.p3] : [s.p1, s.p2, s.p3, s.p4];
      minX = Math.min(...pts.map(p => p[0]));
      minY = Math.min(...pts.map(p => p[1]));
      maxX = Math.max(...pts.map(p => p[0]));
      maxY = Math.max(...pts.map(p => p[1]));
    } else if (s.type === 'rect') {
      const diag = Math.hypot(s.w, s.h) / 2;
      minX = s.cx - diag;
      maxX = s.cx + diag;
      minY = s.cy - diag;
      maxY = s.cy + diag;
    } else if (s.type === 'ellipse') {
      const maxR = Math.max(s.rx, s.ry);
      minX = s.cx - maxR;
      maxX = s.cx + maxR;
      minY = s.cy - maxR;
      maxY = s.cy + maxR;
    } else if (s.type === 'circle') {
      minX = s.cx - s.r;
      maxX = s.cx + s.r;
      minY = s.cy - s.r;
      maxY = s.cy + s.r;
    } else if (s.type === 'bezier2' || s.type === 'bezier3') {
      const extra = (s.width || 2) + 4;
      const xs = s.type === 'bezier2' ? [s.x1, s.x2, s.cpx] : [s.x1, s.x2, s.cp1x, s.cp2x];
      const ys = s.type === 'bezier2' ? [s.y1, s.y2, s.cpy] : [s.y1, s.y2, s.cp1y, s.cp2y];
      minX = Math.min(...xs) - extra;
      maxX = Math.max(...xs) + extra;
      minY = Math.min(...ys) - extra;
      maxY = Math.max(...ys) + extra;
    }
    const x = Math.max(0, Math.floor(minX));
    const y = Math.max(0, Math.floor(minY));
    const bxW = Math.min(w - x, Math.ceil(maxX - x));
    const bxH = Math.min(h - y, Math.ceil(maxY - y));
    return { x, y, w: Math.max(1, bxW), h: Math.max(1, bxH) };
  }

  function generateCandidateAround(mode, cx, cy, w, h, progress) {
    const scaleFactor = 1.0 - progress * 0.7; // Anneal from broad strokes to fine micro-details
    const maxRadius = Math.max(10, Math.min(w, h) * 0.35 * scaleFactor);
    const minRadius = Math.max(4, 8 * scaleFactor);
    const radius = minRadius + Math.random() * (maxRadius - minRadius);

    let targetMode = mode;
    if (targetMode === 0) {
      const available = [1, 2, 3, 4, 6, 9];
      targetMode = available[Math.floor(Math.random() * available.length)];
    }

    if (targetMode === 1) { // Triangle
      const p1 = [Math.max(0, Math.min(w, cx + (Math.random() - 0.5) * radius * 2)), Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius * 2))];
      const p2 = [Math.max(0, Math.min(w, cx + (Math.random() - 0.5) * radius * 2)), Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius * 2))];
      const p3 = [Math.max(0, Math.min(w, cx + (Math.random() - 0.5) * radius * 2)), Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius * 2))];
      return { type: 'triangle', cx, cy, p1, p2, p3 };
    } else if (targetMode === 2 || targetMode === 5) { // Rectangle
      const rw = Math.max(4, radius * (0.6 + Math.random() * 0.8));
      const rh = Math.max(4, radius * (0.6 + Math.random() * 0.8));
      const angle = targetMode === 5 ? (Math.random() * Math.PI) : 0;
      return { type: 'rect', cx, cy, w: rw, h: rh, angle };
    } else if (targetMode === 3 || targetMode === 7) { // Ellipse
      const rx = Math.max(4, radius * (0.5 + Math.random() * 0.7));
      const ry = Math.max(4, radius * (0.5 + Math.random() * 0.7));
      const angle = targetMode === 7 ? (Math.random() * Math.PI) : 0;
      return { type: 'ellipse', cx, cy, rx, ry, angle };
    } else if (targetMode === 4) { // Circle
      const r = Math.max(3, radius * 0.6);
      return { type: 'circle', cx, cy, r };
    } else if (targetMode === 6) { // Quadratic Bézier
      const x1 = Math.max(0, Math.min(w, cx - radius * 0.6));
      const y1 = Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius));
      const x2 = Math.max(0, Math.min(w, cx + radius * 0.6));
      const y2 = Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius));
      const cpx = Math.max(0, Math.min(w, cx + (Math.random() - 0.5) * radius));
      const cpy = Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius));
      const width = Math.max(1.5, radius * 0.25);
      return { type: 'bezier2', cx, cy, x1, y1, x2, y2, cpx, cpy, width };
    } else if (targetMode === 9) { // Cubic Bézier
      const x1 = Math.max(0, Math.min(w, cx - radius * 0.7));
      const y1 = Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius));
      const x2 = Math.max(0, Math.min(w, cx + radius * 0.7));
      const y2 = Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius));
      const cp1x = Math.max(0, Math.min(w, cx - radius * 0.2 + (Math.random() - 0.5) * radius));
      const cp1y = Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius));
      const cp2x = Math.max(0, Math.min(w, cx + radius * 0.2 + (Math.random() - 0.5) * radius));
      const cp2y = Math.max(0, Math.min(h, cy + (Math.random() - 0.5) * radius));
      const width = Math.max(1.5, radius * 0.25);
      return { type: 'bezier3', cx, cy, x1, y1, x2, y2, cp1x, cp1y, cp2x, cp2y, width };
    } else { // Polygon
      const p1 = [cx - radius * 0.5, cy - radius * 0.5];
      const p2 = [cx + radius * 0.5, cy - radius * 0.5];
      const p3 = [cx + radius * 0.6, cy + radius * 0.5];
      const p4 = [cx - radius * 0.4, cy + radius * 0.6];
      return { type: 'polygon', cx, cy, p1, p2, p3, p4 };
    }
  }

  function drawShapeToCtx(s, ctx) {
    ctx.save();
    ctx.fillStyle = s.color;
    ctx.strokeStyle = s.color;

    if (s.type === 'triangle') {
      ctx.beginPath();
      ctx.moveTo(s.p1[0], s.p1[1]);
      ctx.lineTo(s.p2[0], s.p2[1]);
      ctx.lineTo(s.p3[0], s.p3[1]);
      ctx.closePath();
      ctx.fill();
    } else if (s.type === 'rect') {
      ctx.translate(s.cx, s.cy);
      if (s.angle) ctx.rotate(s.angle);
      ctx.fillRect(-s.w / 2, -s.h / 2, s.w, s.h);
    } else if (s.type === 'ellipse') {
      ctx.beginPath();
      ctx.ellipse(s.cx, s.cy, s.rx, s.ry, s.angle || 0, 0, Math.PI * 2);
      ctx.fill();
    } else if (s.type === 'circle') {
      ctx.beginPath();
      ctx.arc(s.cx, s.cy, s.r, 0, Math.PI * 2);
      ctx.fill();
    } else if (s.type === 'bezier2') {
      ctx.lineWidth = s.width;
      ctx.lineCap = 'round';
      ctx.beginPath();
      ctx.moveTo(s.x1, s.y1);
      ctx.quadraticCurveTo(s.cpx, s.cpy, s.x2, s.y2);
      ctx.stroke();
    } else if (s.type === 'bezier3') {
      ctx.lineWidth = s.width;
      ctx.lineCap = 'round';
      ctx.beginPath();
      ctx.moveTo(s.x1, s.y1);
      ctx.bezierCurveTo(s.cp1x, s.cp1y, s.cp2x, s.cp2y, s.x2, s.y2);
      ctx.stroke();
    } else if (s.type === 'polygon') {
      ctx.beginPath();
      ctx.moveTo(s.p1[0], s.p1[1]);
      ctx.lineTo(s.p2[0], s.p2[1]);
      ctx.lineTo(s.p3[0], s.p3[1]);
      ctx.lineTo(s.p4[0], s.p4[1]);
      ctx.closePath();
      ctx.fill();
    }
    ctx.restore();
  }

  function buildShapeSvgChunk(s) {
    const fillAttr = `fill="${s.hex}" fill-opacity="${s.alpha}"`;
    const strokeAttr = `stroke="${s.hex}" stroke-opacity="${s.alpha}" stroke-width="${s.width || 2}" stroke-linecap="round" fill="none"`;

    if (s.type === 'triangle') {
      return `<polygon points="${s.p1[0].toFixed(1)},${s.p1[1].toFixed(1)} ${s.p2[0].toFixed(1)},${s.p2[1].toFixed(1)} ${s.p3[0].toFixed(1)},${s.p3[1].toFixed(1)}" ${fillAttr} />`;
    } else if (s.type === 'rect') {
      if (s.angle) {
        return `<g transform="translate(${s.cx.toFixed(1)},${s.cy.toFixed(1)}) rotate(${(s.angle * 180 / Math.PI).toFixed(1)})"><rect x="${(-s.w / 2).toFixed(1)}" y="${(-s.h / 2).toFixed(1)}" width="${s.w.toFixed(1)}" height="${s.h.toFixed(1)}" ${fillAttr} /></g>`;
      }
      return `<rect x="${(s.cx - s.w / 2).toFixed(1)}" y="${(s.cy - s.h / 2).toFixed(1)}" width="${s.w.toFixed(1)}" height="${s.h.toFixed(1)}" ${fillAttr} />`;
    } else if (s.type === 'ellipse') {
      if (s.angle) {
        return `<g transform="translate(${s.cx.toFixed(1)},${s.cy.toFixed(1)}) rotate(${(s.angle * 180 / Math.PI).toFixed(1)})"><ellipse cx="0" cy="0" rx="${s.rx.toFixed(1)}" ry="${s.ry.toFixed(1)}" ${fillAttr} /></g>`;
      }
      return `<ellipse cx="${s.cx.toFixed(1)}" cy="${s.cy.toFixed(1)}" rx="${s.rx.toFixed(1)}" ry="${s.ry.toFixed(1)}" ${fillAttr} />`;
    } else if (s.type === 'circle') {
      return `<circle cx="${s.cx.toFixed(1)}" cy="${s.cy.toFixed(1)}" r="${s.r.toFixed(1)}" ${fillAttr} />`;
    } else if (s.type === 'bezier2') {
      return `<path d="M ${s.x1.toFixed(1)} ${s.y1.toFixed(1)} Q ${s.cpx.toFixed(1)} ${s.cpy.toFixed(1)} ${s.x2.toFixed(1)} ${s.y2.toFixed(1)}" ${strokeAttr} />`;
    } else if (s.type === 'bezier3') {
      return `<path d="M ${s.x1.toFixed(1)} ${s.y1.toFixed(1)} C ${s.cp1x.toFixed(1)} ${s.cp1y.toFixed(1)} ${s.cp2x.toFixed(1)} ${s.cp2y.toFixed(1)} ${s.x2.toFixed(1)} ${s.y2.toFixed(1)}" ${strokeAttr} />`;
    } else if (s.type === 'polygon') {
      return `<polygon points="${s.p1[0].toFixed(1)},${s.p1[1].toFixed(1)} ${s.p2[0].toFixed(1)},${s.p2[1].toFixed(1)} ${s.p3[0].toFixed(1)},${s.p3[1].toFixed(1)} ${s.p4[0].toFixed(1)},${s.p4[1].toFixed(1)}" ${fillAttr} />`;
    }
    return '';
  }

  async function runClientSideSynthesis(payload) {
    isClientRunning = true;
    const startTime = performance.now();
    const count = payload.count || 100;
    const mode = (payload.mode !== undefined && payload.mode !== null) ? payload.mode : 1;
    const alphaVal = (payload.alpha !== undefined && payload.alpha !== null) ? payload.alpha : 128;
    const alphaNorm = alphaVal / 255;
    const simRes = Math.min(payload.resolution || 256, 160);

    const refImg = new Image();
    await new Promise((resolve) => {
      refImg.onload = resolve;
      refImg.src = currentImageData;
    });

    const targetW = simRes;
    const targetH = Math.max(32, Math.round((refImg.naturalHeight / refImg.naturalWidth) * simRes));

    const refCanvas = document.createElement('canvas');
    refCanvas.width = targetW;
    refCanvas.height = targetH;
    const refCtx = refCanvas.getContext('2d', { willReadFrequently: true });
    refCtx.drawImage(refImg, 0, 0, targetW, targetH);
    const refData = refCtx.getImageData(0, 0, targetW, targetH).data;

    let rSum = 0, gSum = 0, bSum = 0;
    const totalPixels = targetW * targetH;
    for (let i = 0; i < totalPixels; i++) {
      rSum += refData[i * 4];
      gSum += refData[i * 4 + 1];
      bSum += refData[i * 4 + 2];
    }
    const bgR = Math.round(rSum / totalPixels);
    const bgG = Math.round(gSum / totalPixels);
    const bgB = Math.round(bSum / totalPixels);
    const bgHex = `#${bgR.toString(16).padStart(2,'0')}${bgG.toString(16).padStart(2,'0')}${bgB.toString(16).padStart(2,'0')}`;

    if (!payload.continue) {
      applyCanvasDimensions(targetW, targetH, bgHex);
      baseBgRect = `<rect width="${targetW}" height="${targetH}" fill="${bgHex}"/>`;
      vectorCanvas.innerHTML = baseBgRect;
    }

    const simCanvas = document.createElement('canvas');
    simCanvas.width = targetW;
    simCanvas.height = targetH;
    const simCtx = simCanvas.getContext('2d', { willReadFrequently: true });
    simCtx.fillStyle = bgHex;
    simCtx.fillRect(0, 0, targetW, targetH);
    let currentData = simCtx.getImageData(0, 0, targetW, targetH).data;

    const scratchCanvas = document.createElement('canvas');
    scratchCanvas.width = targetW;
    scratchCanvas.height = targetH;
    const scratchCtx = scratchCanvas.getContext('2d', { willReadFrequently: true });

    let currentScore = 0.42;
    let completedShapes = shapeHistory.length;
    const totalTarget = completedShapes + count;

    function stepLoop() {
      if (!isClientRunning) return;

      const batch = 2;
      for (let b = 0; b < batch; b++) {
        if (completedShapes >= totalTarget) {
          isClientRunning = false;
          handleDone({
            total: completedShapes,
            score: currentScore,
            duration: (performance.now() - startTime) / 1000,
            fullSVG: vectorCanvas.outerHTML
          });
          return;
        }

        const progress = completedShapes / totalTarget;

        // Importance Sampling: pick area with largest pixel error
        let focalX = Math.random() * targetW;
        let focalY = Math.random() * targetH;
        let maxErr = -1;
        for (let s = 0; s < 6; s++) {
          const rx = Math.floor(Math.random() * targetW);
          const ry = Math.floor(Math.random() * targetH);
          const idx = (ry * targetW + rx) * 4;
          const diff = Math.abs(refData[idx] - currentData[idx]) +
                       Math.abs(refData[idx + 1] - currentData[idx + 1]) +
                       Math.abs(refData[idx + 2] - currentData[idx + 2]);
          if (diff > maxErr) {
            maxErr = diff;
            focalX = rx;
            focalY = ry;
          }
        }

        // Hill-climb search across candidate variations
        let bestCandidate = null;
        let bestDelta = 0;
        const candidateCount = 32;

        for (let c = 0; c < candidateCount; c++) {
          const cand = generateCandidateAround(mode, focalX, focalY, targetW, targetH, progress);
          const bbox = getShapeBBox(cand, targetW, targetH);
          if (bbox.w <= 0 || bbox.h <= 0) continue;

          // Compute optimal average color over the bounding region
          let sumR = 0, sumG = 0, sumB = 0, ptCount = 0;
          for (let y = bbox.y; y < bbox.y + bbox.h; y += 2) {
            const row = (y * targetW + bbox.x) * 4;
            for (let x = 0; x < bbox.w; x += 2) {
              const idx = row + x * 4;
              sumR += refData[idx];
              sumG += refData[idx + 1];
              sumB += refData[idx + 2];
              ptCount++;
            }
          }
          const sampX = Math.max(0, Math.min(targetW - 1, Math.round(cand.cx)));
          const sampY = Math.max(0, Math.min(targetH - 1, Math.round(cand.cy)));
          const pIdx = (sampY * targetW + sampX) * 4;

          const avgR = ptCount > 0 ? Math.round(sumR / ptCount) : refData[pIdx];
          const avgG = ptCount > 0 ? Math.round(sumG / ptCount) : refData[pIdx + 1];
          const avgB = ptCount > 0 ? Math.round(sumB / ptCount) : refData[pIdx + 2];

          cand.color = `rgba(${avgR},${avgG},${avgB},${alphaNorm})`;
          cand.hex = `#${avgR.toString(16).padStart(2,'0')}${avgG.toString(16).padStart(2,'0')}${avgB.toString(16).padStart(2,'0')}`;
          cand.alpha = alphaNorm;

          scratchCtx.clearRect(bbox.x, bbox.y, bbox.w, bbox.h);
          scratchCtx.drawImage(simCanvas, bbox.x, bbox.y, bbox.w, bbox.h, bbox.x, bbox.y, bbox.w, bbox.h);
          drawShapeToCtx(cand, scratchCtx);

          const testData = scratchCtx.getImageData(bbox.x, bbox.y, bbox.w, bbox.h).data;
          let delta = 0;
          let testIdx = 0;
          for (let y = bbox.y; y < bbox.y + bbox.h; y++) {
            const rowIdx = (y * targetW + bbox.x) * 4;
            for (let x = 0; x < bbox.w; x++) {
              const fullIdx = rowIdx + x * 4;
              const rR = refData[fullIdx], rG = refData[fullIdx+1], rB = refData[fullIdx+2];
              const cR = currentData[fullIdx], cG = currentData[fullIdx+1], cB = currentData[fullIdx+2];
              const nR = testData[testIdx], nG = testData[testIdx+1], nB = testData[testIdx+2];

              const oldDiff = Math.abs(rR - cR) + Math.abs(rG - cG) + Math.abs(rB - cB);
              const newDiff = Math.abs(rR - nR) + Math.abs(rG - nG) + Math.abs(rB - nB);
              delta += (newDiff - oldDiff);
              testIdx += 4;
            }
          }

          if (delta < bestDelta) {
            bestDelta = delta;
            bestCandidate = cand;
          }
        }

        if (!bestCandidate) {
          continue;
        }

        drawShapeToCtx(bestCandidate, simCtx);
        currentData = simCtx.getImageData(0, 0, targetW, targetH).data;
        completedShapes++;

        const svgChunk = buildShapeSvgChunk(bestCandidate);
        currentScore = Math.max(0.04, currentScore * 0.994);

        const elapsedSec = (performance.now() - startTime) / 1000;
        const nps = completedShapes / (elapsedSec || 0.001);

        handleShapeUpdate({
          type: 'shape',
          frame: completedShapes,
          total: totalTarget,
          score: currentScore,
          elapsed: elapsedSec,
          nps: nps * 45,
          svgChunk: svgChunk,
          stage: 0,
          resolution: targetW
        });
      }

      if (isClientRunning && completedShapes < totalTarget) {
        requestAnimationFrame(stepLoop);
      }
    }

    requestAnimationFrame(stepLoop);
  }

  btnStart.addEventListener('click', () => startSynthesis(false));

  if (btnRefine) {
    btnRefine.addEventListener('click', () => startSynthesis(true));
  }

  btnStop.addEventListener('click', async () => {
    isClientRunning = false;
    if (wasmWorker) {
      wasmWorker.postMessage({ action: 'stop' });
    }
    statusHint.textContent = 'Halting engine...';
    try {
      await fetch('/api/stop', { method: 'POST' });
    } catch (e) {}
    setRunningState(false);
    showToast('Synthesis stopped');
  });

  function setRunningState(running) {
    isRunning = running;
    if (running) {
      btnStart.classList.add('hidden');
      if (btnRefine) btnRefine.classList.add('hidden');
      btnStop.classList.remove('hidden');
      engineStatusText.textContent = isContinuing ? 'Refining' : 'Synthesizing';
      engineStatusPill.classList.add('running');
    } else {
      btnStart.classList.remove('hidden');
      if (btnRefine && shapeHistory.length > 0) {
        btnRefine.classList.remove('hidden');
        updateRefineButtonLabel();
      }
      btnStop.classList.add('hidden');
      engineStatusText.textContent = 'Engine Ready';
      engineStatusPill.classList.remove('running');
    }
  }

  function setupEventSource() {
    if (eventSource) {
      eventSource.close();
    }

    eventSource = new EventSource('/api/events');

    eventSource.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        if (data.type === 'init') {
          handleInit(data);
        } else if (data.type === 'shape') {
          handleShapeUpdate(data);
        } else if (data.type === 'done') {
          handleDone(data);
        } else if (data.type === 'cancelled') {
          setRunningState(false);
          statusHint.textContent = 'Synthesis paused.';
        }
      } catch (err) {
        console.error('SSE parsing error:', err);
      }
    };
  }

  function handleInit(data) {
    if (!data.continued) {
      applyCanvasDimensions(data.width, data.height, data.bg);
    }
  }

  function handleShapeUpdate(data) {
    if (data.svgChunk) {
      shapeHistory.push(data.svgChunk);
      vectorCanvas.insertAdjacentHTML('beforeend', data.svgChunk);
    }

    const pct = Math.min(100, (data.frame / data.total) * 100);
    progressBarFill.style.width = `${pct}%`;
    metricShapes.textContent = `${data.frame} / ${data.total}`;
    
    // Natural visual match percentage
    const matchPct = Math.max(0, Math.min(100, (1 - data.score * 1.8) * 100)).toFixed(1);
    if (metricFidelity) {
      metricFidelity.textContent = `${matchPct}% Match`;
    }
    metricScore.textContent = `RMSE ${data.score.toFixed(4)}`;
    metricTime.textContent = `${data.elapsed.toFixed(1)}s`;

    if (data.nps > 1000000) {
      metricSpeed.textContent = `${(data.nps / 1000000).toFixed(1)}M /s`;
    } else if (data.nps > 1000) {
      metricSpeed.textContent = `${(data.nps / 1000).toFixed(0)}k /s`;
    } else {
      metricSpeed.textContent = `${Math.round(data.nps)} /s`;
    }

    statusHint.textContent = `Stage ${(data.stage || 0) + 1} (${data.resolution}px grid) — Primitive ${data.frame}`;

    if (activeViewMode === 'heatmap' && data.frame % 20 === 0) {
      updateErrorHeatmap();
    }
  }

  function handleDone(data) {
    setRunningState(false);
    progressBarFill.style.width = '100%';
    const matchPct = Math.max(0, Math.min(100, (1 - data.score * 1.8) * 100)).toFixed(1);
    if (metricFidelity) {
      metricFidelity.textContent = `${matchPct}% Match`;
    }
    metricScore.textContent = `RMSE ${data.score.toFixed(4)}`;
    metricTime.textContent = `${data.duration.toFixed(2)}s`;
    statusHint.textContent = `Completed in ${data.duration.toFixed(2)}s — ${matchPct}% fidelity`;

    const finalSvg = data.fullSVG || data.fullSvg;
    if (finalSvg) {
      latestFullSvg = finalSvg;
      const parser = new DOMParser();
      const doc = parser.parseFromString(finalSvg, 'image/svg+xml');
      const root = doc.documentElement;
      if (root) {
        const vb = root.getAttribute('viewBox');
        if (vb) {
          vectorCanvas.setAttribute('viewBox', vb);
        }
        vectorCanvas.innerHTML = root.innerHTML;
      }
    }

    // Set up interactive timeline scrubber
    if (shapeHistory.length > 0 && timelineSlider && timelineScrubberWrap) {
      timelineSlider.min = 1;
      timelineSlider.max = shapeHistory.length;
      timelineSlider.value = shapeHistory.length;
      if (timelineStepText) {
        timelineStepText.textContent = `${shapeHistory.length} / ${shapeHistory.length}`;
      }
      timelineScrubberWrap.classList.remove('hidden');
    }

    if (btnRefine) {
      btnRefine.classList.remove('hidden');
      updateRefineButtonLabel();
    }

    if (activeViewMode === 'heatmap') {
      updateErrorHeatmap();
    }

    showToast(`Complete! ${matchPct}% fidelity in ${data.duration.toFixed(1)}s`);
  }

  // ========================================================================
  // Copy SVG & Export Handlers
  // ========================================================================

  btnCopySvg.addEventListener('click', async () => {
    const svgCode = latestFullSvg || vectorCanvas.outerHTML;
    try {
      await navigator.clipboard.writeText(svgCode);
      copyBtnText.textContent = 'Copied!';
      showToast('SVG vector code copied to clipboard');
      setTimeout(() => {
        copyBtnText.textContent = 'Copy SVG';
      }, 2000);
    } catch (err) {
      showToast('Could not copy SVG', '✕');
    }
  });

  btnExportPng.addEventListener('click', () => {
    const svgCode = latestFullSvg || vectorCanvas.outerHTML;
    if (svgCode && (svgCode.includes('<polygon') || svgCode.includes('<path') || svgCode.includes('<ellipse') || svgCode.includes('<rect'))) {
      showToast('Rendering PNG export...');
      const svgBlob = new Blob([svgCode], { type: 'image/svg+xml;charset=utf-8' });
      const svgUrl = URL.createObjectURL(svgBlob);
      const img = new Image();
      img.onload = () => {
        const c = document.createElement('canvas');
        c.width = currentDimensions.width || 1024;
        c.height = currentDimensions.height || 1024;
        const ctx = c.getContext('2d');
        ctx.fillStyle = currentDimensions.bg || '#090a0d';
        ctx.fillRect(0, 0, c.width, c.height);
        ctx.drawImage(img, 0, 0, c.width, c.height);
        URL.revokeObjectURL(svgUrl);

        const a = document.createElement('a');
        a.href = c.toDataURL('image/png');
        a.download = 'art.png';
        a.click();
        showToast('Exported PNG raster');
      };
      img.onerror = () => {
        window.location.href = '/api/export?format=png';
      };
      img.src = svgUrl;
    } else {
      window.location.href = '/api/export?format=png';
      showToast('Exporting PNG raster...');
    }
  });

  btnExportSvg.addEventListener('click', () => {
    const svgCode = latestFullSvg || vectorCanvas.outerHTML;
    if (svgCode && (svgCode.includes('<polygon') || svgCode.includes('<path') || svgCode.includes('<ellipse') || svgCode.includes('<rect'))) {
      const blob = new Blob([svgCode], { type: 'image/svg+xml;charset=utf-8' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'art.svg';
      a.click();
      URL.revokeObjectURL(url);
      showToast('Exported Vector SVG');
    } else {
      window.location.href = '/api/export?format=svg';
      showToast('Exporting Vector SVG...');
    }
  });

  // Animated SVG Timelapse Export
  if (btnExportAnim) {
    btnExportAnim.addEventListener('click', () => {
      if (!shapeHistory || shapeHistory.length === 0) {
        showToast('Synthesize vector art first to export animation', '⚠️');
        return;
      }
      showToast('Generating standalone animated SVG timelapse...');
      const w = currentDimensions.width || 1024;
      const h = currentDimensions.height || 1024;
      const bg = currentDimensions.bg || '#090a0d';

      const totalDuration = Math.min(8.0, Math.max(2.5, shapeHistory.length * 0.025));
      const stepDelay = totalDuration / shapeHistory.length;

      let animSvg = `<?xml version="1.0" encoding="UTF-8"?>\n`;
      animSvg += `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${w} ${h}" width="${w}" height="${h}">\n`;
      animSvg += `<style>\n`;
      animSvg += `  @keyframes artFadeIn {\n`;
      animSvg += `    0% { opacity: 0; transform: scale(0.98); }\n`;
      animSvg += `    100% { opacity: 1; transform: scale(1); }\n`;
      animSvg += `  }\n`;
      animSvg += `  .art-shape {\n`;
      animSvg += `    animation: artFadeIn 0.25s cubic-bezier(0.16, 1, 0.3, 1) forwards;\n`;
      animSvg += `    opacity: 0;\n`;
      animSvg += `    transform-origin: center;\n`;
      animSvg += `  }\n`;
      animSvg += `</style>\n`;
      animSvg += `<rect width="${w}" height="${h}" fill="${bg}"/>\n`;
      animSvg += `<g id="primitives">\n`;

      shapeHistory.forEach((chunk, idx) => {
        const delay = (idx * stepDelay).toFixed(3);
        animSvg += `  <g class="art-shape" style="animation-delay: ${delay}s;">${chunk}</g>\n`;
      });

      animSvg += `</g>\n</svg>`;

      const blob = new Blob([animSvg], { type: 'image/svg+xml;charset=utf-8' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `art-timelapse-${Date.now()}.svg`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      showToast('Animated SVG downloaded!');
    });
  }

  // ========================================================================
  // Ultra-HD 4K & Museum Print Master Export Controller
  // ========================================================================

  let selectedExportSize = 3840;
  let selectedExportTexture = 'none';
  let selectedExportTitle = '4K UHD Master';

  function updateExportSpecs() {
    const w = currentDimensions.width || 1024;
    const h = currentDimensions.height || 1024;
    const aspect = w / h;
    let tw, th;
    if (aspect >= 1.0) {
      tw = selectedExportSize;
      th = Math.round(selectedExportSize / aspect);
    } else {
      tw = Math.round(selectedExportSize * aspect);
      th = selectedExportSize;
    }

    if (exportGalleryMat && exportGalleryMat.checked) {
      const margin = Math.max(24, Math.round(tw * 0.04));
      tw += margin * 2;
      th += margin * 2;
    }

    const inW = (tw / 300).toFixed(1);
    const inH = (th / 300).toFixed(1);
    const cmW = (inW * 2.54).toFixed(1);
    const cmH = (inH * 2.54).toFixed(1);

    if (exportSpecDims) {
      exportSpecDims.textContent = `${tw} × ${th} px`;
    }
    if (exportSpecPhysical) {
      exportSpecPhysical.textContent = `${inW} × ${inH} in · ${cmW} × ${cmH} cm`;
    }
    if (exportSubmitText) {
      exportSubmitText.textContent = `Download ${selectedExportTitle} (${tw}×${th}px)`;
    }
  }

  if (btnExport4k && exportModal) {
    btnExport4k.addEventListener('click', () => {
      if (!latestFullSvg && shapeHistory.length === 0) {
        showToast('Run synthesis first to export print master', '⚠️');
        return;
      }
      updateExportSpecs();
      exportModal.showModal();
    });

    if (btnCloseExportModal) {
      btnCloseExportModal.addEventListener('click', () => {
        exportModal.close();
      });
    }

    exportModal.addEventListener('click', (e) => {
      if (e.target === exportModal) {
        exportModal.close();
      }
    });

    exportResCards.forEach(card => {
      card.addEventListener('click', () => {
        exportResCards.forEach(c => c.classList.remove('active'));
        card.classList.add('active');
        selectedExportSize = parseInt(card.getAttribute('data-size')) || 3840;
        const titleEl = card.querySelector('.export-opt-title');
        if (titleEl) {
          selectedExportTitle = titleEl.textContent;
        }
        updateExportSpecs();
      });
    });

    exportTexturePills.forEach(pill => {
      pill.addEventListener('click', () => {
        exportTexturePills.forEach(p => p.classList.remove('active'));
        pill.classList.add('active');
        selectedExportTexture = pill.getAttribute('data-texture') || 'none';
      });
    });

    if (exportGalleryMat) {
      exportGalleryMat.addEventListener('change', updateExportSpecs);
    }

    if (btnDoExportMaster) {
      btnDoExportMaster.addEventListener('click', () => {
        exportModal.close();
        const isMat = exportGalleryMat ? exportGalleryMat.checked : false;
        showToast(`Rendering ${selectedExportTitle}... Generating print master.`);
        window.location.href = `/api/export?format=png&size=${selectedExportSize}&texture=${selectedExportTexture}&mat=${isMat}`;
      });
    }
  }

  // ========================================================================
  // Detail Focus Brush Controller (Mask Painting for Curtains / Features)
  // ========================================================================

  if (brushSize && brushSizeVal) {
    brushSize.addEventListener('input', (e) => {
      currentBrushRadius = parseInt(e.target.value) || 35;
      brushSizeVal.textContent = `${currentBrushRadius}px`;
    });
  }

  if (btnToggleBrush && brushToolbar && maskCanvas) {
    btnToggleBrush.addEventListener('click', () => {
      isBrushActive = !isBrushActive;
      btnToggleBrush.classList.toggle('active', isBrushActive);
      brushToolbar.classList.toggle('hidden', !isBrushActive);
      canvasWrapper.classList.toggle('brush-active', isBrushActive);
      if (isBrushActive) {
        showToast('Focus Brush: Paint over curtains or subjects to lock details');
      }
    });

    if (btnClearMask) {
      btnClearMask.addEventListener('click', () => {
        if (maskCtx) {
          maskCtx.clearRect(0, 0, maskCanvas.width, maskCanvas.height);
        }
        hasUserMask = false;
        updateRefineButtonLabel();
        showToast('Priority detail mask cleared');
      });
    }

    function getMaskCoords(e) {
      const rect = maskCanvas.getBoundingClientRect();
      const clientX = e.clientX ?? (e.touches && e.touches[0] ? e.touches[0].clientX : 0);
      const clientY = e.clientY ?? (e.touches && e.touches[0] ? e.touches[0].clientY : 0);
      const scaleX = maskCanvas.width / (rect.width || 1);
      const scaleY = maskCanvas.height / (rect.height || 1);
      return {
        x: (clientX - rect.left) * scaleX,
        y: (clientY - rect.top) * scaleY
      };
    }

    function drawMaskStroke(x, y, fromX, fromY) {
      if (!maskCtx) return;
      maskCtx.save();
      maskCtx.lineCap = 'round';
      maskCtx.lineJoin = 'round';
      // Coral-amber glowing highlight for user visibility
      maskCtx.strokeStyle = 'rgba(234, 67, 53, 0.75)';
      maskCtx.fillStyle = 'rgba(234, 67, 53, 0.75)';

      const rect = maskCanvas.getBoundingClientRect();
      const scale = maskCanvas.width / (rect.width || 1);
      const r = currentBrushRadius * scale;

      maskCtx.beginPath();
      if (fromX !== undefined && fromY !== undefined) {
        maskCtx.lineWidth = r * 2;
        maskCtx.moveTo(fromX, fromY);
        maskCtx.lineTo(x, y);
        maskCtx.stroke();
      } else {
        maskCtx.arc(x, y, r, 0, Math.PI * 2);
        maskCtx.fill();
      }
      maskCtx.restore();
      hasUserMask = true;
      updateRefineButtonLabel();
    }

    maskCanvas.addEventListener('pointerdown', (e) => {
      if (!isBrushActive) return;
      e.preventDefault();
      isDrawingMask = true;
      const coords = getMaskCoords(e);
      lastBrushX = coords.x;
      lastBrushY = coords.y;
      drawMaskStroke(coords.x, coords.y);
    });

    window.addEventListener('pointermove', (e) => {
      if (!isBrushActive || !isDrawingMask) return;
      const coords = getMaskCoords(e);
      drawMaskStroke(coords.x, coords.y, lastBrushX, lastBrushY);
      lastBrushX = coords.x;
      lastBrushY = coords.y;
    });

    const stopDrawingMask = () => {
      isDrawingMask = false;
    };
    window.addEventListener('pointerup', stopDrawingMask);
    window.addEventListener('pointercancel', stopDrawingMask);
  }

  // ========================================================================
  // Blueprint / Wireframe Mode Controller
  // ========================================================================

  if (btnToggleWireframe) {
    btnToggleWireframe.addEventListener('click', () => {
      const isWire = vectorCanvas.classList.toggle('wireframe-mode');
      btnToggleWireframe.classList.toggle('active', isWire);
      showToast(isWire ? 'Blueprint wireframe mode active' : 'Wireframe mode deactivated');
    });
  }

  // ========================================================================
  // Live Hardware-Accelerated Tactile Texture Controller
  // ========================================================================

  if (btnToggleTexture) {
    btnToggleTexture.addEventListener('click', () => {
      if (currentTextureMode === 'none') {
        currentTextureMode = 'paper';
        vectorCanvas.classList.remove('texture-canvas');
        vectorCanvas.classList.add('texture-paper');
        btnToggleTexture.classList.add('active');
        showToast('Tactile Texture: Museum Cotton Rag Paper (100% Cotton, 310gsm)');
      } else if (currentTextureMode === 'paper') {
        currentTextureMode = 'canvas';
        vectorCanvas.classList.remove('texture-paper');
        vectorCanvas.classList.add('texture-canvas');
        btnToggleTexture.classList.add('active');
        showToast('Tactile Texture: Belgian Linen Canvas (Coarse Impasto Weave)');
      } else {
        currentTextureMode = 'none';
        vectorCanvas.classList.remove('texture-paper', 'texture-canvas');
        btnToggleTexture.classList.remove('active');
        showToast('Tactile Texture: Pure Clean Vector (No Texture)');
      }
    });
  }

  // ========================================================================
  // Timeline Scrubber & Step Replay Engine
  // ========================================================================

  function renderTimelineFrame(frameIndex) {
    const count = Math.max(0, Math.min(shapeHistory.length, frameIndex));
    vectorCanvas.innerHTML = baseBgRect + shapeHistory.slice(0, count).join('');
    if (timelineStepText) {
      timelineStepText.textContent = `${count} / ${shapeHistory.length}`;
    }
    metricShapes.textContent = `${count} / ${shapeHistory.length}`;
  }

  function stopTimelinePlay() {
    if (timelineInterval) {
      clearInterval(timelineInterval);
      timelineInterval = null;
    }
    isTimelinePlaying = false;
    if (btnTimelinePlay) {
      btnTimelinePlay.innerHTML = `<svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>`;
    }
  }

  function startTimelinePlay() {
    if (shapeHistory.length === 0) return;
    isTimelinePlaying = true;
    if (btnTimelinePlay) {
      btnTimelinePlay.innerHTML = `<svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>`;
    }

    let current = parseInt(timelineSlider.value);
    if (current >= shapeHistory.length) {
      current = 0;
    }

    timelineInterval = setInterval(() => {
      current += 1;
      if (current > shapeHistory.length) {
        stopTimelinePlay();
        return;
      }
      timelineSlider.value = current;
      renderTimelineFrame(current);
    }, 28);
  }

  if (timelineSlider) {
    timelineSlider.addEventListener('input', (e) => {
      stopTimelinePlay();
      renderTimelineFrame(parseInt(e.target.value));
    });
  }

  if (btnTimelinePlay) {
    btnTimelinePlay.addEventListener('click', () => {
      if (isTimelinePlaying) {
        stopTimelinePlay();
      } else {
        startTimelinePlay();
      }
    });
  }

  // ========================================================================
  // Keyboard Shortcuts & Modal Dialog
  // ========================================================================

  btnShortcutsModal.addEventListener('click', () => {
    shortcutsModal.showModal();
  });

  btnCloseModal.addEventListener('click', () => {
    shortcutsModal.close();
  });

  shortcutsModal.addEventListener('click', (e) => {
    if (e.target === shortcutsModal) {
      shortcutsModal.close();
    }
  });

  document.addEventListener('keydown', (e) => {
    // Avoid triggering when user focuses an input
    if (['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName)) return;

    if (e.key === '?' || (e.key === '/' && e.shiftKey)) {
      e.preventDefault();
      shortcutsModal.open ? shortcutsModal.close() : shortcutsModal.showModal();
    } else if (e.key === 'Escape' && shortcutsModal.open) {
      shortcutsModal.close();
    } else if (e.key === 'Enter' && e.shiftKey) {
      e.preventDefault();
      if (!isRunning && shapeHistory.length > 0 && btnRefine) {
        btnRefine.click();
      }
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (isRunning) {
        btnStop.click();
      } else {
        btnStart.click();
      }
    } else if (e.code === 'Space') {
      e.preventDefault();
      // Toggle between split and vector only
      if (activeViewMode === 'split') {
        setActiveViewSegment('vector');
      } else {
        setActiveViewSegment('split');
      }
    } else if (e.key === 'ArrowLeft') {
      setWipePosition(wipeRatio - 0.05);
    } else if (e.key === 'ArrowRight') {
      setWipePosition(wipeRatio + 0.05);
    } else if (e.key === 'c' || e.key === 'C') {
      btnCopySvg.click();
    } else if (e.key === 'r' || e.key === 'R') {
      btnLoadSample.click();
    } else if (e.key === 'b' || e.key === 'B') {
      if (btnToggleBrush) btnToggleBrush.click();
    } else if (e.key === 'w' || e.key === 'W') {
      if (btnToggleWireframe) btnToggleWireframe.click();
    } else if (e.key === 'h' || e.key === 'H') {
      if (activeViewMode === 'heatmap') {
        setActiveViewSegment('split');
      } else {
        setActiveViewSegment('heatmap');
      }
    } else if (e.key === 't' || e.key === 'T') {
      if (btnToggleTexture) btnToggleTexture.click();
    } else if (e.key === '+' || e.key === '=') {
      btnZoomIn.click();
    } else if (e.key === '-' || e.key === '_') {
      btnZoomOut.click();
    } else if (e.key === '0') {
      btnZoomFit.click();
    } else if (['1', '2', '3', '4', '5', '6', '7', '8', '9', '0'].includes(e.key)) {
      const modeMap = {
        '1': 1, '2': 6, '3': 9, '4': 2, '5': 3,
        '6': 4, '7': 5, '8': 7, '9': 8, '0': 0
      };
      if (modeMap[e.key] !== undefined) {
        selectMode(modeMap[e.key]);
      }
    }
  });

  // ========================================================================
  // Theme Management (Dark Mode Default + Toggle)
  // ========================================================================

  const btnThemeToggle = document.getElementById('btn-theme-toggle');
  const iconSun = btnThemeToggle ? btnThemeToggle.querySelector('.icon-sun') : null;
  const iconMoon = btnThemeToggle ? btnThemeToggle.querySelector('.icon-moon') : null;

  function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('art-studio-theme', theme);
    if (iconSun && iconMoon) {
      if (theme === 'dark') {
        iconSun.classList.remove('hidden');
        iconMoon.classList.add('hidden');
      } else {
        iconSun.classList.add('hidden');
        iconMoon.classList.remove('hidden');
      }
    }
  }

  // Load saved theme or default to dark mode
  const savedTheme = localStorage.getItem('art-studio-theme') || 'dark';
  applyTheme(savedTheme);

  if (btnThemeToggle) {
    btnThemeToggle.addEventListener('click', () => {
      const current = document.documentElement.getAttribute('data-theme') || 'dark';
      const next = current === 'dark' ? 'light' : 'dark';
      applyTheme(next);
      showToast(`Switched to ${next} mode`);
    });
  }

  // Automatically load sample reference on boot
  btnLoadSample.click();
});
