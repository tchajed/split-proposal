GO_SRC := $(shell find . -name '*.go') go.mod go.sum

web/static/split-proposal.wasm: $(GO_SRC)
	GOOS=js GOARCH=wasm go build -o web/static/split-proposal.wasm
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/
