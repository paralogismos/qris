// qris.go
//
// Parse quote .txt files into .ris format.
//
// Assumptions:
//
//   The first line is the title of the source
//     - for use of the user in the quote file
//     - not used in creating the .ris file
//
//   The second line is a citation
//     - the citation should be parsed into name, year, raw-citation,
//       and a note-field
//       - notes preceding any quotes are assumed to be attached to the citation
//     - the citation line may optionally begin with "<$>" indicating a new source
//
//   Any line beginning with "<$>" is the citation line for a new source
//
//   Lines following a citation are quotes IF they end in a page number, followed
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
//   Blank lines are ignored
//
//   Other lines are written to a review file in the format:
//     - >[line #]
//       [discarded line]
//
// TODO:
//
// _ - Update record type in WriteQuotes: "TY  - Generic" -> ?
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
	"unicode/utf8"
)

// Definitions of system constants.
const Version = "v0.8.0"
const parsedSuffix = "_PARSED.ris"
const discardSuffix = "_DISCARDS.txt"
const configDir = "qris"
const configFile = "qris.conf"

// The first line of the file is assumed to be the source title.
// The first citation line may begin with the `sourceBegin` token.
// All subsequent citations must begin with the `sourceBegin` token.

// Parsed from the second line of the file  into family-name, year,
// and raw-citation.
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

// Parsed from a `Line` for which `IsQuote` is true.
// Includes line number from original file.
// Note that the original `Line` body should have at least a page number, and
// possibly a supplementary entry, associated with it. These have been extracted
// in a `Quote` and placed in their own fields.
type Quote struct {
	LineNo  int
	Auth    string
	Keyword string
	Body    string
	Page    string
	Supp    string
	Note    string
}

func newQuote(lineNo int, auth, kw, body, page, supp, note string) Quote {
	return Quote{
		LineNo:  lineNo,
		Auth:    auth,
		Keyword: kw,
		Body:    body,
		Page:    page,
		Supp:    supp,
		Note:    note,
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
	if IsDocFile(fpath) {
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

// `cleanLines` removes all empty lines, including whitespace lines,
// from a slice of `Line`s.
//
// ??? Strip leading and trailing whitespace from each `Line`?
func cleanLines(lines []Line) []Line {
	cls := []Line{}

	for _, l := range lines {
		if l.Body == "" {
			continue
		} else {
			cls = append(cls, l)
		}
	}

	return cls
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

func WriteQuotes(pf *ParsedFile, fname string, volume bool) {
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
	tstamp := time.Now().Format("2006/01/02")

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
			fmt.Fprintln(file, "AD  -", tstamp)
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
			fmt.Fprintln(file, "ER  -")
			fmt.Fprintln(file, "")
		}
	}
}

func ValidateUTF8(fpath string) bool {
	file, err := os.Open(fpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

// `WriteResults` iterates over a list of files, ensures that none are
// directories, parses each file,  and writes the results to output files.
func WriteResults(workPath string, dataList []string, volume bool) bool {
	allPassed := true // For UTF8 validation option
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
		WriteQuotes(&pf, pQuotes, volume)
	}

	return allPassed
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
