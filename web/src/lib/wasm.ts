/**
 * WASM module loader for split-proposal
 *
 * Handles loading the Go WASM runtime and the split-proposal WASM module.
 */

let wasmLoaded = false;
let wasmLoadPromise: Promise<void> | null = null;

/**
 * Load the WASM module. Returns a promise that resolves when the module is ready.
 * Safe to call multiple times - subsequent calls return the same promise.
 */
export function loadWasm(): Promise<void> {
	if (wasmLoadPromise) {
		return wasmLoadPromise;
	}

	wasmLoadPromise = doLoadWasm();
	return wasmLoadPromise;
}

async function doLoadWasm(): Promise<void> {
	// Load wasm_exec.js
	const script = document.createElement('script');
	script.src = '/wasm_exec.js';

	await new Promise<void>((resolve, reject) => {
		script.onload = () => resolve();
		script.onerror = () => reject(new Error('Failed to load wasm_exec.js'));
		document.head.appendChild(script);
	});

	// Initialize the Go WASM runtime
	const go = new window.Go();

	// Set up the ready callback
	const readyPromise = new Promise<void>((resolve) => {
		window.wasmReady = () => {
			wasmLoaded = true;
			console.log('WASM module loaded successfully');
			resolve();
		};
	});

	// Load and instantiate the WASM module
	const result = await WebAssembly.instantiateStreaming(
		fetch('/split-proposal.wasm'),
		go.importObject
	);

	// Start the Go runtime
	go.run(result.instance);

	// Wait for the module to signal it's ready
	await readyPromise;
}

/**
 * Check if the WASM module is loaded and ready.
 */
export function isWasmReady(): boolean {
	return wasmLoaded;
}

/**
 * Split a PDF file into its component parts.
 * Throws if the WASM module is not loaded.
 */
export function splitPdf(data: Uint8Array): SplitResult[] {
	if (!wasmLoaded) {
		throw new Error('WASM module not loaded. Call loadWasm() first.');
	}

	const result = window.splitPdf(data);

	if (result.error) {
		throw new Error(result.error);
	}

	return result.results ?? [];
}
