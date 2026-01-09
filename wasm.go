//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func init() {
	// Disable pdfcpu config directory to prevent filesystem access in WASM
	api.DisableConfigDir()
}

func splitPdfWasmWrapper(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		obj := js.Global().Get("Object").New()
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
		obj := js.Global().Get("Object").New()
		obj.Set("error", err.Error())
		return obj
	}

	// Convert results to JavaScript array
	jsResults := js.Global().Get("Array").New(len(results))
	for i, result := range results {
		// Create a Uint8Array in JavaScript
		jsArray := js.Global().Get("Uint8Array").New(len(result.Data))
		js.CopyBytesToJS(jsArray, result.Data)

		// Create result object
		resultObj := js.Global().Get("Object").New()
		resultObj.Set("name", result.Name)
		resultObj.Set("startPage", result.StartPage)
		resultObj.Set("endPage", result.EndPage)
		resultObj.Set("data", jsArray)

		jsResults.SetIndex(i, resultObj)
	}

	// Return object with results
	returnObj := js.Global().Get("Object").New()
	returnObj.Set("results", jsResults)
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
