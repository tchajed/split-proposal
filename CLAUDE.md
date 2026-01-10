split-proposal splits an NSF proposal PDF into submission PDFs.

## Go CLI

The main project is a Go CLI that implements the core splitting functionality using pdfcpu.

Run the Go tests with `go test ./...`. Run the CLI on a sample proposal with `go run . -file sample/main.pdf -outDir sample`.

## Web interface

There is also a web frontend. It uses a wasm interface to the Go code in `wasm.go`. The frontend code is at `web/`: it uses Svelte, bun, and TypeScript.

Check the web frontend code after making any changes with `bun --cwd=web check` and run `bun --cwd=web format` to format the code with prettier.

After making any changes to Go code, compile wasm code with `./web/scripts/build.sh`.
