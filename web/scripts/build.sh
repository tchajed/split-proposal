#!/bin/bash

set -eu

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
# repo root (go source)
cd "$DIR/../.."

GOOS=js GOARCH=wasm go build -o web/static/split-proposal.wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/
