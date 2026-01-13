<script lang="ts">
	import { onMount } from 'svelte';
	import { format } from 'date-fns';
	import { loadWasm, splitPdf } from '$lib/wasm';
	import uploadIcon from '$lib/assets/upload-icon.svg';
	import downloadIcon from '$lib/assets/download-icon.svg';
	import favicon from '$lib/assets/logo.png';

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
			name = `${name} ${format(fileModifiedDate, 'yyyy-MM-dd HH_mm')}`;
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

<main class="mx-auto max-w-3xl p-8">
	<h1 class="mb-2 text-4xl text-gray-800">
		<div class="flex items-center gap-4">
			<img src={favicon} alt="Split NSF Proposal Logo" width="70" height="70" />
			Split NSF Proposal
		</div>
	</h1>
	<p class="mb-8 text-xl text-gray-500">Split a proposal PDF into submission documents</p>

	{#if error}
		<div class="mt-4 rounded-md border border-red-400 bg-red-50 p-4 text-red-800">
			{error}
		</div>
	{/if}

	{#if splitResults}
		<div class="mt-4 rounded-md border border-green-400 bg-green-50 p-4">
			<div class="mb-4">
				<div class="mb-2 flex items-center gap-2">
					<label for="zip-name" class="font-medium">File name:</label>
					<input
						type="text"
						id="zip-name"
						bind:value={zipBaseName}
						class="field-sizing-content min-w-20 rounded bg-slate-300/30 px-2.5 py-1.5 text-base"
					/>
					<span class="text-gray-500">.zip</span>
				</div>
				<label class="flex cursor-pointer items-center gap-2">
					<input type="checkbox" bind:checked={includeDateTime} class="size-4 cursor-pointer" />
					Add timestamp
				</label>
			</div>
			<a href={splitResults.zipUrl} download={getZipFileName()} class="btn-success">
				Download {getZipFileName()}
			</a>
			<ul class="my-2 list-none pl-4">
				{#each splitResults.downloads as download}
					<li class="my-2 flex items-center gap-2">
						<a
							href={download.url}
							download={download.name}
							title="Download"
							class="flex size-7 items-center justify-center rounded
								bg-blue-600 transition-colors hover:bg-blue-900"
						>
							<img src={downloadIcon} alt="Download" class="size-4 filter-invert" />
						</a>
						<a href={download.url} class="font-medium text-blue-700 no-underline hover:underline">
							{download.name}
						</a>
						<span class="text-sm text-gray-500">
							(page{download.startPage == download.endPage ? '' : 's'}
							{prettyRange(download.startPage, download.endPage)})
						</span>
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
			<div class="text-center">
				<div
					class="mx-auto mb-4 size-12 animate-spin rounded-full
						border-4 border-gray-200 border-t-blue-500"
				></div>
				<p>Processing PDF...</p>
			</div>
		{:else}
			<div class="w-full">
				<img src={uploadIcon} alt="Upload" class="mb-4 inline-block size-16 opacity-50" />
				<p class="mb-2 text-xl text-gray-800">Drag and drop your PDF here</p>
				<p class="mb-4 text-gray-400">
					or &nbsp;
					<label>
						<input
							type="file"
							accept=".pdf"
							onchange={handleFileInput}
							disabled={!wasmReady || processing}
							class="hidden"
						/>
						<span class="btn-primary">Choose File</span>
					</label>
				</p>
			</div>
		{/if}
	</div>

	<div class="mt-8 rounded-lg border border-gray-200 bg-white p-6">
		<p class="info-text">
			This tool splits an NSF proposal PDF into separate submission documents for the summary,
			description, and references. <span class="font-bold">This runs entirely locally.</span> Your PDF
			is never shared with the server.
		</p>
		<p class="info-text mt-6">
			For best results, add <code class="code-inline">\pdfbookmark</code> commands to your LaTeX source,
			which are used to identify section page ranges.
		</p>
		<ul class="list-disc pl-5 leading-loose text-gray-500">
			<li>
				<code class="code-inline"
					>\pdfbookmark[0]&#123;Project Description&#125;&#123;Project Description&#125;</code
				> before the project description
			</li>
			<li>
				<code class="code-inline"
					>\pdfbookmark[0]&#123;References cited&#125;&#123;References cited&#125;</code
				> before references
			</li>
			<li>
				(optional)
				<code class="code-inline"
					>\pdfbookmark[0]&#123;Data management plan&#125;&#123;Data management plan&#125;</code
				>
			</li>
			<li>
				(optional)
				<code class="code-inline"
					>\pdfbookmark[0]&#123;Mentoring plan&#125;&#123;Mentoring plan&#125;</code
				>
			</li>
		</ul>
		<p class="info-text">
			Without bookmarks, your project description is assumed to be 15 pages. See this
			<a href="https://github.com/tchajed/split-proposal/blob/main/sample/main.tex">
				sample file
			</a>
			for a complete example.
		</p>
		<div class="my-4 border-b-2 border-gray-200"></div>
		<p class="info-text">You can also use this from the command line:</p>
		<pre
			class="my-2 rounded border border-gray-200 bg-gray-100 p-2 font-mono text-sm">go run github.com/tchajed/split-proposal@latest -file main.pdf</pre>
		<p class="info-text">
			For more information, visit the
			<a href="https://github.com/tchajed/split-proposal">GitHub repository</a>.
		</p>
	</div>
</main>
