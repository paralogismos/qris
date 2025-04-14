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
	if cit.Name != want_name {
		t.Fail()
	}
}
