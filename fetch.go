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
var tabTag = regexp.MustCompile(`<w:tab/>`)
var htmlOpen = regexp.MustCompile(`<w:hyperlink [^>]*>`)
var htmlClose = regexp.MustCompile(`</w:hyperlink>`)

var tabElement = "<w:t>\t</w:t>"

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

func IsDocxFile(s string) bool {
	return isDocx.MatchString(s)
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
		rawLines = append(rawLines, newLine(lineNo, tidyString(scanner.Text())))
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

	// Replace special tab elements with tab characters before parsing.
	sDocBytes := string(docBytes)
	sDocBytes = tabTag.ReplaceAllString(sDocBytes, tabElement)
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
		line = tidyString(line)
		lines = append(lines, newLine(n, line))
	}
	return lines, err
}
