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
var leadingSpace = regexp.MustCompile(`^[\p{Zs}\t]*`)

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
var citationYear = regexp.MustCompile(`\pN{4}\pL*`)

var citNoteEnd = regexp.MustCompile(`-nb\.?$`)
var noteEnd = regexp.MustCompile(`jmr$`)
var noteEndAlt = regexp.MustCompile(`jmr.$`)

var multiLineQuote = regexp.MustCompile(`^[\p{Zs}\t]*///`)

// A quote end is either tab-delimited pp., or space-delimited pp. with
// at least three spaces as the delimiter.
var quoteEnd = regexp.MustCompile(`\t\p{Zs}*[pP]{1,2}\.?\p{Zs}+[\pNiIvVxXlL\?]*.*`)
var quoteEndAlt = regexp.MustCompile(`\p{Zs}{3,}?[pP]{1,2}\.?\p{Zs}+[\pNiIvVxXlL\?]*.*`)

var quotePage = regexp.MustCompile(
	`[pP]{1,2}\.?\p{Zs}*[\pNiIvVxXlL\?]+\p{Zs}*[,-]*\p{Zs}*[\pNiIvVxXlL\?]*`)
var pageNumber = regexp.MustCompile(
	`[\pNiIvVxXlL\?]+\p{Zs}*[,-]*\p{Zs}*[\pNiIvVxXlL\?]*`)

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
	// rls := cleanLines(getLines(fpath))
	rls := getLines(fpath)
	fn := filepath.Base(fpath)
	tit := rls[0].Body
	var cit Citation
	var src Source
	qs := []Quote{}
	ds := []Line{}
	sources := []Source{}
	firstSource := true

	inMultiLineQuote := false
	var fullQuote []string
	for _, l := range rls[1:] {
		if firstSource { // first source in the file
			firstSource = false
			cit = parseCitation(l)
			qs = []Quote{}
			ds = []Line{}
			continue
		}
		// FIX: Don't match new source markup when in multiline quote!
		//if !inMultiLineQuote && sourceBegin.MatchString(l.Body) {
		if sourceBegin.MatchString(l.Body) { // new source encountered
			src = newSource(cit, qs)
			sources = append(sources, src) // save the last source
			cit = parseCitation(l)
			qs = []Quote{} // start new slice of quotes
			continue
		}
		// FIX: Don't match multiline quote markup when in multiline quote!
		//if !inMultiLineQuote && multiLineQuote.MatchString(l.Body) {
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
				// Trim whitespace and attach a newline.
				//fullQuote = strings.TrimSpace(fullQuote) + "\n"
			}
			continue
		}
		q, isQuote := parseQuote(l)
		if isQuote { // quote line may end a multi-line quote
			if inMultiLineQuote {
				//q.Body = append(fullQuote, trailingSpace.ReplaceAllLiteralString(q.Body[0], ""))
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
			//fullQuote = append(fullQuote, trailingSpace.ReplaceAllLiteralString(l.Body, ""))
			fullQuote = append(fullQuote, strings.TrimSpace(l.Body))
			continue
		}
		lastQuoteIdx := len(qs) - 1
		if cn, isCitNote := parseCitNote(l); isCitNote {
			cit.Note = cn
		} else if n, isNote := parseNote(l); isNote {
			if lastQuoteIdx >= 0 {
				qs[lastQuoteIdx].Note = n
			}
		} else if a, isAuth := parseAuth(l); isAuth {
			qs[lastQuoteIdx].Auth = a
		} else if k, isKeyword := parseKeyword(l); isKeyword {
			qs[lastQuoteIdx].Keyword = k
		} else if s, isSupp := parseSupplemental(l); isSupp {
			qs[lastQuoteIdx].Supp = append(qs[lastQuoteIdx].Supp, s)
		} else if u, isURL := parseURL(l); isURL {
			qs[lastQuoteIdx].URL = u
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
	name := strings.TrimSpace(citationName.FindString(tl))
	if !nameInitialPeriod.MatchString(name) {
		name = finalPeriod.ReplaceAllString(name, "")
	}
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
	} else if noteEndAlt.FindStringIndex(body) != nil {
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
		//body = strings.TrimSpace(bodyMatch)
		//body = append(body, trailingSpace.ReplaceAllLiteralString(bodyMatch, ""))
		body = append(body, strings.TrimSpace(bodyMatch))
		// body = bodyMatch // Why doesn't this work?
		// Split end into page and supplementary field
		pageMatchIndices := quotePage.FindStringIndex(endMatch)

		if pageMatchIndices == nil {
			// Unable to parse page number
			page = pageUnknown
		} else {
			page = pageNumber.FindString(strings.TrimSpace(endMatch))
		}

		// REMOVE THIS SPECIAL CASE: IT INTERFERES WITH OTHER MARKUPS THAT
		// INCLUDE PAGE NUMBERS!
		// Special Case: page # at end of quote with no tabs and
		// only one or two spaces delimiting
	} //  else if bodyEnd := len(q.Body) - 30; bodyEnd > 0 {
	// 	simpleEnd := q.Body[bodyEnd:]
	// 	pageMatchIndices := quotePage.FindStringIndex(simpleEnd)
	// 	isQuote = pageMatchIndices != nil

	// 	if isQuote {
	// 		//	body = strings.TrimSpace(q.Body[:pageMatchIndices[0]+bodyEnd])
	// 		body = trailingSpace.ReplaceAllLiteralString(q.Body[:pageMatchIndices[0]+bodyEnd], "")
	// 		// I thought I could use this and trim in the calling function....
	// 		//			body = q.Body[:pageMatchIndices[0]+bodyEnd]
	// 		page = pageNumber.FindString(
	// 			strings.TrimSpace(simpleEnd[pageMatchIndices[0]:]))
	// 	}
	// }

	return newQuote(auth, kw, body, page, supp, note, url), isQuote
}
