// ART Studio — WebAssembly Synthesis Worker
// Runs Go Primitive Engine inside a background Web Worker
importScripts('wasm_exec.js');

const go = new Go();
let isWasmLoaded = false;
let isRunning = false;

async function initWasm() {
  try {
    const wasmResponse = await fetch('primitive.wasm');
    const wasmBytes = await wasmResponse.arrayBuffer();
    const wasmModule = await WebAssembly.instantiate(wasmBytes, go.importObject);
    go.run(wasmModule.instance);
    isWasmLoaded = true;
    postMessage({ type: 'wasm_ready' });
  } catch (err) {
    console.warn('Web Worker WebAssembly load notice:', err);
    postMessage({ type: 'wasm_error', error: err.message });
  }
}

initWasm();

self.onmessage = async (e) => {
  const { action, payload } = e.data;

  if (action === 'start') {
    if (!isWasmLoaded) {
      postMessage({ type: 'wasm_not_ready' });
      return;
    }

    try {
      const initRes = self.primitiveInit(
        payload.imageData || payload.image,
        payload.mode || 1,
        payload.count || 100,
        payload.alpha || 128,
        payload.repeat || 1,
        payload.resolution || 256,
        payload.pyramid !== false
      );

      if (initRes && initRes.error) {
        postMessage({ type: 'error', error: initRes.error });
        return;
      }

      postMessage({ type: 'init', ...initRes });

      isRunning = true;
      while (isRunning) {
        const stepRes = self.primitiveStep();
        if (stepRes.done) {
          postMessage({ type: 'done', ...stepRes });
          isRunning = false;
          break;
        }

        postMessage({ type: 'shape', ...stepRes });

        // Yield execution to allow Stop commands
        await new Promise((r) => setTimeout(r, 0));
      }
    } catch (err) {
      console.error('Error during WebAssembly synthesis:', err);
      postMessage({ type: 'error', error: err.message });
    }
  } else if (action === 'stop') {
    isRunning = false;
    if (typeof self.primitiveStop === 'function') {
      self.primitiveStop();
    }
    postMessage({ type: 'cancelled' });
  } else if (action === 'getSvg') {
    if (typeof self.primitiveGetSvg === 'function') {
      postMessage({ type: 'svg', svg: self.primitiveGetSvg() });
    }
  }
};
