<script lang="ts">
	import { onMount } from 'svelte';
	import { format } from 'date-fns';
	import { loadWasm, splitPdf, type SplitOutput } from '$lib/wasm';
	import uploadIcon from '$lib/assets/upload-icon.svg';

	interface DownloadItem {
		name: string;
		url: string;
		startPage: number;
		endPage: number;
	}

	interface SplitResults {
		downloads: DownloadItem[];
		zipUrl: string;
	}

	let wasmReady = $state(false);
	let processing = $state(false);
	let error = $state<string | null>(null);
	let isDragging = $state(false);
	let splitResults = $state<SplitResults | null>(null);
	let zipBaseName = $state('');
	let includeDateTime = $state(false);
	let fileModifiedDate = $state<Date | null>(null);

	function getZipFileName(): string {
		let name = zipBaseName || 'download';
		if (includeDateTime && fileModifiedDate) {
			name = `${name} ${format(fileModifiedDate, 'yyyy-MM-dd HH:mm')}`;
		}
		return `${name}.zip`;
	}

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

	function cleanupUrls() {
		if (splitResults) {
			for (const download of splitResults.downloads) {
				URL.revokeObjectURL(download.url);
			}
			URL.revokeObjectURL(splitResults.zipUrl);
			splitResults = null;
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
		cleanupUrls();

		try {
			// Read file as ArrayBuffer
			const arrayBuffer = await file.arrayBuffer();
			const uint8Array = new Uint8Array(arrayBuffer);

			// Derive zip name from input filename
			const baseName = file.name.replace(/\.pdf$/i, '');
			zipBaseName = baseName;
			fileModifiedDate = new Date(file.lastModified);

			// Call the WASM function
			const output = splitPdf(uint8Array);

			// Create blob URLs for results
			splitResults = {
				downloads: output.results.map((result) => ({
					name: result.name,
					url: URL.createObjectURL(
						new Blob([result.data as BlobPart], { type: 'application/pdf' })
					),
					startPage: result.startPage,
					endPage: result.endPage
				})),
				zipUrl: URL.createObjectURL(
					new Blob([output.zipFile as BlobPart], { type: 'application/zip' })
				)
			};
		} catch (err) {
			console.error('Error processing file:', err);
			error = 'Error processing file: ' + (err instanceof Error ? err.message : String(err));
		} finally {
			processing = false;
		}
	}
</script>

<svelte:head>
	<title>Split NSF proposal</title>
</svelte:head>

<main>
	<h1>Split NSF Proposal</h1>
	<p class="subtitle">Split a proposal PDF into submission documents</p>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if splitResults}
		<div class="success">
			<div class="zip-options">
				<div class="zip-filename-row">
					<label for="zip-name">download name:</label>
					<input type="text" id="zip-name" bind:value={zipBaseName} class="zip-name-input" />
					<span class="zip-ext">.zip</span>
				</div>
				<label class="checkbox-label">
					<input type="checkbox" bind:checked={includeDateTime} />
					Add timestamp to filename
				</label>
			</div>
			<a href={splitResults.zipUrl} download={getZipFileName()} class="button zip-button">
				Download all ({getZipFileName()})
			</a>
			<ul class="download-list">
				<li></li>
				{#each splitResults.downloads as download}
					<li>
						<a href={download.url} download={download.name} class="download-link">
							{download.name}
						</a>
						<span class="page-info"
							>(page{download.startPage == download.endPage ? '' : 's'}
							{prettyRange(download.startPage, download.endPage)})</span
						>
					</li>
				{/each}
			</ul>
		</div>
	{/if}

	<div
		class="drop-zone"
		class:dragging={isDragging}
		class:disabled={!wasmReady}
		ondragover={wasmReady ? handleDragOver : undefined}
		ondragleave={wasmReady ? handleDragLeave : undefined}
		ondrop={wasmReady ? handleDrop : undefined}
		role="button"
		tabindex={wasmReady ? 0 : -1}
	>
		{#if processing}
			<div class="processing">
				<div class="spinner"></div>
				<p>Processing PDF...</p>
			</div>
		{:else}
			<div class="drop-content">
				<img src={uploadIcon} alt="Upload" class="upload-icon" style="inline-block" />
				<p class="main-text">Drag and drop your PDF here</p>
				<p class="sub-text">
					or
					<label class="file-label">
						<input
							type="file"
							accept=".pdf"
							onchange={handleFileInput}
							disabled={!wasmReady || processing}
						/>
						<span class="button">Choose File</span>
					</label>
				</p>
			</div>
		{/if}
	</div>

	<div class="info">
		<p>
			This tool splits an NSF proposal PDF into separate submission documents: Project Summary,
			Project Description, References.
		</p>
		<p>
			For best results, add <code>\pdfbookmark</code> commands to your LaTeX source, which the tool uses
			to identify section page ranges:
		</p>
		<ul>
			<li>
				<code>\pdfbookmark[0]&#123;Project Description&#125;&#123;Project Description&#125;</code> before
				the project description
			</li>
			<li>
				<code>\pdfbookmark[0]&#123;References cited&#125;&#123;References cited&#125;</code> before references
			</li>
			<li>
				(optional) <code
					>\pdfbookmark[0]&#123;Data management plan&#125;&#123;Data management plan&#125;</code
				>
			</li>
			<li>
				(optional) <code>\pdfbookmark[0]&#123;Mentoring plan&#125;&#123;Mentoring plan&#125;</code>
			</li>
		</ul>
		<p>
			See this
			<a href="https://github.com/tchajed/split-proposal/blob/main/sample/main.tex">sample file</a> for
			a complete example.
		</p>
	</div>
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

	.drop-zone {
		border: 3px dashed #ccc;
		margin-top: 1rem;
		border-radius: 12px;
		padding: 3rem;
		text-align: center;
		background: white;
		transition: all 0.3s ease;
		cursor: pointer;
		min-height: 100px;
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

	.drop-zone.disabled {
		opacity: 0.5;
		cursor: not-allowed;
		pointer-events: none;
	}

	.drop-content {
		width: 100%;
	}

	.upload-icon {
		width: 64px;
		height: 64px;
		margin-bottom: 1rem;
		opacity: 0.5;
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
		margin: 0 0.5rem;
		padding: 0.75rem;
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

	.success ul {
		margin: 0.5rem 0;
		padding-left: 1.5rem;
	}

	.download-list {
		list-style: none;
		padding-left: 0;
	}

	.download-list li {
		margin: 0.5rem 0;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.download-link {
		color: #1565c0;
		text-decoration: none;
		font-weight: 500;
	}

	.download-link:hover {
		text-decoration: underline;
	}

	.page-info {
		color: #666;
		font-size: 0.9em;
	}

	.zip-options {
		margin-bottom: 1rem;
	}

	.zip-filename-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.zip-filename-row label {
		font-weight: 500;
	}

	.zip-name-input {
		padding: 0.4rem 0.6rem;
		border: 1px solid #66bb6a;
		border-radius: 4px;
		font-size: 1rem;
		width: 200px;
	}

	.zip-ext {
		color: #666;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}

	.checkbox-label input[type='checkbox'] {
		width: 16px;
		height: 16px;
		cursor: pointer;
	}

	.zip-button {
		background: #2e7d32;
	}

	.zip-button:hover {
		background: #1b5e20;
	}

	.info {
		margin-top: 2rem;
		padding: 1.5rem;
		background: white;
		border-radius: 8px;
		border: 1px solid #e0e0e0;
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
