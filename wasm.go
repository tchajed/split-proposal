//go:build js && wasm

package main

import (
	"archive/zip"
	"bytes"
	"path"
	"syscall/js"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func init() {
	// Disable pdfcpu config directory to prevent filesystem access in WASM
	api.DisableConfigDir()
}

// resultsToZipFile creates a zip file containing the split PDFs.
//
// The zip file has a single directory outerDir that contains all PDFs.
func resultsToZipFile(outerDir string, results []SplitResult) []byte {
	var out bytes.Buffer
	zipFile := zip.NewWriter(&out)
	for _, result := range results {
		file, err := zipFile.Create(path.Join(outerDir, result.Name))
		if err != nil {
			panic(err)
		}
		file.Write(result.Data)
	}
	zipFile.Close()
	return out.Bytes()
}

func newObject() js.Value {
	return js.Global().Get("Object").New()
}

func bytesToJs(data []byte) js.Value {
	jsArray := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(jsArray, data)
	return jsArray
}

func splitPdfWasmWrapper(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		obj := newObject()
		obj.Set("error", "expected 2 arguments (Uint8Array, string)")
		return obj
	}

	// Get the Uint8Array from JavaScript
	inputArray := args[0]
	length := inputArray.Get("length").Int()
	pdfData := make([]byte, length)
	js.CopyBytesToGo(pdfData, inputArray)

	// Get the zipName from JavaScript
	zipName := args[1].String()

	// Split the PDF
	results, err := splitPdfBytes(pdfData)
	if err != nil {
		obj := newObject()
		obj.Set("error", err.Error())
		return obj
	}

	// Convert results to JavaScript array
	jsResults := js.Global().Get("Array").New(len(results))
	for i, result := range results {
		// Create result object
		resultObj := newObject()
		resultObj.Set("name", result.Name)
		resultObj.Set("startPage", result.StartPage)
		resultObj.Set("endPage", result.EndPage)
		resultObj.Set("data", bytesToJs(result.Data))

		jsResults.SetIndex(i, resultObj)
	}

	// Create the zip file
	zipFile := bytesToJs(resultsToZipFile(zipName, results))

	// Return object with results
	returnObj := newObject()
	returnObj.Set("results", jsResults)
	returnObj.Set("zipFile", zipFile)
	return returnObj
}

func main() {
	// Set up a channel to keep the program running
	done := make(chan struct{})

	// Register the splitPdf function
	js.Global().Set("splitPdf", js.FuncOf(splitPdfWasmWrapper))

	// Signal that wasm is ready
	js.Global().Call("wasmReady")

	<-done
}
