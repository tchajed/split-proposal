# Split Proposal Web Interface

A SvelteKit-based web interface for splitting NSF proposal PDFs into submission documents.

## Features

- Drag-and-drop PDF upload
- Browser-based PDF processing using WebAssembly
- Automatically determine where to split based on PDF bookmarks.

## Development

Run `./scripts/build.sh` to build the WASM module and copy `wasm_exec.js` from the Go distribution.

```bash
# Install dependencies
bun install

# Start dev server
bun run dev

# Build for production
bun run build

# Preview production build
bun run preview
```

## How it works

The Go split-proposal tool is compiled to WebAssembly, and the main splitting function is exposed as a JavaScript function. This requires a small JavaScript runtime provided by Go (`wasm_exec.js`). This means all processing is done in the browser, and the server just serves static files.

## Testing

You can test the application with the sample PDF:

- Build the WASM code with `./scripts/build.sh`
- Run the dev server with `bun run dev`
- Navigate to <http://localhost:5173>
- Drag and drop `../sample/main.pdf` onto the drop zone
