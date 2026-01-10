# Split Proposal Web Interface

A SvelteKit-based web interface for splitting NSF proposal PDFs into submission documents.

## Features

- Drag-and-drop PDF upload
- Browser-based PDF processing using WebAssembly
- Automatic download of split files
- No server-side processing required

## Development

```bash
# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Rebuilding the WASM module

If you modify the Go code, rebuild the WASM module from the project root:

```bash
cd ..
GOOS=js GOARCH=wasm go build -o web/static/split-proposal.wasm
```

The `wasm_exec.js` file is copied from Go's standard library and is needed to run Go WASM in the browser. If you need to update it for a new Go version:

```bash
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/
```

## How it works

1. The Go split-proposal tool is compiled to WebAssembly
2. The WASM module is loaded in the browser
3. PDF files are processed entirely client-side
4. Split files are automatically downloaded

## Testing

You can test the application with the sample PDF:

- Navigate to http://localhost:5173
- Drag and drop `../sample/main.pdf` onto the drop zone
- The PDF will be split and downloaded automatically
