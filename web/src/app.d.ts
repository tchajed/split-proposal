// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}

	interface SplitResult {
		name: string;
		startPage: number;
		endPage: number;
		data: Uint8Array;
	}

	interface Window {
		Go: new () => {
			importObject: WebAssembly.Imports;
			run: (instance: WebAssembly.Instance) => void;
		};
		splitPdf: (data: Uint8Array) => {
			error?: string;
			results?: Array<SplitResult>;
		};
		wasmReady: () => void;
	}
}

export {};
