# Split Proposal Web Interface

A SvelteKit-based web interface for splitting NSF proposal PDFs into submission documents.

## Features

- Drag-and-drop PDF upload
- Browser-based PDF processing using WebAssembly
- Automatically determine where to split based on PDF bookmarks.

## Development

Setup:

```bash
# Install dependencies (only needed once)
bun install
```

You can use these commands:

- `bun dev` start a development server
- `bun wasm` re-compile the wasm module
- `bun run build` build the site for production

There are some automated browser tests using Playwright. Run `bunx playwright install` and then `bun run test`. You can use `bun run test:ui` to also see screenshots of each test run.

## How it works

The Go split-proposal tool is compiled to WebAssembly (`GOARCH=wasm`), and the main splitting function is exposed as a JavaScript function. This requires a small JavaScript runtime provided by Go (`wasm_exec.js`). This means all processing is done in the browser, and the server just serves static files.

## Testing

You can test the application with the sample PDF:

- Build the WASM code with `bun wasm`
- Run the dev server with `bun run dev`
- Navigate to <http://localhost:5173>
- Drag and drop `../sample/main.pdf` onto the drop zone
