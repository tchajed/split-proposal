//go:build js && wasm

package main

import (
	"archive/zip"
	"bytes"
	"syscall/js"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func init() {
	// Disable pdfcpu config directory to prevent filesystem access in WASM
	api.DisableConfigDir()
}

// resultsToZipFile creates a zip file containing the split PDFs.
func resultsToZipFile(results []SplitResult) []byte {
	var out bytes.Buffer
	zipFile := zip.NewWriter(&out)
	for _, result := range results {
		file, err := zipFile.Create(result.Name)
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
	if len(args) != 1 {
		obj := newObject()
		obj.Set("error", "expected 1 argument (Uint8Array)")
		return obj
	}

	// Get the Uint8Array from JavaScript
	inputArray := args[0]
	length := inputArray.Get("length").Int()
	pdfData := make([]byte, length)
	js.CopyBytesToGo(pdfData, inputArray)

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
	zipFile := bytesToJs(resultsToZipFile(results))

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
