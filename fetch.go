// fetch.go
//
// Manage different input file types.
package qris

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var isDocx = regexp.MustCompile(`\.docx$`)
var isTxt = regexp.MustCompile(`\.txt$`)
var isDiscard = regexp.MustCompile(discardSuffix + `$`)

//var isRis = regexp.MustCompile(`\.ris`)
//var isParsed = regexp.MustCompile(parsedSuffix + `$`)

var tabTag = regexp.MustCompile(`<w:tab/>`)
var noBreakHyphen = regexp.MustCompile(`<w:noBreakHyphen/>`)
var htmlOpen = regexp.MustCompile(`<w:hyperlink [^>]*>`)
var htmlClose = regexp.MustCompile(`</w:hyperlink>`)

var tabElement = "<w:t>\t</w:t>"
var hyphenElement = "<w:t>-</w:t>"

// A `Line` is a string of content coupled with a line number reference to the
// original file; 1-indexed.
type Line struct {
	LineNo int
	Body   string
}

func newLine(lineNo int, body string) Line {
	return Line{
		LineNo: lineNo,
		Body:   body,
	}
}

func isDocxFile(s string) bool {
	return isDocx.MatchString(s)
}

func isTxtFile(s string) bool {
	return isTxt.MatchString(s)
}

// `isDiscardFile` returns true if `s` ends with `discardSuffix`.
func isDiscardFile(s string) bool {
	return isDiscard.MatchString(s)
	//	return discardFile.FindStringIndex(f) != nil
}

// `notInputFile` returns true if `s` should NOT be processed.
func notInputFile(s string) bool {
	return !(isDocxFile(s) ||
		(isTxtFile(s) && !isDiscardFile(s)))
}

// Takes a `fpath` argument which leads to a .txt file and
// returns a slice of `Line`s.
func TxtToLines(fpath string) ([]Line, error) {
	file, err := os.Open(fpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	rawLines := []Line{}
	scanner := bufio.NewScanner(file)
	lineNo := 1
	for scanner.Scan() {
		rawLines = append(rawLines, newLine(lineNo, scanner.Text()))
		lineNo++
	}

	return rawLines, err
}

// Takes a `path` argument which leads to a .docx file and
// returns a slice of `Line`s.
func DocxToLines(path string) ([]Line, error) {
	docSection := "word/document.xml"
	lines := []Line{}

	// Decompress .docx file
	r, err := zip.OpenReader(path)
	if err != nil {
		return lines, err
	}
	defer r.Close()

	var docBytes []byte
	for _, f := range r.File {
		if f.Name == docSection {
			ts, err := f.Open()
			if err != nil {
				return lines, err
			}
			docBytes, err = io.ReadAll(ts)
			if err != nil {
				return lines, err
			}
			ts.Close()
			break
		}
	}

	// Initial preprocessing of file content.
	sDocBytes := string(docBytes)

	// Replace special tab elements with tab characters before parsing.
	sDocBytes = tabTag.ReplaceAllString(sDocBytes, tabElement)

	// Convert non-breaking hyphen characters to ASCII dashes.
	sDocBytes = noBreakHyphen.ReplaceAllString(sDocBytes, hyphenElement)

	// Expose hyperlink text before parsing.
	sDocBytes = htmlOpen.ReplaceAllString(sDocBytes, "")
	sDocBytes = htmlClose.ReplaceAllString(sDocBytes, "")

	docBytes = []byte(sDocBytes)

	// Replace html tags with text.

	type Rawline struct {
		Runs []string `xml:"r>t"`
	}

	type Document struct {
		Lines []Rawline `xml:"body>p"`
	}

	document := Document{}
	err = xml.Unmarshal(docBytes, &document)
	if err != nil {
		return lines, err
	}

	for n, rl := range document.Lines {
		line := strings.Join(rl.Runs, "")
		lines = append(lines, newLine(n, line))
	}
	return lines, err
}
