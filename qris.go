// qris.go
//
// Parse quote .txt files into .ris format.
//
// Assumptions:
//
//	 The first line is the title of the source.
//	   - for use of the user in the quote file
//	   - not used in creating the .ris file
//
//	 The second line is a citation.
//	   - the citation line should be parsed into citation author, year, raw-citation
//	   - a line following the citation is a citation note if it ends with
//	     "jmr" or "jmr."; the citation note is added to the citation information
//	     that is written for each quote from a source
//	   - the first citation line may optionally begin with "<$>" indicating a new source
//
//	 Any line beginning with "<$>" is the citation line for a new source.
//
//	 Lines following a citation are quotes IF they end in a page number.
//	   - quote lines are parsed into quote body and page number
//	   - a page number must be tab-delimited
//	   - a page number must follow one of the formats:
//	     - \t p. N
//	     - \t pp. N
//	     - the "p" indicator is case-insensitive
//	     - the . is optional
//	     - there MUST BE at least one space following p. or p
//	     - N is the page number group:
//	       - n : a number in [0-9,i, I, v, V, x, X, l, L, or ?]
//	       - n,n
//	       - n-n
//	       - n--n (for em dash transliteration)
//	       - commas and dashes may be surrounded by spaces, e.g., n -- n
//
//	A line following a citation that begins with "///" starts a multi-line quote
//	which ends with a page number.
//
//	A line following a quote that ends with "%%" attches a supplementary note.
//
//	A line following a quote that ends with "jmr" or "jmr." attaches a quote note.
//
//	A line following a quote that begins with "^S:" or "^s:" attaches a keyword
//	or a keyword list.
//
//	A line following a quote that begins with "https://" or "http://" attaches a URL.
//
//	A line following a quote that begins with ">>>" specifies a quote author.
//	  - if a quote author is specified, this name is attached as the primary author
//	    of the quote and the citation author is attached as the secondary author
//
//	 Blank lines are ignored
//
//	 Other lines are written to a review file in the format:
//	   - >[line #]
//	     [discarded line]
//
// TODO:
//
//   - `qris [path]` should do something reasonable
//
//   - currently this prints version information, but that is confusing:
//
//   - or at least print a message so that the user knows that nothing was processed
//
//   - need to think about these issues more....
//
//   - Can the `-b` flag be modified so that `-b` is used instead of `-b .`
//     to process all files in the qris working directory?
//
//   - then `-b .` could indicate processing all files in the current working
//     directory
//
//   - or: maybe the entire working directory idea should be scrapped...?
//
//   - I think that many of the calls to `ReplaceAllString` could be replaced
//     by `ReplaceAllLiteralString`.
//
//   - Can I modify so that _all_ citations must begin with "<$>"?
//
//   - not now: maybe in the future
//
//   - thus, the first line after the title line would no longer be
//     automatically considered a citation
//
//   - this would allow the user to include any number of descriptive lines
//     before source information in the input file
//
//   - this preliminary descriptive information could be collected
//     in the DISCARDS file
//
//   - Should DISCARDS output be optional?
//
//   - It makes more sense to me that %% should _precede_ a supplementary note
//
//   - then all markup comes at the beginning of a line
//
//   - except quote notes which end in jmr
//
//   - and quotes which end in page numbers
//
//   - The input file title field `tit` is not being used.
//
//   - Can I remove this?
//
//   - The title line could be sent to the _DISCARDS.ris file.
//
//   - Should I move `Line` from `fetch.go` back into this file?
//
// _ - Add functionality to store up to N notes following a quote.
//   - Store notes in a slice or dictionary
//   - Map notes array to numbered tags, e.g., C1, C2, ....
//   - Need to discuss this with Jack
//
// _ - GetConfigPath should perhaps create a config file if none exists.
//   - This file would contain the default path to a working directory.
//
// _ - Move functionality to get the working directory into a function.
//   - GetWorkDir
//   - Once this is established, the part that sets up a default working
//     directory could be moved into GetConfigPath.
//
// _ - Make eligible system constants unexported as soon as possible.
package qris

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf16"
)

type Encoding int

const (
	None Encoding = iota
	Ascii
	Ansi
	Utf8
	Utf16
)

type OutOpts struct {
	Volume    bool
	DateStamp bool
	Encoding  Encoding
}

// The first line of the file is assumed to be the source title.
// The first citation line may begin with the `sourceBegin` token.
// All subsequent citations must begin with the `sourceBegin` token.

// Parsed from the second line of the file into name, year, body. The note
// field me be supplied when subsequent file lines are parsed.
type Citation struct {
	Name string
	Year string
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

// Parsed from a `Line` for which `IsQuote` is true, or from the `Line`s of a
// multi-line quote. Includes line number from original file.
// Body and page are parsed from the lines of a quote. Other fields are supplied
// as lines are processed.
type Quote struct {
	Auth    string
	Keyword string
	Body    []string
	Page    string
	Supp    []string
	Note    string
	URL     string
}

func newQuote(auth, kw string, body []string, page string, supp []string, note, url string) Quote {
	return Quote{
		Auth:    auth,
		Keyword: kw,
		Body:    body,
		Page:    page,
		Supp:    supp,
		Note:    note,
		URL:     url,
	}
}

// A file may include multiple sources.
type Source struct {
	Citation Citation
	Quotes   []Quote
}

func newSource(cit Citation, qs []Quote) Source {
	return Source{
		Citation: cit,
		Quotes:   qs,
	}
}

// Results of parsing one file.
// `State` is initially `Start`, passing through other `ParseState`s during
// processing. The `State` is set to `Finished` after processing is completed.
// `Discards` is a slice of `Line`s which were not recognized. These can be
// reviewed manually by the user.
type ParsedFile struct {
	Filepath string // full filepath
	State    ParseState
	Sources  []Source
	Discards []Line
}

// func newParsedFile(fp string, ss []Source, ds []Line) ParsedFile {
// 	return ParsedFile{
// 		Filepath: fp,
// 		Sources:  ss,
// 		Discards: ds,
// 	}
// }

// `getLines` takes a file specified by `fpath` and returns a slice
// containing raw lines from the file keyed by the original line number
// on which each line occurred.
func getLines(fpath string) []Line {
	rawLines := []Line{}
	var err error
	if IsDocxFile(fpath) {
		rawLines, err = DocxToLines(fpath)
	} else { // Assume that `fpath` leads to a .txt file.
		rawLines, err = TxtToLines(fpath)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return rawLines
}

func WriteDiscards(ds []Line, fname string) {
	file, err := os.Create(fname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	for _, d := range ds {
		fmt.Fprintln(file, "<", d.LineNo, ">")
		fmt.Fprintln(file, d.Body)
	}
}

func writeFieldToFile(f *os.File, field string, data string, enc Encoding) {
	line := field + "  - " + data + LineEnding
	writeToFile(f, line, enc)
}

func writeToFile(f *os.File, data string, enc Encoding) {
	var mapping map[rune]string
	switch enc {
	case Utf16:
		writeToFileUtf16(f, data)
		return // write utf16 and early return
	case Utf8:
		mapping = nil
	case Ascii:
		mapping = utf8ToAscii()
	case Ansi:
		mapping = utf8ToAnsi()
	default:
		mapping = utf8ToAnsi()
	}
	fmt.Fprint(f, utf8ToNormalized(data, mapping))
}

func writeToFileUtf16(f *os.File, data string) {
	runes := []rune(data)
	codePoints := utf16.Encode(runes) // convert runes to utf-16
	binary.Write(f, binary.NativeEndian, codePoints)
}

func WriteQuotes(pf ParsedFile, fname string, outOpts OutOpts) {
	file, err := os.Create(fname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	// Use encoding:
	enc := outOpts.Encoding

	// batch ID
	bid := filepath.Base(filepath.Dir(fname))

	// file ID
	fid := filepath.Base(pf.Filepath)
	fid = strings.TrimSuffix(fid, filepath.Ext(fid))

	// timestamp: when file was processed
	dStamp := time.Now().Format("2006/01/02")

	// Start file with a blank line per RIS specification.
	writeToFile(file, LineEnding, enc)

	for _, s := range pf.Sources { // loop over sources of the parsed file
		citBody := s.Citation.Body
		citName := s.Citation.Name
		citYear := s.Citation.Year
		citNote := s.Citation.Note

		for _, q := range s.Quotes { // loop over quotes of each source
			writeFieldToFile(file, "TY", "", enc)
			if outOpts.Volume {
				writeFieldToFile(file, "VL", bid, enc)
			}
			writeFieldToFile(file, "UR", fid, enc)
			if outOpts.DateStamp {
				writeFieldToFile(file, "AD", dStamp, enc)
			}
			writeFieldToFile(file, "AB", citBody, enc)

			// A1 gets citation name unless a primary quote author was specified
			if q.Auth != "" {
				writeFieldToFile(file, "A1", q.Auth, enc)
				familyName := citationFamilyName.FindString(citName)
				familyName = strings.TrimSpace(familyName)
				writeFieldToFile(file, "A2", "in "+familyName, enc)
			} else {
				writeFieldToFile(file, "A1", citName, enc)
			}

			if citYear != "" {
				writeFieldToFile(file, "Y1", citYear, enc)
			}
			if citNote != "" {
				writeFieldToFile(file, "T2", citNote, enc)
			}
			if q.Keyword != "" {
				writeFieldToFile(file, "KW", q.Keyword, enc)
			}
			if q.Body != nil {
				for _, line := range q.Body {
					writeFieldToFile(file, "T1", line, enc)
				}
			}
			if q.Page != "" {
				writeFieldToFile(file, "SP", q.Page, enc)
			}
			if q.Supp != nil {
				for _, supp := range q.Supp {
					writeFieldToFile(file, "PB", supp, enc)
				}
			}
			if q.Note != "" {
				writeFieldToFile(file, "CY", q.Note, enc)
			}
			if q.URL != "" {
				writeFieldToFile(file, "UR", q.URL, enc)
			}
			writeFieldToFile(file, "ER", "", enc)
			writeToFile(file, LineEnding, enc)
		}
	}
}

// `ProcessQuoteFiles` iterates over a list of files and returns
// a list of `ParsedFile`s.
func ProcessQuoteFiles(workPath string, dataList []string) []ParsedFile {
	var parsedFiles []ParsedFile
	for _, file := range dataList {
		// Don't process parsed file artifacts.
		if isParsedFile(file) || isDiscardFile(file) {
			continue
		}
		fmt.Printf("Processing %s...\n", file) // Display file name as it is processed
		pFile := filepath.Join(workPath, file) // File path to process
		//parsedFiles = append(parsedFiles, ParseFile(pFile))
		parsedFiles = append(parsedFiles, ProcessFile(pFile))
	}
	return parsedFiles
}

// `WriteResults` iterates over a list of files, ensures that none are
// directories, parses each file,  and writes the results to output files.
func WriteResults(parsedFiles []ParsedFile, outOpts OutOpts) {
	for _, pf := range parsedFiles {
		fpath := pf.Filepath
		base := strings.TrimSuffix(fpath, filepath.Ext(fpath))
		pQuotes := base + parsedSuffix // File to store parsed quotes

		WriteQuotes(pf, pQuotes, outOpts)

		// Only write a _DISCARD file if there were discarded lines.
		if len(pf.Discards) > 0 {
			pDiscard := base + discardSuffix // File to store discarded lines
			WriteDiscards(pf.Discards, pDiscard)
		}
	}
}
