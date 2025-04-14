// qris_test.go

package qris

import (
	"testing"
)

// Test parseCitation
func TestParseCitation(t *testing.T) {
	ex_citation := newLine(0, `Brown, Jason W. "Neuropsychology and the self-concept." The Journal of Nervous and Mental Disease. 187 no.3 (1999e): 131-41.`)
	cit := parseCitation(ex_citation)
	want_name := "Brown, Jason"
	want_year := "1999e"
	want_body := `Brown, Jason W. "Neuropsychology and the self-concept." The Journal of Nervous and Mental Disease. 187 no.3 (1999e): 131-41.`
	if cit.Name != want_name || cit.Year != want_year || cit.Body != want_body {
		t.Fail()
	}
}

// Test parseNote
func TestParseNote(t *testing.T) {
	ex_note := newLine(0, `page numbers are accoring to the pdf made by jmr --jmr `)
	note, isNote := parseNote(ex_note)
	want_note := `page numbers are accoring to the pdf made by jmr --jmr`
	if !isNote || note != want_note {
		t.Fail()
	}
}
