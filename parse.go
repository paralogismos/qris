// parse.go
//
// Regular expressions used by the `qris` package.
package qris

import (
	"regexp"
	"strings"
)

// Markup Tokens
var blankLine = regexp.MustCompile(`^[\p{Zs}\t]*$`)
var commentLine = regexp.MustCompile(`^##`)
var discardLine = regexp.MustCompile(`^__`)
var citationLine = regexp.MustCompile(`^<\$>`)
var citationNoteLine = regexp.MustCompile(`-nb\.?$`)
var mlqLine = regexp.MustCompile(`^///`)
var quoteNoteLine = regexp.MustCompile(`jmr\.?$`)
var quoteAuthorLine = regexp.MustCompile(`^>>>`)
var supplementLine = regexp.MustCompile(`%%$`)
var keywordLine = regexp.MustCompile(`^\^[sS]:`)
var urlLine = regexp.MustCompile(`^https?://`)

// A quote end is either tab-delimited pp., or space-delimited pp. with
// at least three spaces as the delimiter.
var quoteLine = regexp.MustCompile(`\t\p{Zs}*[pP]{1,2}\.?\p{Zs}+[\pNiIvVxXlL\?]*.*`)
var quoteLineAlt = regexp.MustCompile(`\p{Zs}{3,}?[pP]{1,2}\.?\p{Zs}+[\pNiIvVxXlL\?]*.*`)

// Citation Line Information
var citationName = regexp.MustCompile(`^[^“"{]*`)
var finalPeriod = regexp.MustCompile(`\.$`)
var nameInitialPeriod = regexp.MustCompile(`[\p{Zs}.]{1}[\pL]{1}\.$`)
var citationFamilyName = regexp.MustCompile(`^[^,]*`)
var citationYear = regexp.MustCompile(`[^\pN]\pN{4}\pL?[^\pN]`)
var citationYearAlt = regexp.MustCompile(`[^\pN]\pN{4}\pL?[\pP]?$`)
var citationYearTrim = regexp.MustCompile(`\pN{4}\pL?`)

// Quote Line Information
var quotePage = regexp.MustCompile(
	`[pP]{1,2}\.?\p{Zs}*[\pNiIvVxXlL\?]+[f]{0,2}\p{Zs}*[,\pPd]*\p{Zs}*[\pNiIvVxXlL\?]*[f]{0,2}`)
var pageNumber = regexp.MustCompile(
	`[\pNiIvVxXlL\?]+[f]{0,2}\p{Zs}*[,\pPd]*\p{Zs}*[\pNiIvVxXlL\?]*[f]{0,2}`)

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
	UnknownLn LineType = iota
	DiscardLn
	BlankLn
	CommentLn
	CitationLn
	CitationNoteLn
	QuoteLn
	MultiQuoteLn
	QuoteNoteLn
	QuoteAuthorLn
	KeywordLn
	SupplementLn
	UrlLn
)

func (lt LineType) String() string {
	var s string
	switch lt {
	case UnknownLn:
		s = "UnknownLn"
	case DiscardLn:
		s = "DiscardLn"
	case BlankLn:
		s = "BlankLn"
	case CommentLn:
		s = "CommentLn"
	case CitationLn:
		s = "CitationLn"
	case CitationNoteLn:
		s = "CitationNoteLn"
	case QuoteLn:
		s = "QuoteLn"
	case MultiQuoteLn:
		s = "MultiQuoteLn"
	case QuoteNoteLn:
		s = "QuoteNoteLn"
	case QuoteAuthorLn:
		s = "QuoteAuthorLn"
	case KeywordLn:
		s = "KeywordLn"
	case SupplementLn:
		s = "SupplementLn"
	case UrlLn:
		s = "UrlLn"
	}
	return s
}

func determineLineType(body string, ps ParseState) LineType {
	switch {
	case blankLine.MatchString(body) && ps != InMultiQuote:
		return BlankLn
	case commentLine.MatchString(body):
		return CommentLn
	case discardLine.MatchString(body):
		return DiscardLn
	case citationLine.MatchString(body) || ps == Start:
		return CitationLn
	case citationNoteLine.FindStringIndex(body) != nil:
		return CitationNoteLn
	case quoteLine.FindStringIndex(body) != nil ||
		quoteLineAlt.FindStringIndex(body) != nil:
		return QuoteLn
	case mlqLine.MatchString(body):
		return MultiQuoteLn
	case quoteNoteLine.MatchString(body):
		return QuoteNoteLn
	case quoteAuthorLine.MatchString(body):
		return QuoteAuthorLn
	case keywordLine.MatchString(body):
		return KeywordLn
	case supplementLine.MatchString(body):
		return SupplementLn
	case urlLine.MatchString(body):
		return UrlLn
	default:
		return UnknownLn
	}
}

func ProcessFile(fpath string) ParsedFile {
	var pf ParsedFile
	curSrc := -1 // No sources yet.
	curQte := -1 // No quotes yet.
	pf.Filepath = fpath
	pf.State = Start
	rls := getLines(fpath)
	for _, l := range rls[1:] { // Always ignore first line of input file.
		body := strings.TrimSpace(l.Body)
		lineType := determineLineType(body, pf.State)
		if lineType == CommentLn || lineType == BlankLn { // Skipped lines.
			continue
		}
		if lineType == DiscardLn ||
			(lineType == UnknownLn && pf.State != InMultiQuote) {
			pf.Discards = append(pf.Discards, l)
		}
		switch pf.State {
		case Start:
			if lineType == CitationLn {
				pf.Sources = append(pf.Sources, getSource(body))
				curSrc += 1 // Added a source.
				curQte = -1 // No quotes yet.
				pf.State = InSource
				break
			}
		case InSource:
			if lineType == CitationNoteLn {
				pf.Sources[curSrc].Citation.Note = getCitationNote(body)
				break
			}
			if lineType == QuoteLn {
				b, p := getQuote(body)
				pf.Sources[curSrc].Quotes =
					append(pf.Sources[curSrc].Quotes, Quote{Body: []string{b}, Page: p})
				curQte += 1 // Added a quote.
				pf.State = InQuote
				break
			}
			if lineType == MultiQuoteLn {
				b := beginMultiQuote(body)
				pf.Sources[curSrc].Quotes =
					append(pf.Sources[curSrc].Quotes, Quote{Body: []string{b}})
				curQte += 1 // Added a quote.
				pf.State = InMultiQuote
				break
			}
		case InMultiQuote:
			if lineType == QuoteLn { // This line ends a multi-line quote.
				b, p := getQuote(body)
				pf.Sources[curSrc].Quotes[curQte].Body =
					append(pf.Sources[curSrc].Quotes[curQte].Body, b)
				pf.Sources[curSrc].Quotes[curQte].Page = p
				pf.State = InQuote
				break
			}
			// All non-skipped, non-single-quote lines are captured.
			pf.Sources[curSrc].Quotes[curQte].Body =
				append(pf.Sources[curSrc].Quotes[curQte].Body, strings.TrimSpace(body))
		case InQuote:
			if lineType == CitationLn {
				pf.Sources = append(pf.Sources, getSource(body))
				curSrc += 1 // Added a source.
				curQte = -1 // No quotes yet.
				pf.State = InSource
				break
			}
			if lineType == QuoteLn {
				b, p := getQuote(body)
				pf.Sources[curSrc].Quotes =
					append(pf.Sources[curSrc].Quotes, Quote{Body: []string{b}, Page: p})
				curQte += 1 // Added a quote.
				pf.State = InQuote
				break
			}
			if lineType == MultiQuoteLn {
				b := beginMultiQuote(body)
				pf.Sources[curSrc].Quotes =
					append(pf.Sources[curSrc].Quotes, Quote{Body: []string{b}})
				curQte += 1 // Added a quote.
				pf.State = InMultiQuote
				break
			}
			if lineType == QuoteNoteLn {
				pf.Sources[curSrc].Quotes[curQte].Note = getNote(body)
			}
			if lineType == QuoteAuthorLn {
				pf.Sources[curSrc].Quotes[curQte].Auth = getQuoteAuthor(body)
			}
			if lineType == KeywordLn {
				pf.Sources[curSrc].Quotes[curQte].Keyword = getKeyword(body)
			}
			if lineType == SupplementLn {
				pf.Sources[curSrc].Quotes[curQte].Supp =
					append(pf.Sources[curSrc].Quotes[curQte].Supp, getSupplement(body))
			}
			if lineType == UrlLn {
				pf.Sources[curSrc].Quotes[curQte].Url = getUrl(body)
			}
		default: // Unrecognized state: discard line for review.
			pf.Discards = append(pf.Discards, l)
		}
	}
	pf.State = Finished
	return pf
}

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

func getSource(b string) Source {
	tb := citationLine.ReplaceAllString(b, "") // trim `citationLine` token
	tb = strings.TrimSpace(tb)
	name := strings.TrimSpace(citationName.FindString(tb))
	if !nameInitialPeriod.MatchString(name) {
		name = finalPeriod.ReplaceAllString(name, "")
	}
	year := citationYearAlt.FindString(tb) // year at end of citation
	if year == "" {
		yearMatches := citationYear.FindAllString(tb, -1) // year buried in citation
		countMatches := len(yearMatches)
		if countMatches > 0 {
			year = yearMatches[countMatches-1]
		}
	}
	year = citationYearTrim.FindString(year)
	cit := Citation{
		Name: name,
		Year: year,
		Body: tb,
	}
	src := Source{Citation: cit}
	return src
}

func getCitationNote(b string) string {
	return strings.TrimSpace(citationNoteLine.ReplaceAllLiteralString(b, ""))
}

func getQuote(b string) (string, string) {
	// Malformed page numbers are recorded using `pageUnknown`.
	const pageUnknown = "?"
	var page string
	endMatchIndices := quoteLine.FindStringIndex(b) // Tab-delimited quote ends.
	if endMatchIndices == nil {                     // Alternate Case: space-delimited quote ends.
		endMatchIndices = quoteLineAlt.FindStringIndex(b)
	}
	// Split quote into body and end
	bodyMatch := b[:endMatchIndices[0]]
	endMatch := b[endMatchIndices[0]:]

	// Get quote body: single-line quote may begin with multi-quote token.
	body := strings.TrimSpace(mlqLine.ReplaceAllLiteralString(bodyMatch, ""))

	// Split end into page and supplementary field
	pageMatchIndices := quotePage.FindStringIndex(endMatch)

	if pageMatchIndices == nil { // Unable to parse page number
		page = pageUnknown
	} else {
		page = strings.TrimSpace(pageNumber.FindString(strings.TrimSpace(endMatch)))
	}
	return body, page
}

func beginMultiQuote(b string) string {
	return strings.TrimSpace(mlqLine.ReplaceAllLiteralString(b, ""))
}

func getNote(b string) string {
	return b
}

func getQuoteAuthor(b string) string {
	return strings.TrimSpace(quoteAuthorLine.ReplaceAllString(b, ""))
}

func getKeyword(b string) string {
	return strings.TrimSpace(keywordLine.ReplaceAllString(b, ""))
}

func getSupplement(b string) string {
	return strings.TrimSpace(supplementLine.ReplaceAllString(b, ""))
}

func getUrl(b string) string {
	return b // Don't remove http prefix from url.
}
