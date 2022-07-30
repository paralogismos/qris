// parse.go
//
// Regular expressions used by the `qris` package.
package qris

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Regular Expressions
var citationName = regexp.MustCompile(`^\pL+,\pZs*\pL+`)
var citationYear = regexp.MustCompile(`\pN{4}\pL*`)
var noteEnd = regexp.MustCompile(`jmr$`)
var quoteEnd = regexp.MustCompile(`\t\s*[pP]+\..*`)
var quotePage = regexp.MustCompile(
	`[pP]{1,2}\.\s*[\pNiIvVxXlL]+\s*[,-]*\s*[\pNiIvVxXlL]*`)
var pageNumber = regexp.MustCompile(
	`[\pNiIvVxXlL]+\s*[,-]*\s*[\pNiIvVxXlL]*`)
var parsedFile = regexp.MustCompile(ParsedSuffix + `$`)
var discardFile = regexp.MustCompile(DiscardSuffix + `$`)

// `isParsedFile` returns true if `f` ends with `ParsedSuffix`.
func isParsedFile(f string) bool {
	return parsedFile.FindStringIndex(f) != nil
}

// `isDiscardFile` returns true if `f` ends with `DiscardSuffix`.
func isDiscardFile(f string) bool {
	return discardFile.FindStringIndex(f) != nil
}

func ParseFile(fpath string) ParsedFile {
	rls := cleanLines(getLines(fpath))
	fn := filepath.Base(fpath)
	tit := rls[0].Body
	cit := parseCitation(rls[1])
	qs := Quotes{}
	ds := Lines{}

	for _, l := range rls[2:] {
		q, isQuote := parseQuote(l)
		if n, isNote := parseNote(l); !isQuote && isNote {
			lastQuoteIdx := len(qs) - 1
			if lastQuoteIdx >= 0 {
				qs[lastQuoteIdx].Note = n
			} else {
				cit.Note = n
			}
		} else {
			if isQuote {
				qs = append(qs, q)
			} else {
				ds = append(ds, l)
			}
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
	note := ""

	return newCitation(name, year, body, note)
}

func parseNote(l Line) (string, bool) {
	isNote := false
	body := strings.TrimSpace(l.Body)
	if noteEnd.FindStringIndex(body) != nil {
		isNote = true
	}
	return body, isNote
}

func parseQuote(q Line) (Quote, bool) {
	lineNo, body, page, supp := 0, "", "", ""

	// Malformed page numbers are recoreded using `pageUnknown`.
	const pageUnknown = "?"

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

		if pageMatchIndices == nil {
			// Unable to parse page number
			page = pageUnknown
		} else {
			page = pageNumber.FindString(
				strings.TrimSpace(endMatch[:pageMatchIndices[1]]))

			supp = strings.TrimSpace(endMatch[pageMatchIndices[1]:])
		}

		// Special Case: simple page # at end of quote, no tabs
	} else if bodyEnd := len(q.Body) - 13; bodyEnd > 0 {
		simpleEnd := q.Body[bodyEnd:]
		pageMatchIndices := quotePage.FindStringIndex(simpleEnd)
		isQuote = pageMatchIndices != nil

		if isQuote {
			body = strings.TrimSpace(q.Body[:pageMatchIndices[0]+bodyEnd])
			page = pageNumber.FindString(
				strings.TrimSpace(simpleEnd[pageMatchIndices[0]:]))
		}
	}

	note := ""
	return newQuote(lineNo, body, page, supp, note), isQuote
}
