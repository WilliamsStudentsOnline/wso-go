package transcriptpdf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// One character every 4.32pt, like the reports' monospaced columns
func text(value string, x, top float64) []char {
	var out []char
	for i, r := range value {
		out = append(out, char{text: string(r), x0: x + float64(i)*4.32, top: top})
	}
	return out
}

func testPage(kind string) page {
	label := "Academic Progress Report"
	if kind == "b" {
		label = "Internal Transcript - CONFIDENTIAL"
	}
	var chars []char
	chars = append(chars, text(label, 20, 20)...)
	chars = append(chars, text("WILLIAMS COLLEGE", 310, 20)...)
	chars = append(chars, text("Office of the Registrar", 310, 30)...)
	chars = append(chars, text("STUDENT-SECRET", 20, 50)...)
	return page{width: 612, height: 792, chars: chars}
}

func progressRow(top float64, term string) []char {
	var out []char
	out = append(out, text("Example Course", 35, top)...)
	out = append(out, text("TEST", 199, top)...)
	out = append(out, text("127T", 225, top)...)
	out = append(out, text("A+", 247, top)...)
	out = append(out, text(term, 264, top)...)
	return append(out, text("1.6", 281, top)...)
}

func TestProgressReportKeepsOnlyClasses(t *testing.T) {
	p := testPage("a")
	p.chars = append(p.chars, text("Classes", 26, 120)...)
	p.chars = append(p.chars, progressRow(150, "S26")...)
	p.chars = append(p.chars, text("Tests", 26, 480)...)
	p.chars = append(p.chars, progressRow(500, "F25")...) // placement tests aren't coursework

	rep, err := extract(p)
	assert.NoError(t, err)
	assert.Equal(t, FormatProgressReport, rep.Format)
	assert.Equal(t, []Course{{Subject: "TEST", Number: "127T", Title: "Example Course", Credits: 1.6, Term: "S26"}}, rep.Courses)
	assert.NotContains(t, fmt.Sprint(rep), "A+")
	assert.NotContains(t, fmt.Sprint(rep), "SECRET")
}

func TestTranscriptColumnsFractionalCreditAndTransfer(t *testing.T) {
	p := testPage("b")
	for _, c := range []struct {
		offset  float64
		heading string
	}{{0, "Fall 2025-2026"}, {285.12, "Winter Study 2025-2026"}} {
		o := c.offset
		p.chars = append(p.chars, text(c.heading, 26+o, 100)...)
		p.chars = append(p.chars, text("TEST", 48+o, 120)...)
		p.chars = append(p.chars, text("23", 69+o, 120)...)
		p.chars = append(p.chars, text("Example", 91+o, 120)...)
		p.chars = append(p.chars, text("INSTRUCTOR", 212+o, 120)...)
		p.chars = append(p.chars, text(".5", 268+o, 120)...)
		p.chars = append(p.chars, text("A+", 281+o, 120)...)
		p.chars = append(p.chars, text("Transfer Credit", 26+o, 140)...)
		p.chars = append(p.chars, text("TEST", 48+o, 150)...)
		p.chars = append(p.chars, text("99", 69+o, 150)...)
	}
	rep, err := extract(p)
	assert.NoError(t, err)
	assert.Len(t, rep.Courses, 2)
	assert.Equal(t, "F25", rep.Courses[0].Term)
	assert.Equal(t, "W26", rep.Courses[1].Term)
	for _, c := range rep.Courses {
		assert.Equal(t, 0.5, c.Credits)
	}
	assert.NotContains(t, fmt.Sprint(rep), "INSTRUCTOR")
}

func TestBadLayoutAndMalformedRowFail(t *testing.T) {
	p := testPage("a")
	_, err := extract(p)
	assert.Equal(t, ErrUnsupported, err)

	p.chars = append(p.chars, text("Classes", 26, 120)...)
	p.chars = append(p.chars, progressRow(150, "???")...)
	p.chars = append(p.chars, text("Tests", 26, 480)...)
	_, err = extract(p)
	assert.Equal(t, ErrUnsupported, err)

	p.width = 700
	_, err = extract(p)
	assert.Equal(t, ErrUnsupported, err)
}

func TestDuplicateCourseFails(t *testing.T) {
	p := testPage("a")
	p.chars = append(p.chars, text("Classes", 26, 120)...)
	p.chars = append(p.chars, progressRow(150, "S26")...)
	p.chars = append(p.chars, progressRow(160, "S26")...)
	p.chars = append(p.chars, text("Tests", 26, 480)...)
	_, err := extract(p)
	assert.Equal(t, ErrUnsupported, err)
}

func TestReadRejectsNonPDF(t *testing.T) {
	_, err := Read([]byte("hello"))
	assert.Equal(t, ErrUnsupported, err)
	_, err = Read([]byte("%PDF-1.4 garbage"))
	assert.Equal(t, ErrUnsupported, err)
}

func TestStrm(t *testing.T) {
	for term, want := range map[string]int{"F26": 1271, "W27": 1272, "S27": 1273, "F23": 1241} {
		assert.Equal(t, want, Course{Term: term}.Strm(), term)
	}
}
