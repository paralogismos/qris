// qris_test.go

package qris

import (
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
		`Focal attention is focal perception, not perception focused by attention. 		p. 3`,
		`Self and world are not separate entities that interact, but relations in a continuum of transformation. The world is a product of the same process that lays down the self. 	p. 4 `,
		`... [W]e conclude that the belief in a real world does not depend on the realness of the perception but on the coherence of perceptions within and across the different modalities. 	p. 5  w^activity_pattern  `,
		`The self in a semantic state has a different quality of consciousness from the self in a phonological state. The pathological material helps to dispose of the notion that the self is a substance and mental states are its properties. The properties of the self are states of its existence, not changing features of a self that is unchanging. 	p. 7 `,
		`Contrastive features are important. Contrast involves relations which implicate temporal process. In the relation of figure to ground, there is a microtemporal transition in the kinetics of synchronicity that carves local unities out of the simultaneity of the whole. 	p. 7`,
		`[S]elf, image, and world are a single object in transition. 	p. 8  epigraph `,
		`The duration of the present creates a theater for experience, but does not fully explain the unity of the self. We need also to explore the nature of categories and the analogy of a duration and its instants to a category and its members. A duration is a "container" of temporal parts; a category is a "container" of spatial parts. A duration is a container of arbitrary length that enfolds a number of instants; a category is a container of an arbitrary set of members. A category is like a duration in that both are enclosures with fuzzy boundaries for virtual parts that are, themselves, potential containers. An instant is like a member of a category in that it can be a category for another member. Duration is the primordial manifestation of categorization. 		p. 8  (notice: according to Brown, time is primordial)  --jmr `,
		`The relation of succession in a nontemporal (vertical) becoming transforms to one of precedence in the serialization of the (horizontal) present by events. The nontemporal actualization precipitates cotemporal entities. The concurrence of these entities in the specious present is consciousness, i.e., the conscious state. The self forms the past boundary, the surface images and objects form the actual boundary, of this duration. 	p. 9  space precipitates simultaneous objects which equals consciousness -jmr `,
		`Finally, to sum up what this signifies for the biological basis of the self, the neuropsychological material demonstrates that the self is deposited in the process of object realization, that it distributes into images and objects, and that a truncation of this process results in an erosion of the self that is similar across the different perceptual modalities. The self is categorical and relational, achieving autonomy in the context of a complete derivation. The autonomy depends on the completeness. The preliminary locus of the self in the mental state entails a holistic or multimodal phase of potential prior to perceptual individuation. This, together with a relation to feeling, to the personal history and the immediate past, point to a limbic transition in the outward development of the mental state. 	p. 14  penultimate paragraph`}

	wantQuotes := []Quote{
		Quote{
			Body: `Focal attention is focal perception, not perception focused by attention.`,
			Page: `3`,
		},
		Quote{
			Body: `Self and world are not separate entities that interact, but relations in a continuum of transformation. The world is a product of the same process that lays down the self.`,
			Page: `4`,
		},
		Quote{
			Body: `... [W]e conclude that the belief in a real world does not depend on the realness of the perception but on the coherence of perceptions within and across the different modalities.`,
			Page: `5`,
		},
		Quote{
			Body: `The self in a semantic state has a different quality of consciousness from the self in a phonological state. The pathological material helps to dispose of the notion that the self is a substance and mental states are its properties. The properties of the self are states of its existence, not changing features of a self that is unchanging.`,
			Page: `7`,
		},
		Quote{
			Body: `Contrastive features are important. Contrast involves relations which implicate temporal process. In the relation of figure to ground, there is a microtemporal transition in the kinetics of synchronicity that carves local unities out of the simultaneity of the whole.`,
			Page: `7`,
		},
		Quote{
			Body: `[S]elf, image, and world are a single object in transition.`,
			Page: `8`,
		},
		Quote{
			Body: `The duration of the present creates a theater for experience, but does not fully explain the unity of the self. We need also to explore the nature of categories and the analogy of a duration and its instants to a category and its members. A duration is a "container" of temporal parts; a category is a "container" of spatial parts. A duration is a container of arbitrary length that enfolds a number of instants; a category is a container of an arbitrary set of members. A category is like a duration in that both are enclosures with fuzzy boundaries for virtual parts that are, themselves, potential containers. An instant is like a member of a category in that it can be a category for another member. Duration is the primordial manifestation of categorization.`,
			Page: `8`,
		},
		Quote{
			Body: `The relation of succession in a nontemporal (vertical) becoming transforms to one of precedence in the serialization of the (horizontal) present by events. The nontemporal actualization precipitates cotemporal entities. The concurrence of these entities in the specious present is consciousness, i.e., the conscious state. The self forms the past boundary, the surface images and objects form the actual boundary, of this duration.`,
			Page: `9`,
		},
		Quote{
			Body: `Finally, to sum up what this signifies for the biological basis of the self, the neuropsychological material demonstrates that the self is deposited in the process of object realization, that it distributes into images and objects, and that a truncation of this process results in an erosion of the self that is similar across the different perceptual modalities. The self is categorical and relational, achieving autonomy in the context of a complete derivation. The autonomy depends on the completeness. The preliminary locus of the self in the mental state entails a holistic or multimodal phase of potential prior to perceptual individuation. This, together with a relation to feeling, to the personal history and the immediate past, point to a limbic transition in the outward development of the mental state.`,
			Page: `14`,
		},
	}

	var exQuoteLines []Line
	for n, l := range exTestLines {
		exQuoteLines = append(exQuoteLines, newLine(n, l))
	}

	for n, ql := range exQuoteLines {
		q, isQuote := parseQuote(ql)
		q.Body = strings.TrimSpace(q.Body)
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
