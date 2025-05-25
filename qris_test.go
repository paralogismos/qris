// qris_test.go

package qris

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestDetermineLineType(t *testing.T) {
	testCases := []struct {
		input    string
		ps       ParseState
		wantType LineType
	}{
		{
			input:    `This line should be discarded.`,
			ps:       InSource,
			wantType: UnknownLn,
		},
		{
			input:    `# This line is a comment before any sources.`,
			ps:       Start,
			wantType: CommentLn,
		},
		{
			input:    `# This line is a comment in a source.`,
			ps:       InSource,
			wantType: CommentLn,
		},
		{
			input: `Brown, Jason W. "Neuropsychology and the self-concept." ` +
				`The Journal of Nervous and Mental Disease. 187 no.3 (1999e): 131-41.`,
			ps:       Start,
			wantType: CitationLn,
		},
	}
	for n, tc := range testCases {
		lt := determineLineType(tc.input, tc.ps)
		if lt != tc.wantType {
			t.Errorf("failure in [%d]\n"+
				"found: %v\n"+
				"want: %v",
				n, lt, tc.wantType)
		}
	}
}

func TestGetSource(t *testing.T) {
	testCases := []struct {
		input    string
		wantName string
		wantYear string
		wantBody string
	}{
		{
			input: `Brown, Jason W. "Neuropsychology and the self-concept." ` +
				`The Journal of Nervous and Mental Disease. 187 no.3 (1999e): 131-41.`,
			wantName: "Brown, Jason W.",
			wantYear: "1999e",
			wantBody: `Brown, Jason W. "Neuropsychology and the self-concept." ` +
				`The Journal of Nervous and Mental Disease. 187 no.3 (1999e): 131-41.`,
		},
		{
			input: `Varela, Francisco J.` +
				` “Resonant cell assemblies: A new approach to cognitive functions and neuronal synchrony.”` +
				` {Biological Research}. 28 no.1 (1995): 81-95.`,
			wantName: "Varela, Francisco J.",
			wantYear: "1995",
			wantBody: `Varela, Francisco J.` +
				` “Resonant cell assemblies: A new approach to cognitive functions and neuronal synchrony.”` +
				` {Biological Research}. 28 no.1 (1995): 81-95.`,
		},
		{
			input: `Bermúdez, José Luis.` +
				` {Thinking without Words}. New York, NY: Oxford University Press, 2003.`,
			wantName: `Bermúdez, José Luis`,
			wantYear: `2003`,
			wantBody: `Bermúdez, José Luis. {Thinking without Words}. New York, NY: Oxford University Press, 2003.`,
		},
		{
			input:    `Gurwitsch, Aron. {Field of Consciousness}. Pittsburgh: Duquesne University Press, 1964 (1957).`,
			wantName: `Gurwitsch, Aron`,
			wantYear: `1957`,
			wantBody: `Gurwitsch, Aron. {Field of Consciousness}. Pittsburgh: Duquesne University Press, 1964 (1957).`,
		},
	}
	for n, tc := range testCases {
		src := getSource(tc.input)
		if src.Citation.Name != tc.wantName {
			t.Errorf("failure in Name of [%d]\n"+
				"Name = %s\n"+
				"want: %s",
				n, src.Citation.Name, tc.wantName)
		}
		if src.Citation.Year != tc.wantYear {
			t.Errorf("failure in Year of [%d]\n"+
				"Year = %s\n"+
				"want: %s",
				n, src.Citation.Year, tc.wantYear)
		}
		if src.Citation.Body != tc.wantBody {
			t.Errorf("failure in Body of [%d]\n"+
				"Body = %s\n"+
				"want: %s",
				n, src.Citation.Body, tc.wantBody)
		}
	}
}

func TestGetNote(t *testing.T) {
	qNoteLine := `page numbers are accoring to the pdf made by jmr --jmr`
	note := getNote(qNoteLine)
	want_note := `page numbers are accoring to the pdf made by jmr --jmr`
	if note != want_note {
		t.Fail()
	}
}

func TestGetQuote(t *testing.T) {
	testCases := []struct {
		input    string
		wantBody string
		wantPage string
	}{
		{
			input:    "Quote body followed by a single tab.\tp. 1 ",
			wantBody: `Quote body followed by a single tab.`,
			wantPage: `1`,
		},
		{
			input:    "Quote body followed by 3 spaces.   p. 2",
			wantBody: `Quote body followed by 3 spaces.`,
			wantPage: `2`,
		},
		{
			input:    "Quote body followed by 1 space and 1 tab. \tp. 3 ",
			wantBody: `Quote body followed by 1 space and 1 tab.`,
			wantPage: `3`,
		},
		{
			input:    "Quote body followed by page number and extra junk. \tp. 4  EXTRA JUNK HERE  ",
			wantBody: `Quote body followed by page number and extra junk.`,
			wantPage: `4`,
		},
		{
			input:    "Quote body followed by two page numbers.\tpp. 5, 6",
			wantBody: `Quote body followed by two page numbers.`,
			wantPage: `5, 6`,
		},
		{
			input:    "Quote body followed by two page numbers. \t pp. 10,11",
			wantBody: `Quote body followed by two page numbers.`,
			wantPage: `10,11`,
		},
		{
			input:    "Quote body followed by two page numbers or page range. \t\t pp. 100 101",
			wantBody: `Quote body followed by two page numbers or page range.`,
			wantPage: `100 101`,
		},
		{
			input:    "Quote body followed by page range. \t pp. 240-42",
			wantBody: `Quote body followed by page range.`,
			wantPage: `240-42`,
		},
		{
			input:    "Quote body followed by page range. \t pp. 240 - 42",
			wantBody: `Quote body followed by page range.`,
			wantPage: `240 - 42`,
		},
		{
			input:    "Quote body followed by page range. \t pp. 240--42",
			wantBody: `Quote body followed by page range.`,
			wantPage: `240--42`,
		},
		{
			input:    "Quote body followed by p 42. \t p 42",
			wantBody: `Quote body followed by p 42.`,
			wantPage: `42`,
		},
		{
			input:    "Quote body followed by pp 42, 43. \t p 42, 43",
			wantBody: `Quote body followed by pp 42, 43.`,
			wantPage: `42, 43`,
		},
		{
			input:    "Quote body followed by pp 42-43. \t pp  42-43",
			wantBody: `Quote body followed by pp 42-43.`,
			wantPage: `42-43`,
		},
	}
	for n, tc := range testCases {
		b, p := getQuote(tc.input)
		if b != tc.wantBody {
			t.Errorf("failure in Body of <exTestLines[%d]>\n"+
				"Body = %s\n\n"+
				" want: %s",
				n, b, tc.wantBody)
		}
		if p != tc.wantPage {
			t.Errorf("failure in Page of <exTestLines[%d]>\n"+
				"Page = %s\n want: %s\n",
				n, p, tc.wantPage)
		}
	}
}

// I may make some changes here:
// - handle multiple single test files
// - handle testing of batch processing files
// - I'm not sure that the cleanup process is ideal
//   - maybe results should not be deleted in the
//     case of a failing test
//   - would Cleanup be beneficial here?
func TestWriteResults(t *testing.T) {
	// Set Unix LF line endings for tests.
	LineEnding = "\n"
	workDir := GetWorkDir("")
	batchPath := "" // not testing batch mode
	testDir := "test_files"
	testFiles := []string{
		"test_descriptive_citations.docx",
		"test_example_citations.docx",
		"bib22e_FUNKY.docx",
		"24Brown1997_Qu.docx",
	}

	volume := false    // no volume information written
	dateStamp := false // no datestamp information written
	enc := Utf8        // write UTF-8 encoded output
	t.Chdir(testDir)   //
	for _, tf := range testFiles {
		dataList, workPath := GetWorkPath(workDir, batchPath, tf)
		// Process a test file.
		parsedFiles := ProcessQuoteFiles(workPath, dataList)
		// Write results to test directory.
		WriteResults(parsedFiles, OutOpts{Volume: volume, DateStamp: dateStamp, Encoding: enc})

		// Compare with expected results.
		resultPath := strings.TrimSuffix(tf, filepath.Ext(tf)) + "_PARSED.ris"
		discardPath := strings.TrimSuffix(tf, filepath.Ext(tf)) + "_DISCARD.txt"

		result, err := os.ReadFile(resultPath)
		if err != nil {
			t.Fatalf("%v: unable to open %s", err, resultPath)
		}

		wantPath := strings.TrimSuffix(tf, filepath.Ext(tf)) + "_EXPECT.ris"
		want, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatalf("%v: unable to open %s", err, wantPath)
		}

		if !slices.Equal(result, want) {
			t.Errorf("%s does not match %s\n", resultPath, wantPath)
		}

		_ = os.Remove(resultPath)
		_ = os.Remove(discardPath)
	}
}
