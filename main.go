package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func subtractFromPage(b []pdfcpu.Bookmark, offset int) {
	for i := range b {
		b[i].PageFrom -= offset
		b[i].PageThru -= offset
		subtractFromPage(b[i].Kids, offset)
	}
}

func bookmarksInRange(bookmarks []pdfcpu.Bookmark, startPageNr int, endPageNr int) []pdfcpu.Bookmark {
	var result []pdfcpu.Bookmark
	for _, bookmark := range bookmarks {
		// Check if bookmark overlaps with the range
		if bookmark.PageThru >= startPageNr && bookmark.PageFrom <= endPageNr {
			// Create a copy of the bookmark
			newBookmark := pdfcpu.Bookmark{
				Title:    bookmark.Title,
				PageFrom: bookmark.PageFrom,
				PageThru: bookmark.PageThru,
				Kids:     bookmark.Kids,
				Bold:     bookmark.Bold,
				Italic:   bookmark.Italic,
				Color:    bookmark.Color,
			}

			// trim to range if overlapping
			if newBookmark.PageFrom < startPageNr {
				newBookmark.PageFrom = startPageNr
			}
			if newBookmark.PageThru > endPageNr {
				newBookmark.PageThru = endPageNr
			}

			// Recursively filter kids
			newBookmark.Kids = bookmarksInRange(bookmark.Kids, startPageNr, endPageNr)
			for i := range newBookmark.Kids {
				newBookmark.Kids[i].Parent = &newBookmark
			}
			result = append(result, newBookmark)
		}
	}
	return result
}

func applyBookmarks(rs io.ReadSeeker, bookmarks []pdfcpu.Bookmark, startPageNr int, endPageNr int) (bytes.Buffer, error) {
	newBookmarks := bookmarksInRange(bookmarks, startPageNr, endPageNr)
	subtractFromPage(newBookmarks, startPageNr-1)
	var buf bytes.Buffer
	if len(newBookmarks) > 0 {
		err := api.AddBookmarks(rs, &buf, newBookmarks, true, model.NewDefaultConfiguration())
		if err != nil {
			return bytes.Buffer{}, fmt.Errorf("could not add bookmarks: %w", err)
		}
	} else {
		b, err := io.ReadAll(rs)
		if err != nil {
			return bytes.Buffer{}, fmt.Errorf("could not read PDF content: %w", err)
		}
		buf = *bytes.NewBuffer(b)
	}
	return buf, nil
}

func pageCount(rs io.ReadSeeker, conf *model.Configuration) (int, error) {
	info, err := api.PDFInfo(rs, "input.pdf", []string{}, false, conf)
	if err != nil {
		return 0, fmt.Errorf("could not read PDF info: %w", err)
	}
	return info.PageCount, nil
}

func extractPages(rs io.ReadSeeker, bookmarks []pdfcpu.Bookmark, r pageRange, conf *model.Configuration) (bytes.Buffer, error) {
	if r.end < 0 {
		count, err := pageCount(rs, conf)
		if err != nil {
			return bytes.Buffer{}, err
		}
		r.end = count
	}
	newBookmarks := bookmarksInRange(bookmarks, r.start, r.end)
	subtractFromPage(newBookmarks, r.start-1)
	var buf bytes.Buffer
	err := api.Trim(rs, &buf, []string{fmt.Sprintf("%d-%d", r.start, r.end)}, conf)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("could not select pages: %w", err)
	}
	newBuf, applyErr := applyBookmarks(bytes.NewReader(buf.Bytes()), bookmarks, r.start, r.end)
	if applyErr != nil {
		fmt.Fprintf(os.Stderr, "WARN: %v\n", applyErr)
	} else {
		buf = newBuf
	}
	return buf, nil
}

type pageRange struct {
	start int
	end   int
}

func (br pageRange) String() string {
	if br.end < 0 {
		return fmt.Sprintf("%d-end", br.start)
	}
	return fmt.Sprintf("%d-%d", br.start, br.end)
}

var descriptionRe = regexp.MustCompile(`(?i)project\s+description`)
var summaryRe = regexp.MustCompile(`(?i)(project\s+)?summary`)
var referencesRe = regexp.MustCompile(`(?i)references(\s+cited)?`)
var dataManagementPlanRe = regexp.MustCompile(`(?i)data\s+management\s+plan?`)
var mentoringPlanRe = regexp.MustCompile(`(?i)mentoring\s+plan?`)

func hasSection(titleRe *regexp.Regexp, bookmarks []pdfcpu.Bookmark) bool {
	for _, b := range bookmarks {
		if titleRe.MatchString(b.Title) {
			return true
		}
	}
	return false
}

func bookmarkRange(r *regexp.Regexp, bookmarks []pdfcpu.Bookmark, defaultRange pageRange) pageRange {
	for i, b := range bookmarks {
		if r.MatchString(b.Title) {
			start := b.PageFrom
			if i == len(bookmarks)-1 {
				return pageRange{start, -1}
			} else {
				return pageRange{start, bookmarks[i+1].PageFrom - 1}
			}
		}
	}
	return defaultRange
}

func splitPdf(rs io.ReadSeeker, outDir string) error {
	conf := model.NewDefaultConfiguration()
	conf.Offline = true

	bookmarks, err := api.Bookmarks(rs, conf)
	if err != nil {
		return fmt.Errorf("could not read PDF bookmarks: %w", err)
	}

	splitSection := func(name string, titleRe *regexp.Regexp, defaultRange pageRange) error {
		r := bookmarkRange(titleRe, bookmarks, defaultRange)
		if r.start <= 0 {
			return fmt.Errorf("section %s not found", name)
		}
		fmt.Printf("%s: %v\n", name, r)
		buf, err := extractPages(rs, bookmarks, r, conf)
		if err != nil {
			return fmt.Errorf("could not extract %s: %w", name, err)
		}
		err = os.WriteFile(filepath.Join(outDir, fmt.Sprintf("submit-%s.pdf", name)), buf.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("could not write output PDF: %w", err)
		}
		return nil
	}

	err = splitSection("summary", summaryRe, pageRange{1, 1})
	if err != nil {
		return err
	}
	err = splitSection("project-description", descriptionRe, pageRange{2, 16})
	if err != nil {
		return err
	}
	err = splitSection("references", referencesRe, pageRange{17, -1})
	if err != nil {
		return err
	}
	if hasSection(dataManagementPlanRe, bookmarks) {
		err = splitSection("data-mgmt-plan", dataManagementPlanRe, pageRange{})
		if err != nil {
			return err
		}
	}
	if hasSection(mentoringPlanRe, bookmarks) {
		err = splitSection("mentoring-plan", mentoringPlanRe, pageRange{})
		if err != nil {
			return err
		}
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
