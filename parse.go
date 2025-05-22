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

type ParseState int

const (
	Start ParseState = iota
	InSource
	InMultiQuote
	InQuote
	Finished
)

type LineType int

const (
	Unknown LineType = iota
	Blank
	Comment
)

func ProcessFile(fpath string) ParsedFile {
	var pf ParsedFile
	pf.Filepath = fpath
	pf.State = Start
	rls := getLines(fpath)
	for _, l := range rls[1:] { // Always ignore first line of input file.
		if isSkipLine(l, pf) {
			continue
		}
		processLine(l, &pf)
	}
	pf.State = Finished
	return pf
}

func processLine(l Line, pf *ParsedFile) /*error*/ {
	body := l.Body
	switch pf.State {
	case Start:
		if parseCitation(body, pf) {
			break
		}
	case InSource:
		if parseCitNote(body, pf) || parseQuote(body, pf) || parseMultiQuote(body, pf) {
			break
		}
	case InMultiQuote:
		if parseQuote(body, pf) { // This line ends a multi-line quote.
			break
		}
		// All non-skipped, non-single-quote lines are captured.
		sourceIdx := len(pf.Sources) - 1
		quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
		pf.Sources[sourceIdx].Quotes[quoteIdx].Body =
			append(pf.Sources[sourceIdx].Quotes[quoteIdx].Body, strings.TrimSpace(body))
	case InQuote:
		if parseCitation(body, pf) || parseQuote(body, pf) ||
			parseMultiQuote(body, pf) || parseNote(body, pf) ||
			parseAuth(body, pf) || parseKeyword(body, pf) ||
			parseSupplemental(body, pf) || parseURL(body, pf) {
			break
		}
	default:
		pf.Discards = append(pf.Discards, l)
	}
}

// THIS FUNCTION SHOULD BE REFACTORED TO STREAMLINE THE LOGIC
// func ParseFile(fpath string) ParsedFile {
// 	rls := getLines(fpath)
// 	var cit Citation
// 	var src Source
// 	qs := []Quote{}
// 	ds := []Line{}
// 	sources := []Source{}
// 	firstSource := true

// 	inMultiLineQuote := false
// 	var fullQuote []string
// 	for _, l := range rls[1:] { // Always ignore first line of input file.
// 		if commentLine.MatchString(l.Body) {
// 			continue // Ignore comment lines.
// 		}
// 		if firstSource { // first source in the file
// 			firstSource = false
// 			cit = parseCitation(l)
// 			qs = []Quote{}
// 			ds = []Line{}
// 			continue
// 		}
// 		// THE LOGIC BELOW IS CONVOLUTED.
// 		// A STATE MACHINE WOULD HELP CLEAN THIS UP.
// 		//
// 		// Do not match new source line or new multiline quote
// 		// if already in a multiline quote.
// 		if !inMultiLineQuote {
// 			if sourceBegin.MatchString(l.Body) { // new source encountered
// 				src = newSource(cit, qs)
// 				sources = append(sources, src) // save the last source
// 				cit = parseCitation(l)
// 				qs = []Quote{} // start new slice of quotes
// 				continue
// 			}
// 			if multiLineQuote.MatchString(l.Body) { // begin multi-line quote
// 				if q, isQuote := parseQuote(l); isQuote { // really a single-line quote?
// 					// Cleanup front and back of the actual quote body and save it.
// 					singleQuote := multiLineQuote.ReplaceAllLiteralString(q.Body[0], "")
// 					q.Body[0] = strings.TrimSpace(singleQuote)
// 					qs = append(qs, q)
// 				} else { // otherwise, actually a multi-line quote
// 					inMultiLineQuote = true
// 					// Remove multi-line quote token and surrounding whitespace before collecting the line.
// 					fullQuote = append(fullQuote,
// 						strings.TrimSpace(multiLineQuote.ReplaceAllLiteralString(l.Body, "")))
// 				}
// 				continue
// 			}
// 		}
// 		q, isQuote := parseQuote(l)
// 		if isQuote { // quote line may end a multi-line quote
// 			if inMultiLineQuote {
// 				q.Body = append(fullQuote, strings.TrimSpace(q.Body[0]))
// 			} else {
// 				q.Body[0] = strings.TrimSpace(q.Body[0])
// 			}
// 			qs = append(qs, q)
// 			inMultiLineQuote = false // reset multi-line quote parameters
// 			fullQuote = nil
// 			continue
// 		}
// 		if inMultiLineQuote { // any line in multi-line quote is saved
// 			// Remove trailing whitespace.
// 			fullQuote = append(fullQuote, strings.TrimSpace(l.Body))
// 			continue
// 		}

// 		// Citation notes should only follow a new citation line.
// 		lastQuoteIdx := len(qs) - 1
// 		if cn, isCitNote := parseCitNote(l); isCitNote && lastQuoteIdx < 0 {
// 			cit.Note = cn
// 			continue
// 		}

// 		// The remaining fields should only follow a quote.
// 		if lastQuoteIdx >= 0 {
// 			if n, isNote := parseNote(l); isNote {
// 				qs[lastQuoteIdx].Note = n
// 				continue
// 			}
// 			if a, isAuth := parseAuth(l); isAuth {
// 				qs[lastQuoteIdx].Auth = a
// 				continue
// 			}
// 			if k, isKeyword := parseKeyword(l); isKeyword {
// 				qs[lastQuoteIdx].Keyword = k
// 				continue
// 			}
// 			if s, isSupp := parseSupplemental(l); isSupp {
// 				qs[lastQuoteIdx].Supp = append(qs[lastQuoteIdx].Supp, s)
// 				continue
// 			}
// 			if u, isURL := parseURL(l); isURL {
// 				qs[lastQuoteIdx].URL = u
// 				continue
// 			}
// 		}
// 		// Send malformed lines to DISCARD.
// 		if !blankLine.MatchString(l.Body) {
// 			ds = append(ds, l)
// 		}
// 	}
// 	src = newSource(cit, qs)
// 	sources = append(sources, src) // save the last parsed source
// 	return newParsedFile(fpath, sources, ds)
// }

// `isSkipLine` returns `true` if `l` should be ignored during processing,
// or `false` otherwise.
func isSkipLine(l Line, pf ParsedFile) bool {
	body := l.Body
	isSkip := false
	if commentLine.MatchString(body) { // Always ignore comments.
		isSkip = true
	}
	if blankLine.MatchString(body) && pf.State != InMultiQuote { // Usually ignore blank lines.
		isSkip = true
	}
	return isSkip
}

// `parseCitation` parses a line into a `Citation` struct.
// func parseCitation(rl Line) Citation {
//
// `parseCitation` parses the body of a line into a `Citation`, adds it to the
// list of `Source`s of the provided `ParsedFile`, updates the `ParseState`, and
// returns `true`, or returns `false` if the provided `Line` is not a citation.
func parseCitation(l string, pf *ParsedFile) bool {
	if !sourceBegin.MatchString(l) { // This line is not a citation.
		return false
	}
	tl := sourceBegin.ReplaceAllString(l, "") // trim `sourceBegin` token
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
	//body := tl
	//note := ""

	cit := Citation{
		Name: name,
		Year: year,
		Body: tl,
	}
	src := Source{Citation: cit}
	pf.Sources = append(pf.Sources, src)
	pf.State = InSource
	return true
	//return newCitation(name, year, body, note)
}

// func parseCitNote(l Line) (string, bool) {
func parseCitNote(l string, pf *ParsedFile) bool {
	isCitNote := false
	body := strings.TrimSpace(l)
	if citNoteEnd.FindStringIndex(body) != nil {
		isCitNote = true
		// body = strings.TrimSpace(citNoteEnd.ReplaceAllLiteralString(body, ""))
		sourceIdx := len(pf.Sources) - 1
		pf.Sources[sourceIdx].Citation.Note =
			strings.TrimSpace(citNoteEnd.ReplaceAllLiteralString(body, ""))
	}
	return isCitNote
	// return body, isCitNote
}

func parseQuote(l string, pf *ParsedFile) bool {
	// Malformed page numbers are recorded using `pageUnknown`.
	const pageUnknown = "?"
	var page string

	// Predominant Case: tab-delimited quote ends
	endMatchIndices := quoteEnd.FindStringIndex(l)

	// Alternate Case: space-delimited quote ends
	if endMatchIndices == nil {
		endMatchIndices = quoteEndAlt.FindStringIndex(l)
	}
	isQuote := endMatchIndices != nil

	if isQuote {
		// Split quote into body and end
		bodyMatch := l[:endMatchIndices[0]]
		endMatch := l[endMatchIndices[0]:]

		// Get quote body: single-line quote may begin with multi-quote token.
		//body := []string{strings.TrimSpace(multiLineQuote.ReplaceAllLiteralString(bodyMatch, ""))}
		body := strings.TrimSpace(multiLineQuote.ReplaceAllLiteralString(bodyMatch, ""))

		// Split end into page and supplementary field
		pageMatchIndices := quotePage.FindStringIndex(endMatch)

		if pageMatchIndices == nil {
			// Unable to parse page number
			page = pageUnknown
		} else {
			page = pageNumber.FindString(strings.TrimSpace(endMatch))
		}
		sourceIdx := len(pf.Sources) - 1
		if pf.State == InMultiQuote { // This line ends a multi-line quote.
			quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
			pf.Sources[sourceIdx].Quotes[quoteIdx].Body =
				append(pf.Sources[sourceIdx].Quotes[quoteIdx].Body, body)
			pf.Sources[sourceIdx].Quotes[quoteIdx].Page = page
		} else {
			pf.Sources[sourceIdx].Quotes =
				append(pf.Sources[sourceIdx].Quotes, Quote{Body: []string{body}, Page: page})
		}
		pf.State = InQuote
	}
	return isQuote
	// return newQuote(auth, kw, body, page, supp, note, url), isQuote
}

func parseMultiQuote(l string, pf *ParsedFile) bool {
	l = strings.TrimSpace(l)
	isMultiQuote := multiLineQuote.MatchString(l)
	if isMultiQuote {
		body := strings.TrimSpace(multiLineQuote.ReplaceAllLiteralString(l, ""))
		sourceIdx := len(pf.Sources) - 1
		//quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
		pf.Sources[sourceIdx].Quotes =
			append(pf.Sources[sourceIdx].Quotes, Quote{Body: []string{body}})
		pf.State = InMultiQuote
	}
	return isMultiQuote
}

// func parseNote(l Line) (string, bool) {
func parseNote(l string, pf *ParsedFile) bool {
	l = strings.TrimSpace(l)
	isNote := noteEnd.MatchString(l)
	if isNote {
		sourceIdx := len(pf.Sources) - 1
		quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
		pf.Sources[sourceIdx].Quotes[quoteIdx].Note = l
	}
	return isNote
}

// func parseURL(l Line) (string, bool) {
func parseURL(l string, pf *ParsedFile) bool {
	l = strings.TrimSpace(l)
	isURL := URLLine.MatchString(l)
	if isURL {
		// Don't remove http prefix from url.
		sourceIdx := len(pf.Sources) - 1
		quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
		pf.Sources[sourceIdx].Quotes[quoteIdx].URL = l
	}
	return isURL
}

// func parseAuth(l Line) (string, bool) {
func parseAuth(l string, pf *ParsedFile) bool {
	l = strings.TrimSpace(l)
	isAuth := authorLine.MatchString(l)
	if isAuth {
		body := strings.TrimSpace(authorLine.ReplaceAllString(l, ""))
		sourceIdx := len(pf.Sources) - 1
		quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
		pf.Sources[sourceIdx].Quotes[quoteIdx].Auth = body
	}
	return isAuth
}

// func parseKeyword(l Line) (string, bool) {
func parseKeyword(l string, pf *ParsedFile) bool {
	l = strings.TrimSpace(l)
	isKeyword := keywordLine.MatchString(l)
	if isKeyword {
		body := strings.TrimSpace(keywordLine.ReplaceAllString(l, ""))
		sourceIdx := len(pf.Sources) - 1
		quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
		pf.Sources[sourceIdx].Quotes[quoteIdx].Keyword = body
	}
	return isKeyword
}

// func parseSupplemental(l Line) (string, bool) {
func parseSupplemental(l string, pf *ParsedFile) bool {
	l = strings.TrimSpace(l)
	isSupp := supplementalLine.MatchString(l)
	if isSupp {
		body := strings.TrimSpace(supplementalLine.ReplaceAllString(l, ""))
		sourceIdx := len(pf.Sources) - 1
		quoteIdx := len(pf.Sources[sourceIdx].Quotes) - 1
		pf.Sources[sourceIdx].Quotes[quoteIdx].Supp =
			append(pf.Sources[sourceIdx].Quotes[quoteIdx].Supp, body)
	}
	return isSupp
}

// func parseQuote(q Line) (Quote, bool) {
// 	var auth, kw, page, note, url string
// 	var body, supp []string

// 	// Malformed page numbers are recorded using `pageUnknown`.
// 	const pageUnknown = "?"

// 	// Predominant Case: tab-delimited quote ends
// 	endMatchIndices := quoteEnd.FindStringIndex(q.Body)

// 	// Alternate Case: space-delimited quote ends
// 	if endMatchIndices == nil {
// 		endMatchIndices = quoteEndAlt.FindStringIndex(q.Body)
// 	}

// 	isQuote := endMatchIndices != nil

// 	if isQuote {
// 		// Split quote into body and end
// 		bodyMatch := q.Body[:endMatchIndices[0]]
// 		endMatch := q.Body[endMatchIndices[0]:]

// 		// Get quote body
// 		body = append(body, strings.TrimSpace(bodyMatch))
// 		// Split end into page and supplementary field
// 		pageMatchIndices := quotePage.FindStringIndex(endMatch)

// 		if pageMatchIndices == nil {
// 			// Unable to parse page number
// 			page = pageUnknown
// 		} else {
// 			page = pageNumber.FindString(strings.TrimSpace(endMatch))
// 		}
// 	}
// 	return newQuote(auth, kw, body, page, supp, note, url), isQuote
// }
