// qris.go
//
// Parse quote .txt files into .ris format.
//
// Assumptions:
//
//   The first line is the title of the source.
//     - for use of the user in the quote file
//     - not used in creating the .ris file
//
//   The second line is a citation.
//     - the citation line should be parsed into citation author, year, raw-citation
//     - a line following the citation is a citation note if it ends with
//       "jmr" or "jmr."; the citation note is added to the citation information
//       that is written for each quote from a source
//     - the first citation line may optionally begin with "<$>" indicating a new source
//
//   Any line beginning with "<$>" is the citation line for a new source.
//
//   Lines following a citation are quotes IF they end in a page number.
//     - quote lines are parsed into quote body and page number
//     - a page number must be tab-delimited
//     - a page number must follow one of the formats:
//       - \t p. N
//       - \t pp. N
//       - the "p" indicator is case-insensitive
//       - the . is optional
//       - there MUST BE at least one space following p. or p
//       - N is the page number group:
//         - n : a number in [0-9,i, I, v, V, x, X, l, L, or ?]
//         - n,n
//         - n-n
//         - n--n (for em dash transliteration)
//         - commas and dashes may be surrounded by spaces, e.g., n -- n
//
//  A line following a citation that begins with "///" starts a multi-line quote
//  which ends with a page number.
//
//  A line following a quote that ends with "%%" attches a supplementary note.
//
//  A line following a quote that ends with "jmr" or "jmr." attaches a quote note.
//
//  A line following a quote that begins with "^S:" or "^s:" attaches a keyword
//  or a keyword list.
//
//  A line following a quote that begins with "https://" or "http://" attaches a URL.
//
//  A line following a quote that begins with ">>>" specifies a quote author.
//    - if a quote author is specified, this name is attached as the primary author
//      of the quote and the citation author is attached as the secondary author
//
//   Blank lines are ignored
//
//   Other lines are written to a review file in the format:
//     - >[line #]
//       [discarded line]
//
// TODO:
//
//   - Explore better TUI interface:
//     - ability to work on files directly in the shell working directory
//       - note that the `-b` flag accepts the `.` argument to process all files
//         in the current directory: this should be enough....
//     - `qris [path]` should do something reasonable
//       - currently this prints version information, but that is confusing:
//         - `qris -f [path]` processes a file
//         - `qris` prints version information
//         - `qris [path]` seems like it ought to process a file
//         - or at least print a message so that the user knows that nothing was processed
//         - need to think about these issues more....
//     - Can the `-b` flag be modified so that `-b` is used instead of `-b .`
//       to process all files in the qris working directory?
//       - then `-b .` could indicate processing all files in the current working
//         directory
//       - or: maybe the entire working directory idea should be scrapped...?
//   - Try to handle .doc files:
//     - currently have "zip: not a valid zip file" failing error
//     - do .doc files have the same format as .docx, but without zip compression?
//   - Update unit tests.
//   - Write integration tests.
//   - Review DISCARDS file:
//     - .docx -> .ris file lines have newlines between each line
//     - .txt  -> .ris file lines do not have the extra lines
//   - I think that many of the calls to `ReplaceAllString` could be replaced
//     by `ReplaceAllLiteralString`.
//
//   - Can I modify so that _all_ citations must begin with "<$>"?
//     - thus, the first line after the title line would no longer be
//       automatically considered a citation
//     - this would allow the user to include any number of descriptive lines
//       before source information in the input file
//     - this preliminary descriptive information could be collected
//       in the DISCARDS file
//   - Should DISCARDS output be optional?
//   - Would the user like to preserve leading whitespace in multi-line quotes?
//     - note that this would require preserving whitespace for both the
//       intermediate lines of the multi-line quote, and for the final line
//       which is a simple quote line ending with a page number; this could
//       be accomplished by checking the `inMultiLineQuote` boolean
//   - It makes more sense to me that %% should _precede_ a supplementary note
//     - then all markup comes at the beginning of a line
//     - except quote notes which end in jmr
//     - and quotes which end in page numbers
//       - note that multi-line quotes can end with the page on a separate line
//   - Review trimming of whitespace:
//     - when and where does it occur?
//     - when and where should it occur?
//     - make this more methodical and consistent
//     - whitespace should not have to be trimmed in tests (as it is now)
//   - The input file title field `tit` is not being used.
//     - Can I remove this?
//     - The title line could be sent to the _DISCARDS.ris file.
//
//   - Should I move `Line` from `fetch.go` back into this file?
// _ - Add functionality to store up to N notes following a quote.
//     - Store notes in a slice or dictionary
//     - Map notes array to numbered tags, e.g., C1, C2, ....
//     - Need to discuss this with Jack
// _ - GetConfigPath should perhaps create a config file if none exists.
//     - This file would contain the default path to a working directory.
// _ - Move functionality to get the working directory into a function.
//     - GetWorkDir
//     - Once this is established, the part that sets up a default working
//       directory could be moved into GetConfigPath.
// _ - Make eligible system constants unexported as soon as possible.
//

package qris

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Definitions of system constants.
const Version = "v0.12.1"
const parsedSuffix = "_PARSED.ris"
const discardSuffix = "_DISCARD.txt"
const configDir = "qris"
const configFile = "qris.conf"

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
	Body    string
	Page    string
	Supp    string
	Note    string
	URL     string
}

func newQuote(auth, kw, body, page, supp, note, url string) Quote {
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
// `Discards` is a slice of `Line`s which aren't quotes, to be reviewed manually
// by the user.
type ParsedFile struct {
	Filename string
	Title    string // first line of parsed file
	Sources  []Source
	Discards []Line
}

func newParsedFile(fn, tit string, srcs []Source, ds []Line) ParsedFile {
	return ParsedFile{
		Filename: fn,
		Title:    tit,
		Sources:  srcs,
		Discards: ds,
	}
}

// `GetConfigPath` checks for a configuration directory and
// creates one if none exists.
func GetConfigPath() string {
	configPath := ""
	userConfig, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		configDirPath := filepath.Join(userConfig, configDir)
		_, err := os.ReadDir(configDirPath)
		if err != nil {
			err := os.Mkdir(configDirPath, 0666)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		configPath = filepath.Join(userConfig, configDir, configFile)
	}
	return configPath
}

// `GetWorkDir` looks in `configPath` for a configuration file. If one exists,
// the configured working directory is returned. Otherwise `os.Getwd()` is used
// to get the current working directory from the system and this path is returned.
func GetWorkDir(configPath string) string {
	workDir := ""
	config, err := os.Open(configPath)
	if err == nil {
		defer config.Close()
		scanner := bufio.NewScanner(config)
		if scanner.Scan() {
			workDir = scanner.Text()
			if err := os.Chdir(workDir); err != nil {
				fmt.Fprintln(os.Stderr, err)
				workDir, err = os.Getwd()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}
	} else {
		workDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	return workDir
}

// `SetWorkDir` stores a new working directory path in the configuration file
// and sets the current working directory on the users system.
// If `SetWorkDir` cannot set the working directory for some reason, the
// default working directory already established by the system and discovered
// by `GetWorkDir` should be available for the user to use.
func SetWorkDir(dirPath, configPath string) {
	if dirPath != "" {
		workDir, err := filepath.Abs(dirPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.Chdir(workDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		config, err := os.Create(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			defer config.Close()
			fmt.Fprintln(config, workDir)
		}
	}
}

// *** I don't think that this function is being used anymore....
// *** Maybe this should be removed.
// `GetFileList` creates a list of filenames from the text file specified by `fpath`.
// This file should have one filename per line.
func GetFileList(fpath string) []string {
	file, err := os.Open(fpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	files := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}

	return files
}

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

// Converts problematic unicode characters to reliable characters.
func tidyString(l string) string {
	conversions := map[rune]string{
		'“': `"`, '”': `"`, '‘': `'`, '’': `'`,
		'–': `-`, '—': `--`, '…': `...`,
		'«': `<<`, '»': `>>`, '†': ``,
		'À': `A`, 'È': `E`, 'Ì': `I`, 'Ò': `O`, 'Ù': `U`,
		'à': `a`, 'è': `e`, 'ì': `i`, 'ò': `o`, 'ù': `u`,
		'Á': `A`, 'É': `E`, 'Í': `I`, 'Ó': `O`, 'Ú': `U`, 'Ý': `Y`,
		'á': `a`, 'é': `e`, 'í': `i`, 'ó': `o`, 'ú': `u`, 'ý': `y`,
		'Â': `A`, 'Ê': `E`, 'Î': `I`, 'Ô': `O`, 'Û': `U`,
		'â': `a`, 'ê': `e`, 'î': `i`, 'ô': `o`, 'û': `u`,
		'Ã': `A`, 'Ñ': `N`, 'Õ': `O`,
		'ã': `a`, 'ñ': `n`, 'õ': `o`,
		'Ä': `A`, 'Ë': `E`, 'Ï': `I`, 'Ö': `O`, 'Ü': `U`, 'Ÿ': `Y`,
		'ä': `a`, 'ë': `e`, 'ï': `i`, 'ö': `o`, 'ü': `u`, 'ÿ': `y`,
		'Æ': `ae`, 'Œ': `OE`,
		'æ': `ae`, 'œ': `oe`,
		'Ç': `C`,
		'ç': `c`,
	}

	tidyResult := ""
	rs := []rune(l)
	for _, r := range rs { // check for runes to be replaced by a single rune
		if newRunes, replace := conversions[r]; replace {
			tidyResult += newRunes
		} else {
			tidyResult += string(r)
		}
	}

	return tidyResult
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

func WriteQuotes(pf *ParsedFile, fname string, volume bool, dateStamp bool) {
	file, err := os.Create(fname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	// batch ID
	bid := filepath.Base(filepath.Dir(fname))

	// file ID
	fid := strings.TrimSuffix(pf.Filename, filepath.Ext(pf.Filename))

	// timestamp: when file was processed
	dStamp := time.Now().Format("2006/01/02")

	for _, s := range pf.Sources { // loop over sources of the parsed file
		citBody := s.Citation.Body
		citName := s.Citation.Name
		citYear := s.Citation.Year
		citNote := s.Citation.Note

		for _, q := range s.Quotes { // loop over quotes of each source
			fmt.Fprintln(file, "TY  - ")
			if volume {
				fmt.Fprintln(file, "VL  -", bid)
			}
			fmt.Fprintln(file, "UR  -", fid)
			if dateStamp {
				fmt.Fprintln(file, "AD  -", dStamp)
			}
			fmt.Fprintln(file, "AB  -", citBody)

			// A1 gets citation name unless a primary quote author was specified
			if q.Auth != "" {
				fmt.Fprintln(file, "A1  -", q.Auth)
				familyName := citationFamilyName.FindString(citName)
				familyName = strings.TrimSpace(familyName)
				fmt.Fprintln(file, "A2  -", familyName)
			} else {
				fmt.Fprintln(file, "A1  -", citName)
			}

			if citYear != "" {
				fmt.Fprintln(file, "Y1  -", citYear)
			}
			if citNote != "" {
				fmt.Fprintln(file, "T2  -", citNote)
			}
			if q.Keyword != "" {
				fmt.Fprintln(file, "KW  -", q.Keyword)
			}
			if q.Body != "" {
				fmt.Fprintln(file, "T1  -", q.Body)
			}
			if q.Page != "" {
				fmt.Fprintln(file, "SP  -", q.Page)
			}
			if q.Supp != "" {
				fmt.Fprintln(file, "PB  -", q.Supp)
			}
			if q.Note != "" {
				fmt.Fprintln(file, "CY  -", q.Note)
			}
			if q.URL != "" {
				fmt.Fprintln(file, "UR  -", q.URL)
			}
			fmt.Fprintln(file, "ER  -")
			fmt.Fprintln(file, "")
		}
	}
}

// `WriteResults` iterates over a list of files, ensures that none are
// directories, parses each file,  and writes the results to output files.
func WriteResults(workPath string, dataList []string, volume bool, dateStamp bool) {
	for _, file := range dataList {
		// Don't process parsed file artifacts.
		if isParsedFile(file) || isDiscardFile(file) {
			continue
		}

		// Display file name as it is processed
		fmt.Println(file)

		// File path to process
		pFile := filepath.Join(workPath, file)

		// File to store parsed quotes
		base := strings.TrimSuffix(pFile, filepath.Ext(pFile))
		pQuotes := base + parsedSuffix

		// File to store discarded lines
		pDiscard := base + discardSuffix

		pf := ParseFile(pFile)
		WriteDiscards(pf.Discards, pDiscard)
		WriteQuotes(&pf, pQuotes, volume, dateStamp)
	}
}

// `GetBatchList` takes a path argument and returns a list of all files found
// in the directory specified by the path. Directories found in the specified
// directory are not included in the list. It is an error if the path does not
// lead to a directory.
func GetBatchList(path string) []string {
	var dataList []string
	dirEnts, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Create list of files in specified directory.
	for _, dirEnt := range dirEnts {
		if dirEnt.IsDir() {
			continue
		}
		dataList = append(dataList, dirEnt.Name())
	}

	return dataList
}

// `GetWorkPath` takes as arguments an absolute path to the current working
// directory, a path to a batch directory to be processed (`bFlag`), and a path
// to a file to be processed (`fFlag`). The second two paths are relative to
// the current working directory. This information is used to create a list
// of files for processing which is returned to the caller.
func GetWorkPath(workDir, bFlag, fFlag string) ([]string, string) {
	var dList []string
	var wPath string
	if bFlag == "" {
		if fFlag != "" {
			// Add a single file to `dataList` if one was supplied.
			wPath, _ = filepath.Abs(fFlag)
			var wFile string
			wPath, wFile = filepath.Split(wPath)
			dList = append(dList, wFile)
		}
	} else {
		// Batch process files.
		// Allow dot argument to indicate batch files found in working directory.
		if bFlag == "." {
			wPath = workDir
		} else {
			wPath, _ = filepath.Abs(bFlag)
		}
		dList = GetBatchList(wPath)
	}
	return dList, wPath
}
