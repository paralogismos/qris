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
var sourceBegin = regexp.MustCompile(`^<\$>`)

var authorLine = regexp.MustCompile(`^>>>`)
var keywordLine = regexp.MustCompile(`^\^[sS]:`)
var supplementalLine = regexp.MustCompile(`%%$`)

//var citationName = regexp.MustCompile(`^\pL+,\pZs*\pL+`)
var citationName = regexp.MustCompile(`^[^“"{]*`)
var citationFamilyName = regexp.MustCompile(`^[^,]*`)
var citationYear = regexp.MustCompile(`\pN{4}\pL*`)

var noteEnd = regexp.MustCompile(`jmr$`)

//var noteEndAlt = regexp.MustCompile(`jmr.*`) // Was this a mistake?
var noteEndAlt = regexp.MustCompile(`jmr.$`)

var multiLineQuote = regexp.MustCompile(`^///`)

// A quote end is either tab-delimited pp., or space-delimited pp. with
// at least three spaces as the delimiter.
var quoteEnd = regexp.MustCompile(`\t\s*[pP]+\..*`)
var quoteEndAlt = regexp.MustCompile(`\s{3,}?[pP]+\..*`)

var quotePage = regexp.MustCompile(
	`[pP]{1,2}\.\s*[\pNiIvVxXlL\?]+\s*[,-]*\s*[\pNiIvVxXlL\?]*`)
var pageNumber = regexp.MustCompile(
	`[\pNiIvVxXlL\?]+\s*[,-]*\s*[\pNiIvVxXlL\?]*`)

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
	rls := cleanLines(getLines(fpath))
	fn := filepath.Base(fpath)
	tit := rls[0].Body
	var cit Citation
	var src Source
	qs := []Quote{}
	ds := []Line{}
	sources := []Source{}
	firstSource := true

	inMultiLineQuote := false
	fullQuote := ""
	for _, l := range rls[1:] {
		if firstSource { // first source in the file
			firstSource = false
			cit = parseCitation(l)
			qs = []Quote{}
			ds = []Line{}
			continue
		}
		if sourceBegin.MatchString(l.Body) { // new source encountered
			src = newSource(cit, qs)
			sources = append(sources, src) // save the last source
			cit = parseCitation(l)
			qs = []Quote{} // start new slice of quotes
			continue
		}
		if multiLineQuote.MatchString(l.Body) { // begin multi-line quote
			inMultiLineQuote = true
			fullQuote = multiLineQuote.ReplaceAllString(l.Body, "")
			// fullQuote = strings.TrimSpace(fullQuote)
			continue
		}
		q, isQuote := parseQuote(l)
		if isQuote { // quote line may end a multi-line quote
			q.Body = fullQuote + q.Body
			qs = append(qs, q)
			inMultiLineQuote = false // reset multi-line quote parameters
			fullQuote = ""
			continue
		}
		if inMultiLineQuote { // any line in multi-line quote is saved
			fullQuote += l.Body
			//fullQuote += strings.TrimSpace(l.Body)
			continue
		}
		lastQuoteIdx := len(qs) - 1
		if n, isNote := parseNote(l); isNote {
			if lastQuoteIdx >= 0 {
				qs[lastQuoteIdx].Note = n
			} else {
				cit.Note = n
			}
		} else if a, isAuth := parseAuth(l); isAuth {
			qs[lastQuoteIdx].Auth = a
		} else if k, isKeyword := parseKeyword(l); isKeyword {
			qs[lastQuoteIdx].Keyword = k
		} else if s, isSupp := parseSupplemental(l); isSupp {
			qs[lastQuoteIdx].Supp = s
		} else {
			ds = append(ds, l)
		}
	}
	src = newSource(cit, qs)
	sources = append(sources, src) // save the last parsed source
	return newParsedFile(fn, tit, sources, ds)
}

// `parseCitation` parses a line into a `Citation` struct.
func parseCitation(rl Line) Citation {
	tl := sourceBegin.ReplaceAllString(rl.Body, "") // trim `sourceBegin` token
	tl = strings.TrimSpace(tl)
	name := citationName.FindString(tl)

	year := ""
	yearMatches := citationYear.FindAllStringSubmatch(tl, -1)
	countMatches := len(yearMatches)
	if countMatches > 0 {
		year = yearMatches[countMatches-1][0]
	}

	body := tl
	note := ""

	return newCitation(name, year, body, note)
}

func parseNote(l Line) (string, bool) {
	isNote := false
	body := strings.TrimSpace(l.Body)
	if noteEnd.FindStringIndex(body) != nil {
		isNote = true
	} else if noteEndAlt.FindStringIndex(body) != nil {
		isNote = true
	}
	return body, isNote
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
	lineNo, auth, kw, body, page, supp, note := 0, "", "", "", "", "", ""

	// Malformed page numbers are recoreded using `pageUnknown`.
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
		body = strings.TrimSpace(bodyMatch)

		// Split end into page and supplementary field
		pageMatchIndices := quotePage.FindStringIndex(endMatch)

		if pageMatchIndices == nil {
			// Unable to parse page number
			page = pageUnknown
		} else {
			page = pageNumber.FindString(strings.TrimSpace(endMatch))
		}

		// Special Case: page # at end of quote with no tabs and
		// only one or two spaces delimiting
	} else if bodyEnd := len(q.Body) - 30; bodyEnd > 0 {
		simpleEnd := q.Body[bodyEnd:]
		pageMatchIndices := quotePage.FindStringIndex(simpleEnd)
		isQuote = pageMatchIndices != nil

		if isQuote {
			body = strings.TrimSpace(q.Body[:pageMatchIndices[0]+bodyEnd])
			page = pageNumber.FindString(
				strings.TrimSpace(simpleEnd[pageMatchIndices[0]:]))
		}
	}

	return newQuote(lineNo, auth, kw, body, page, supp, note), isQuote
}
