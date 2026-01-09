//go:build !wasm

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func splitPdf(rs io.ReadSeeker, outDir string) error {
	pdfData, err := io.ReadAll(rs)
	if err != nil {
		return fmt.Errorf("could not read PDF data: %w", err)
	}

	results, err := splitPdfBytes(pdfData)
	if err != nil {
		return err
	}

	// Write each result to the output directory
	for _, result := range results {
		err = os.WriteFile(filepath.Join(outDir, result.Name), result.Data, 0644)
		if err != nil {
			return fmt.Errorf("could not write output PDF %s: %w", result.Name, err)
		}
		pageRange := fmt.Sprintf("%d-%d", result.StartPage, result.EndPage)
		if result.EndPage < 0 {
			pageRange = fmt.Sprintf("%d-end", result.StartPage)
		}
		fmt.Printf("%s: %s\n", result.Name, pageRange)
	}

	return nil
}

func main() {
	fileName := flag.String("file", "main.pdf", "proposal PDF file")
	outDir := flag.String("outDir", ".", "directory to write output PDFs")
	flag.Parse()

	f, err := os.Open(*fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	_, err = os.Stat(*outDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "output directory does not exist: %v\n", err)
		os.Exit(1)
	}

	if err := splitPdf(f, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "error splitting: %v\n", err)
		os.Exit(1)
	}
}
