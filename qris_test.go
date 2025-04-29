// qris_test.go

package qris

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCitation(t *testing.T) {
	ex_citation := newLine(0,
		`Brown, Jason W. "Neuropsychology and the self-concept." `+
			`The Journal of Nervous and Mental Disease. 187 no.3 (1999e): 131-41.`)
	cit := parseCitation(ex_citation)
	want_name := "Brown, Jason W."
	want_year := "1999e"
	want_body := `Brown, Jason W. "Neuropsychology and the self-concept." ` +
		`The Journal of Nervous and Mental Disease. 187 no.3 (1999e): 131-41.`
	if cit.Name != want_name ||
		cit.Year != want_year ||
		cit.Body != want_body {
		t.Fail()
	}
}

func TestParseNote(t *testing.T) {
	ex_note := newLine(0, `page numbers are accoring to the pdf made by jmr --jmr `)
	note, isNote := parseNote(ex_note)
	want_note := `page numbers are accoring to the pdf made by jmr --jmr`
	if !isNote || note != want_note {
		t.Fail()
	}
}

func TestParseQuote(t *testing.T) {
	exTestLines := []string{
		"Quote body followed by a single tab.\tp. 1 ",
		"Quote body followed by 3 spaces.   p. 2",
		"Quote body followed by 1 space and 1 tab. \tp. 3 ",
		"Quote body followed by page number and extra junk. \tp. 4  EXTRA JUNK HERE  ",
		"Quote body followed by two page numbers.\tpp. 5, 6",
		"Quote body followed by two page numbers. \t pp. 10,11",
		"Quote body followed by two page numbers or page range. \t\t pp. 100 101",
		"Quote body followed by page range. \t pp. 240-42",
		"Quote body followed by page range. \t pp. 240 - 42",
		"Quote body followed by page range. \t pp. 240--42",
		"Quote body followed by p 42. \t p 42",
		"Quote body followed by pp 42, 43. \t p 42, 43",
		"Quote body followed by pp 42-43. \t pp  42-43",
	}

	wantQuotes := []Quote{
		Quote{
			Body: `Quote body followed by a single tab.`,
			Page: `1`,
		},
		Quote{

			Body: `Quote body followed by 3 spaces.`,
			Page: `2`,
		},
		Quote{
			Body: `Quote body followed by 1 space and 1 tab.`,
			Page: `3`,
		},
		Quote{
			Body: `Quote body followed by page number and extra junk.`,
			Page: `4`,
		},
		Quote{
			Body: `Quote body followed by two page numbers.`,
			Page: `5, 6`,
		},
		Quote{
			Body: `Quote body followed by two page numbers.`,
			Page: `10,11`,
		},
		Quote{
			Body: `Quote body followed by two page numbers or page range.`,
			Page: `100 101`,
		},
		Quote{
			Body: `Quote body followed by page range.`,
			Page: `240-42`,
		},
		Quote{
			Body: `Quote body followed by page range.`,
			Page: `240 - 42`,
		},

		Quote{
			Body: `Quote body followed by page range.`,
			Page: `240--42`,
		},
		Quote{
			Body: `Quote body followed by p 42.`,
			Page: `42`,
		},
		Quote{
			Body: `Quote body followed by pp 42, 43.`,
			Page: `42, 43`,
		},

		Quote{
			Body: `Quote body followed by pp 42-43.`,
			Page: `42-43`,
		},
	}

	var exQuoteLines []Line
	for n, l := range exTestLines {
		exQuoteLines = append(exQuoteLines, newLine(n, l))
	}

	for n, ql := range exQuoteLines {
		q, isQuote := parseQuote(ql)
		//q.Body = strings.TrimSpace(q.Body)
		q.Page = strings.TrimSpace(q.Page)
		want := wantQuotes[n]
		if !isQuote {
			t.Logf("unable to parse <exTestLines[%d]\n", n)
			t.FailNow()
		}
		if q.Body != want.Body {
			t.Errorf("failure in Body of <exTestLines[%d]>\n"+
				"Body = %s\n\n"+
				" want: %s",
				n, q.Body, want.Body)
		}
		if q.Page != want.Page {
			t.Errorf("failure in Page of <exTestLines[%d]>\n"+
				"Page = %s\n want: %s\n",
				n, q.Page, want.Page)
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
	workDir := GetWorkDir("")
	batchPath := "" // not testing batch mode
	testDir := "test_files"
	testFiles := []string{
		"test_descriptive_citations.docx",
		"test_example_citations.docx",
		"bib22e_FUNKY.docx",
	}

	volume := false    // no volume information written
	dateStamp := false // no datestamp information written
	t.Chdir(testDir)   //
	for _, tf := range testFiles {
		dataList, workPath := GetWorkPath(workDir, batchPath, tf)

		// Write results to test directory.
		WriteResults(workPath, dataList, volume, dateStamp)

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

		if !bytes.Equal(result, want) {
			t.Errorf("%s does not match %s\n", resultPath, wantPath)
		}

		_ = os.Remove(resultPath)
		_ = os.Remove(discardPath)
	}
}
