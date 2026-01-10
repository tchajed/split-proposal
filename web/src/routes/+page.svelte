<script lang="ts">
	import { onMount } from 'svelte';
	import { loadWasm, splitPdf } from '$lib/wasm';

	let wasmReady = $state(false);
	let processing = $state(false);
	let error = $state<string | null>(null);
	let isDragging = $state(false);
	let results = $state<SplitResult[]>([]);

	function prettyRange(start: number, end: number): string {
		if (start === end) {
			return `${start}`;
		}
		if (end < 0) {
			return `${start}-end`;
		}
		return `${start}-${end}`;
	}

	onMount(async () => {
		try {
			await loadWasm();
			wasmReady = true;
		} catch (err) {
			console.error('Failed to load WASM:', err);
			error = 'Failed to load WASM module: ' + (err instanceof Error ? err.message : String(err));
		}
	});

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		isDragging = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
	}

	async function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;

		const files = e.dataTransfer?.files;
		if (files && files.length > 0) {
			await processFile(files[0]);
		}
	}

	async function handleFileInput(e: Event) {
		const target = e.target as HTMLInputElement;
		const files = target.files;
		if (files && files.length > 0) {
			await processFile(files[0]);
		}
	}

	async function processFile(file: File) {
		if (!wasmReady) {
			error = 'WASM module not ready yet. Please wait...';
			return;
		}

		if (!file.name.endsWith('.pdf')) {
			error = 'Please select a PDF file';
			return;
		}

		processing = true;
		error = null;
		results = [];

		try {
			// Read file as ArrayBuffer
			const arrayBuffer = await file.arrayBuffer();
			const uint8Array = new Uint8Array(arrayBuffer);

			// Call the WASM function
			results = splitPdf(uint8Array);

			// Automatically download all files
			for (const splitFile of results) {
				downloadFile(splitFile.name, splitFile.data);
			}
		} catch (err) {
			console.error('Error processing file:', err);
			error = 'Error processing file: ' + (err instanceof Error ? err.message : String(err));
		} finally {
			processing = false;
		}
	}

	function downloadFile(filename: string, uint8Array: Uint8Array) {
		const blob = new Blob([uint8Array as BlobPart], { type: 'application/pdf' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = filename;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}
</script>

<svelte:head>
	<title>Split Proposal - NSF Proposal PDF Splitter</title>
</svelte:head>

<main>
	<h1>Split Proposal</h1>
	<p class="subtitle">Split NSF proposal PDFs into submission documents</p>

	{#if !wasmReady}
		<div class="loading">Loading WASM module...</div>
	{:else}
		<div
			class="drop-zone"
			class:dragging={isDragging}
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
			ondrop={handleDrop}
			role="button"
			tabindex="0"
		>
			{#if processing}
				<div class="processing">
					<div class="spinner"></div>
					<p>Processing PDF...</p>
				</div>
			{:else}
				<div class="drop-content">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="64"
						height="64"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="17 8 12 3 7 8"></polyline>
						<line x1="12" y1="3" x2="12" y2="15"></line>
					</svg>
					<p class="main-text">Drag and drop your PDF here</p>
					<p class="sub-text">or</p>
					<label class="file-label">
						<input type="file" accept=".pdf" onchange={handleFileInput} disabled={processing} />
						<span class="button">Choose File</span>
					</label>
				</div>
			{/if}
		</div>

		{#if error}
			<div class="error">{error}</div>
		{/if}

		{#if results.length > 0}
			<div class="success">
				<h2>Successfully split PDF!</h2>
				<p>Downloaded {results.length} file{results.length > 1 ? 's' : ''}:</p>
				<ul>
					{#each results as result}
						<li>{result.name} (pages {prettyRange(result.startPage, result.endPage)})</li>
					{/each}
				</ul>
			</div>
		{/if}

		<div class="info">
			<h3>How it works</h3>
			<p>
				This tool splits an NSF proposal PDF into separate submission documents. It uses PDF
				bookmarks to identify sections:
			</p>
			<ul>
				<li>Project Summary (default: page 1)</li>
				<li>Project Description (default: pages 2-16)</li>
				<li>References Cited (default: pages 17-end)</li>
				<li>Data Management Plan (if present)</li>
				<li>Mentoring Plan (if present)</li>
			</ul>
			<p>
				For best results, add <code>\pdfbookmark</code> commands to your LaTeX source. See the
				<a href="https://github.com/tchajed/split-proposal" target="_blank">GitHub repository</a>
				for details.
			</p>
		</div>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
		font-family:
			-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
		background: #f5f5f5;
	}

	main {
		max-width: 800px;
		margin: 0 auto;
		padding: 2rem;
	}

	h1 {
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
		color: #333;
	}

	.subtitle {
		font-size: 1.2rem;
		color: #666;
		margin-bottom: 2rem;
	}

	.loading {
		text-align: center;
		padding: 3rem;
		font-size: 1.2rem;
		color: #666;
	}

	.drop-zone {
		border: 3px dashed #ccc;
		border-radius: 12px;
		padding: 3rem;
		text-align: center;
		background: white;
		transition: all 0.3s ease;
		cursor: pointer;
		min-height: 300px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.drop-zone:hover {
		border-color: #4a90e2;
		background: #f8f9fa;
	}

	.drop-zone.dragging {
		border-color: #4a90e2;
		background: #e3f2fd;
		transform: scale(1.02);
	}

	.drop-content {
		width: 100%;
	}

	.drop-content svg {
		color: #999;
		margin-bottom: 1rem;
	}

	.main-text {
		font-size: 1.3rem;
		color: #333;
		margin-bottom: 0.5rem;
	}

	.sub-text {
		color: #999;
		margin-bottom: 1rem;
	}

	.file-label input {
		display: none;
	}

	.button {
		display: inline-block;
		padding: 0.75rem 1.5rem;
		background: #4a90e2;
		color: white;
		border-radius: 6px;
		cursor: pointer;
		font-size: 1rem;
		transition: background 0.2s;
	}

	.button:hover {
		background: #357abd;
	}

	.processing {
		text-align: center;
	}

	.spinner {
		border: 4px solid #f3f3f3;
		border-top: 4px solid #4a90e2;
		border-radius: 50%;
		width: 50px;
		height: 50px;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	.error {
		margin-top: 1rem;
		padding: 1rem;
		background: #ffebee;
		border: 1px solid #ef5350;
		border-radius: 6px;
		color: #c62828;
	}

	.success {
		margin-top: 1rem;
		padding: 1rem;
		background: #e8f5e9;
		border: 1px solid #66bb6a;
		border-radius: 6px;
		color: #2e7d32;
	}

	.success h2 {
		margin-top: 0;
		font-size: 1.3rem;
	}

	.success ul {
		margin: 0.5rem 0;
		padding-left: 1.5rem;
	}

	.info {
		margin-top: 3rem;
		padding: 1.5rem;
		background: white;
		border-radius: 8px;
		border: 1px solid #e0e0e0;
	}

	.info h3 {
		margin-top: 0;
		color: #333;
	}

	.info p {
		color: #666;
		line-height: 1.6;
	}

	.info ul {
		color: #666;
		line-height: 1.8;
	}

	.info code {
		background: #f5f5f5;
		padding: 0.2rem 0.4rem;
		border-radius: 3px;
		font-family: 'Courier New', monospace;
		font-size: 0.9em;
	}

	.info a {
		color: #4a90e2;
		text-decoration: none;
	}

	.info a:hover {
		text-decoration: underline;
	}
</style>
