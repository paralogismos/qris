// qris.go
//
// experiments in parsing ris files
//
// Assumptions:
//
//   The first line is the title of the source
//
//   The second line is a citation
//     - the citation should be parsed into name, year, raw-citation,
//       and a note-field
//       - notes preceding any quotes are assumed to be attached to the citation
//
//   The remaining lines are quotes IF they end in a page number, followed
//   by an optional supplementary field.
//     - predicate to determine quoteness
//     - parse raw quotes into quote, page-number, supplementary-field,
//       and note-field
//       - supplementary field starts with "^w" delimiter to be discarded
//
//   Some quotes are followed by a note ending with "-jmr"
//     - these should be identified and added to the Note field of a Quote
//       of the most recent quote, if one exists.
//
//   Blank lines are discarded
//
//   Other lines are written to a review file in the format:
//     - >[line #]
//       [discarded line]
//
// TODO:
//
// x - Omit quote teaser
// x - Tag quote body with T1 instead of T3
// x - Preserve contiguous letter suffixes in citation dates
// x - Substitute "?" for "UNKNOWN" in malformed page number cases
// x - Capture page numbers in roman numerals
// x - Strip page number indicators from page numbers
// x - Fix bug in parsing supplementary fields ending in note marker ("jmr")
//
// _ - Add functionality to store up to N notes following a quote.
//     - Store notes in a slice or dictionary
//     - Map notes array to numbered tags, e.g., C1, C2, ....
//     - Need to discuss this with Jack
//

package qris

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

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

// `Lines` is a slice of `Line`s.
type Lines []Line

// The first line of the file is assumed to be the source title.

// Parsed from the second line of the file  into family-name, year,
// and raw-citation.
type Citation struct {
	Name string
	Year string // type doesn't matter since we are writing to a file
	Body string
	Note string
}

func newCitation(name, year, body, note string) Citation {
	return Citation{
		Name: name,
		Year: year,
		Body: body,
		Note: note,
	}
}

// Parsed from a `Line` for which `IsQuote` is true.
// Includes line number from original file.
// Note that the original `Line` body should have at least a page number, and
// possibly a supplementary entry, associated with it. These have been extracted
// in a `Quote` and placed in their own fields.
type Quote struct {
	LineNo int
	Body   string
	Page   string
	Supp   string
	Note   string
}

func newQuote(lineNo int, body, page, supp, note string) Quote {
	return Quote{
		LineNo: lineNo,
		Body:   body,
		Page:   page,
		Supp:   supp,
		Note:   note,
	}
}

// `Quotes` is a slice of `Quote`s.
type Quotes []Quote

// Results of parsing one file.
// `Discards` is a slice of `Line`s which aren't quotes, to be reviewed manually
// by the user.

type ParsedFile struct {
	Filename string
	Title    string // first line of parsed file
	Citation Citation
	Quotes   Quotes
	Discards Lines
}

func newParsedFile(fn, tit string, cit Citation, qs Quotes, ds Lines) ParsedFile {
	return ParsedFile{
		Filename: fn,
		Title:    tit,
		Citation: cit,
		Quotes:   qs,
		Discards: ds,
	}
}

// Creates a list of filenames from the text file specified by `fpath`.
// This file should have one filename per line.
func GetFileList(fpath string) []string {
	file, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	files := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}

	return files
}

// Takes a file specified by `fpath` and returns a map containing
// raw lines from the file keyed by the original line number on which
// each line occurred.
func getLines(fpath string) Lines {
	file, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	rawLines := []Line{}
	scanner := bufio.NewScanner(file)
	lineNo := 1
	for scanner.Scan() {
		rawLines = append(rawLines, newLine(lineNo, scanner.Text()))
		lineNo++
	}

	return rawLines
}

// Removes all empty lines, including whitespace lines,
// from a `Lines` instance.
//
// Strip leading and trailing whitespace from each `Line`?
func cleanLines(lines Lines) Lines {
	cls := Lines{}

	for _, l := range lines {
		if l.Body == "" {
			continue
		} else {
			cls = append(cls, l)
		}
	}

	return cls
}

func WriteDiscards(ds Lines, fname string) {
	file, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for _, d := range ds {
		fmt.Fprintln(file, "<", d.LineNo, ">")
		fmt.Fprintln(file, d.Body)
	}
}

func WriteQuotes(pf *ParsedFile, fname string) {
	file, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// batch ID
	bid := filepath.Base(filepath.Dir(fname))

	// file ID
	fid := strings.TrimSuffix(pf.Filename, filepath.Ext(pf.Filename))

	// timestamp: when file was processed
	tstamp := time.Now().Format("2006/01/02")

	citBody := pf.Citation.Body
	citName := pf.Citation.Name
	citYear := pf.Citation.Year
	citNote := pf.Citation.Note

	for _, q := range pf.Quotes {

		// Abstract to hold first characters of quote body.
		// Removed functionality; will remove completely as soon as verified.
		/*
			const abstSize = 100
			var abst string
			if len(q.Body) < abstSize {
				abst = q.Body
			} else {
				abst = q.Body[:abstSize]
			}
		*/

		fmt.Fprintln(file, "TY  - Generic")
		fmt.Fprintln(file, "VL  -", bid)
		fmt.Fprintln(file, "UR  -", fid)
		fmt.Fprintln(file, "AD  -", tstamp)
		fmt.Fprintln(file, "AB  -", citBody)
		fmt.Fprintln(file, "A1  -", citName)
		fmt.Fprintln(file, "Y1  -", citYear)
		fmt.Fprintln(file, "T2  -", citNote)
		fmt.Fprintln(file, "T1  -", q.Body)
		fmt.Fprintln(file, "SP  -", q.Page)
		fmt.Fprintln(file, "PB  -", q.Supp)
		fmt.Fprintln(file, "CY  -", q.Note)
		fmt.Fprintln(file, "ER  -")
	}
}

func ValidateUTF8(fpath string) bool {
	file, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 1
	isValid := true
	for scanner.Scan() {
		if !utf8.ValidString(scanner.Text()) {
			fmt.Printf("  + First invalid UTF8 in line %d\n", lineNo)
			isValid = false
			break
		}
		lineNo++
	}
	return isValid
}
