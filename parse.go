// parse.go
//
// Regular expressions used by the `qris` package.
package qris

import (
	"regexp"
	"strings"
)

// Regular Expressions
var leadingSpace = regexp.MustCompile(`^[\p{Zs}\t]*`)
var blankLine = regexp.MustCompile(`^[\p{Zs}\t]*$`)
var commentLine = regexp.MustCompile(`^#`)

// var trailingSpace = regexp.MustCompile(`[\p{Zs}\t\n]*$`)
var sourceBegin = regexp.MustCompile(`^[\p{Zs}\t]*<\$>`)

var authorLine = regexp.MustCompile(`^>>>`)
var keywordLine = regexp.MustCompile(`^\^[sS]:`)
var supplementalLine = regexp.MustCompile(`%%$`)
var URLLine = regexp.MustCompile(`^[\p{Zs}\t]*https?://`)

var citationName = regexp.MustCompile(`^[^“"{]*`)
var finalPeriod = regexp.MustCompile(`\.$`)
var nameInitialPeriod = regexp.MustCompile(`[\p{Zs}.]{1}[\pL]{1}\.$`)
var citationFamilyName = regexp.MustCompile(`^[^,]*`)

var citationYear = regexp.MustCompile(`[^\pN]\pN{4}\pL?[^\pN]`)
var citationYearAlt = regexp.MustCompile(`[^\pN]\pN{4}\pL?[\pP]?$`)
var citationYearTrim = regexp.MustCompile(`\pN{4}\pL?`)

//var citationYear = regexp.MustCompile(`\pN{4}\pL*`)

var citNoteEnd = regexp.MustCompile(`-nb\.?$`)
var noteEnd = regexp.MustCompile(`jmr\.?$`)

var multiLineQuote = regexp.MustCompile(`^[\p{Zs}\t]*///`)

// A quote end is either tab-delimited pp., or space-delimited pp. with
// at least three spaces as the delimiter.
var quoteEnd = regexp.MustCompile(`\t\p{Zs}*[pP]{1,2}\.?\p{Zs}+[\pNiIvVxXlL\?]*.*`)
var quoteEndAlt = regexp.MustCompile(`\p{Zs}{3,}?[pP]{1,2}\.?\p{Zs}+[\pNiIvVxXlL\?]*.*`)

var quotePage = regexp.MustCompile(
	`[pP]{1,2}\.?\p{Zs}*[\pNiIvVxXlL\?]+[f]{0,2}\p{Zs}*[,\pPd]*\p{Zs}*[\pNiIvVxXlL\?]*[f]{0,2}`)
var pageNumber = regexp.MustCompile(
	`[\pNiIvVxXlL\?]+[f]{0,2}\p{Zs}*[,\pPd]*\p{Zs}*[\pNiIvVxXlL\?]*[f]{0,2}`)

var parsedFile = regexp.MustCompile(parsedSuffix + `$`)
var discardFile = regexp.MustCompile(discardSuffix + `$`)

// `isParsedFile` returns true if `f` ends with `parsedSuffix`.
func isParsedFile(f string) bool {
	return parsedFile.FindStringIndex(f) != nil
}

// `isDiscardFile` returns true if `f` ends with `discardSuffix`.
func isDiscardFile(f string) bool {
	return discardFile.FindStringIndex(f) != nil
}

// THIS FUNCTION SHOULD BE REFACTORED TO STREAMLINE THE LOGIC
func ParseFile(fpath string) ParsedFile {
	rls := getLines(fpath)
	var cit Citation
	var src Source
	qs := []Quote{}
	ds := []Line{}
	sources := []Source{}
	firstSource := true

	inMultiLineQuote := false
	var fullQuote []string
	for _, l := range rls[1:] { // Always ignore first line of input file.
		if commentLine.MatchString(l.Body) {
			continue // Ignore comment lines.
		}
		if firstSource { // first source in the file
			firstSource = false
			cit = parseCitation(l)
			qs = []Quote{}
			ds = []Line{}
			continue
		}
		// THE LOGIC BELOW IS CONVOLUTED.
		// A STATE MACHINE WOULD HELP CLEAN THIS UP.
		//
		// Do not match new source line or new multiline quote
		// if already in a multiline quote.
		if !inMultiLineQuote {
			if sourceBegin.MatchString(l.Body) { // new source encountered
				src = newSource(cit, qs)
				sources = append(sources, src) // save the last source
				cit = parseCitation(l)
				qs = []Quote{} // start new slice of quotes
				continue
			}
			if multiLineQuote.MatchString(l.Body) { // begin multi-line quote
				if q, isQuote := parseQuote(l); isQuote { // really a single-line quote?
					// Cleanup front and back of the actual quote body and save it.
					singleQuote := multiLineQuote.ReplaceAllLiteralString(q.Body[0], "")
					q.Body[0] = strings.TrimSpace(singleQuote)
					qs = append(qs, q)
				} else { // otherwise, actually a multi-line quote
					inMultiLineQuote = true
					// Remove multi-line quote token and surrounding whitespace before collecting the line.
					fullQuote = append(fullQuote,
						strings.TrimSpace(multiLineQuote.ReplaceAllLiteralString(l.Body, "")))
				}
				continue
			}
		}
		q, isQuote := parseQuote(l)
		if isQuote { // quote line may end a multi-line quote
			if inMultiLineQuote {
				q.Body = append(fullQuote, strings.TrimSpace(q.Body[0]))
			} else {
				q.Body[0] = strings.TrimSpace(q.Body[0])
			}
			qs = append(qs, q)
			inMultiLineQuote = false // reset multi-line quote parameters
			fullQuote = nil
			continue
		}
		if inMultiLineQuote { // any line in multi-line quote is saved
			// Remove trailing whitespace.
			fullQuote = append(fullQuote, strings.TrimSpace(l.Body))
			continue
		}

		// Citation notes should only follow a new citation line.
		lastQuoteIdx := len(qs) - 1
		if cn, isCitNote := parseCitNote(l); isCitNote && lastQuoteIdx < 0 {
			cit.Note = cn
			continue
		}

		// The remaining fields should only follow a quote.
		if lastQuoteIdx >= 0 {
			if n, isNote := parseNote(l); isNote {
				qs[lastQuoteIdx].Note = n
				continue
			}
			if a, isAuth := parseAuth(l); isAuth {
				qs[lastQuoteIdx].Auth = a
				continue
			}
			if k, isKeyword := parseKeyword(l); isKeyword {
				qs[lastQuoteIdx].Keyword = k
				continue
			}
			if s, isSupp := parseSupplemental(l); isSupp {
				qs[lastQuoteIdx].Supp = append(qs[lastQuoteIdx].Supp, s)
				continue
			}
			if u, isURL := parseURL(l); isURL {
				qs[lastQuoteIdx].URL = u
				continue
			}
		}
		// Send malformed lines to DISCARD.
		if !blankLine.MatchString(l.Body) {
			ds = append(ds, l)
		}
	}
	src = newSource(cit, qs)
	sources = append(sources, src) // save the last parsed source
	return newParsedFile(fpath, sources, ds)
}

// `parseCitation` parses a line into a `Citation` struct.
func parseCitation(rl Line) Citation {
	tl := sourceBegin.ReplaceAllString(rl.Body, "") // trim `sourceBegin` token
	tl = strings.TrimSpace(tl)

	name := strings.TrimSpace(citationName.FindString(tl))
	if !nameInitialPeriod.MatchString(name) {
		name = finalPeriod.ReplaceAllString(name, "")
	}

	year := citationYearAlt.FindString(tl) // year at end of citation
	if year == "" {
		yearMatches := citationYear.FindAllString(tl, -1) // year buried in citation
		countMatches := len(yearMatches)
		if countMatches > 0 {
			year = yearMatches[countMatches-1]

		}
	}
	year = citationYearTrim.FindString(year)
	body := tl
	note := ""

	return newCitation(name, year, body, note)
}

func parseCitNote(l Line) (string, bool) {
	isCitNote := false
	body := strings.TrimSpace(l.Body)
	if citNoteEnd.FindStringIndex(body) != nil {
		isCitNote = true
		body = strings.TrimSpace(citNoteEnd.ReplaceAllLiteralString(body, ""))
	}
	return body, isCitNote
}

func parseNote(l Line) (string, bool) {
	isNote := false
	body := strings.TrimSpace(l.Body)
	if noteEnd.FindStringIndex(body) != nil {
		isNote = true
	}
	return body, isNote
}

func parseURL(l Line) (string, bool) {
	isURL := false
	url := ""
	body := strings.TrimSpace(l.Body)
	if URLLine.MatchString(body) {
		isURL = true
		// Don't remove http prefix from url.
		url = strings.TrimSpace(body)
	}
	return url, isURL
}

func parseAuth(l Line) (string, bool) {
	isAuth := false
	author := ""
	body := strings.TrimSpace(l.Body)
	if authorLine.MatchString(body) {
		isAuth = true
		author = strings.TrimSpace(authorLine.ReplaceAllString(body, ""))
	}
	return author, isAuth
}

func parseKeyword(l Line) (string, bool) {
	isKeyword := false
	keyword := ""
	body := strings.TrimSpace(l.Body)
	if keywordLine.MatchString(body) {
		isKeyword = true
		keyword = strings.TrimSpace(keywordLine.ReplaceAllString(body, ""))
	}
	return keyword, isKeyword
}

func parseSupplemental(l Line) (string, bool) {
	isSupp := false
	supp := ""
	body := strings.TrimSpace(l.Body)
	if supplementalLine.MatchString(body) {
		isSupp = true
		supp = strings.TrimSpace(supplementalLine.ReplaceAllString(body, ""))
	}
	return supp, isSupp
}

func parseQuote(q Line) (Quote, bool) {
	var auth, kw, page, note, url string
	var body, supp []string

	// Malformed page numbers are recorded using `pageUnknown`.
	const pageUnknown = "?"

	// Predominant Case: tab-delimited quote ends
	endMatchIndices := quoteEnd.FindStringIndex(q.Body)

	// Alternate Case: space-delimited quote ends
	if endMatchIndices == nil {
		endMatchIndices = quoteEndAlt.FindStringIndex(q.Body)
	}

	isQuote := endMatchIndices != nil

	if isQuote {
		// Split quote into body and end
		bodyMatch := q.Body[:endMatchIndices[0]]
		endMatch := q.Body[endMatchIndices[0]:]

		// Get quote body
		body = append(body, strings.TrimSpace(bodyMatch))
		// Split end into page and supplementary field
		pageMatchIndices := quotePage.FindStringIndex(endMatch)

		if pageMatchIndices == nil {
			// Unable to parse page number
			page = pageUnknown
		} else {
			page = pageNumber.FindString(strings.TrimSpace(endMatch))
		}
	}
	return newQuote(auth, kw, body, page, supp, note, url), isQuote
}
