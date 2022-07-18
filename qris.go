// qris.go
//
// experiments in parsing ris files
//
// Assumptions:
//
//   The first line is the title of the source
//
//   The second line is a citation
//     - the citation should be parsed into family-name, year, and raw-citation
//
//   The remaining lines are quotes IF they end in a page number, followed
//   by an optional supplementary field.
//     - predicate to determine quoteness
//     - parse raw quotes into quote, page-number, and supplementary-field
//       - supplementary field starts with "^w" delimiter to be discarded
//
//   Some quotes are followed by a descriptive line
//     - can these be identified?
//
//   Blank lines are discarded
//
//   Other lines are written to a review file in the format:
//     - >[line #]
//       [discarded line]
//

package qris

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
}

func newCitation(name, year, body string) Citation {
	return Citation{
		Name: name,
		Year: year,
		Body: body,
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
}

func newQuote(lineNo int, body, page, supp string) Quote {
	return Quote{
		LineNo: lineNo,
		Body:   body,
		Page:   page,
		Supp:   supp,
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

// Regular Expressions
var citationName = regexp.MustCompile(`^\pL+`)
var citationYear = regexp.MustCompile(`\pN{4}`)
var quoteEnd = regexp.MustCompile(`\t\s*p+\..*`)
var quotePage = regexp.MustCompile(`p{1,2}\.\s*\pN+,*\s*\pN*`)

func ParseFile(fpath string) ParsedFile {
	rls := cleanLines(getLines(fpath))
	fn := filepath.Base(fpath)
	tit := rls[0].Body
	cit := parseCitation(rls[1])
	qs := Quotes{}
	ds := Lines{}

	for _, l := range rls[2:] {
		if q, ok := parseQuote(l); ok {
			qs = append(qs, q) // no page or supp yet
		} else {
			ds = append(ds, l)
		}
	}

	return newParsedFile(fn, tit, cit, qs, ds)
}

// `parseCitation` parses a line into a `Citation` struct.
func parseCitation(rl Line) Citation {
	name := citationName.FindString(rl.Body)

	year := ""
	yearMatches := citationYear.FindAllStringSubmatch(rl.Body, -1)
	countMatches := len(yearMatches)
	if countMatches > 0 {
		year = yearMatches[countMatches-1][0]
	}

	body := rl.Body

	return newCitation(name, year, body) // no name or year yet
}

// A function stub that always returns an unparsed quote at the moment.
func parseQuote(q Line) (Quote, bool) {
	lineNo, body, page, supp := 0, "", "", ""

	// Predominant Case: tab-delimited quote ends
	endMatchIndices := quoteEnd.FindStringIndex(q.Body)
	isQuote := endMatchIndices != nil

	if isQuote {
		// Split quote into body and end
		bodyMatch := q.Body[:endMatchIndices[0]]
		endMatch := q.Body[endMatchIndices[0]:]

		// Get quote body
		body = strings.TrimSpace(bodyMatch)

		// Split end into page and supplementary field
		pageMatchIndices := quotePage.FindStringIndex(endMatch)
		page = strings.TrimSpace(endMatch[:pageMatchIndices[1]])
		supp = strings.TrimSpace(endMatch[pageMatchIndices[1]:])

		// Special Case: simple page # at end of quote, no tabs
	} else if bodyEnd := len(q.Body) - 13; bodyEnd > 0 {
		simpleEnd := q.Body[bodyEnd:]
		pageMatchIndices := quotePage.FindStringIndex(simpleEnd)
		isQuote = pageMatchIndices != nil

		if isQuote {
			body = strings.TrimSpace(q.Body[:pageMatchIndices[0]+bodyEnd])
			page = strings.TrimSpace(simpleEnd[pageMatchIndices[0]:])
		}
	}

	return newQuote(lineNo, body, page, supp), isQuote
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

	id := strings.TrimSuffix(pf.Filename, filepath.Ext(pf.Filename))
	src := pf.Title
	cit := pf.Citation.Body
	name := pf.Citation.Name
	year := pf.Citation.Year

	for _, q := range pf.Quotes {
		fmt.Fprintln(file, "TY  - ABST")
		fmt.Fprintln(file, "ID  -", id)
		fmt.Fprintln(file, "AB  -", cit)
		fmt.Fprintln(file, "A1  -", name)
		fmt.Fprintln(file, "Y1  -", year)
		fmt.Fprintln(file, "N1  -", q.Body)
		fmt.Fprintln(file, "SP  -", q.Page)
		fmt.Fprintln(file, "AD  -", q.Supp)
		fmt.Fprintln(file, "ER  -")
	}
}
