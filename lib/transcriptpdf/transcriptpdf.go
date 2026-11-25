// Package transcriptpdf reads courses off Williams progress report (WMSR102A) and transcript (WMSR102B) PDFs, never grades or names
package transcriptpdf

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

const MaxBytes = 2 * 1024 * 1024

// Deliberately vague, so nothing from the document leaks
var ErrUnsupported = errors.New("transcriptpdf: unsupported or malformed report")

const (
	FormatProgressReport = "wmsr102a"
	FormatTranscript     = "wmsr102b"
)

type Course struct {
	Subject string // e.g. "CSCI"
	Number  string // as printed, e.g. "136" or "397T"
	Title   string
	Credits float64
	Term    string // as printed: F23, W24, S24
}

// 1 + academic year's spring year + 1 fall, 2 winter, 3 spring: F26 is 1271
func (c Course) Strm() int {
	yy, _ := strconv.Atoi(c.Term[1:])
	switch c.Term[0] {
	case 'F':
		return 1000 + (yy+1)*10 + 1
	case 'W':
		return 1000 + yy*10 + 2
	default:
		return 1000 + yy*10 + 3
	}
}

type Report struct {
	Format  string
	Courses []Course
}

func Read(data []byte) (rep Report, err error) {
	if len(data) > MaxBytes || !bytes.HasPrefix(data, []byte("%PDF-")) {
		return Report{}, ErrUnsupported
	}
	// The PDF library panics on some malformed files
	defer func() {
		if recover() != nil {
			rep, err = Report{}, ErrUnsupported
		}
	}()
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || r.NumPage() != 1 {
		return Report{}, ErrUnsupported
	}
	p := r.Page(1)
	box := p.V.Key("MediaBox")
	pg := page{width: box.Index(2).Float64() - box.Index(0).Float64(), height: box.Index(3).Float64() - box.Index(1).Float64()}
	for _, t := range p.Content().Text {
		// Column positions measure from the top of the page to the top of the glyph
		pg.chars = append(pg.chars, char{text: t.S, x0: t.X, top: pg.height - t.Y - t.FontSize})
	}
	return extract(pg)
}

type char struct {
	text    string
	x0, top float64
}

type page struct {
	width, height float64
	chars         []char
}

func cell(chars []char, start, end float64) string {
	in := make([]char, 0, len(chars))
	for _, c := range chars {
		if start <= c.x0 && c.x0 < end {
			in = append(in, c)
		}
	}
	sort.SliceStable(in, func(i, j int) bool { return in[i].x0 < in[j].x0 })
	var b strings.Builder
	for _, c := range in {
		b.WriteString(c.text)
	}
	return strings.TrimSpace(b.String())
}

func rows(chars []char) [][]char {
	sorted := append([]char(nil), chars...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].top != sorted[j].top {
			return sorted[i].top < sorted[j].top
		}
		return sorted[i].x0 < sorted[j].x0
	})
	var out [][]char
	var lastTop float64
	for _, c := range sorted {
		if len(out) == 0 || abs(lastTop-c.top) > 1 {
			out = append(out, nil)
			lastTop = c.top
		}
		out[len(out)-1] = append(out[len(out)-1], c)
	}
	return out
}

func filter(chars []char, keep func(char) bool) []char {
	var out []char
	for _, c := range chars {
		if keep(c) {
			out = append(out, c)
		}
	}
	return out
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

var (
	subjectPattern = regexp.MustCompile(`^[A-Z]{2,4}$`)
	numberPattern  = regexp.MustCompile(`^\d{2,3}T?$`)
	creditsPattern = regexp.MustCompile(`^(?:\d+(?:\.\d+)?|\.\d+)$`)
	termPattern    = regexp.MustCompile(`^[FSW]\d{2}$`)
	headingPattern = regexp.MustCompile(`^(Fall|Spring|Winter Study) (20\d{2})-(20\d{2})$`)
)

type span struct{ start, end float64 }

func (s span) shift(by float64) span { return span{s.start + by, s.end + by} }

func record(chars []char, title, subject, number, credits span, term string) (Course, error) {
	c := Course{
		Subject: cell(chars, subject.start, subject.end),
		Number:  cell(chars, number.start, number.end),
		Title:   cell(chars, title.start, title.end),
		Term:    term,
	}
	cr := cell(chars, credits.start, credits.end)
	if !subjectPattern.MatchString(c.Subject) || !numberPattern.MatchString(c.Number) ||
		!creditsPattern.MatchString(cr) || !termPattern.MatchString(term) || c.Title == "" {
		return Course{}, ErrUnsupported
	}
	c.Credits, _ = strconv.ParseFloat(cr, 64)
	if c.Credits <= 0 || c.Credits > 10 {
		return Course{}, ErrUnsupported
	}
	return c, nil
}

func extract(p page) (Report, error) {
	if abs(p.width-612) > 1 || abs(p.height-792) > 1 {
		return Report{}, ErrUnsupported
	}
	var header []string
	for _, row := range rows(filter(p.chars, func(c char) bool { return c.top < 80 })) {
		header = append(header, cell(row, 0, 612))
	}
	head := strings.Join(header, "\n")
	if !strings.Contains(head, "WILLIAMS COLLEGE") || !strings.Contains(head, "Office of the Registrar") {
		return Report{}, ErrUnsupported
	}
	var rep Report
	switch {
	case strings.Contains(head, "Academic Progress Report"):
		rep.Format = FormatProgressReport
	case strings.Contains(head, "Internal Transcript - CONFIDENTIAL"):
		rep.Format = FormatTranscript
	default:
		return Report{}, ErrUnsupported
	}

	if rep.Format == FormatProgressReport {
		active, foundEnd := false, false
		for _, row := range rows(filter(p.chars, func(c char) bool { return c.x0 < 298 })) {
			label := cell(row, 0, 298)
			if label == "Classes" {
				active = true
			} else if active && label == "Tests" {
				foundEnd = true
				break
			} else if active && subjectPattern.MatchString(cell(row, 198, 223)) {
				c, err := record(row, span{35, 198}, span{198, 223}, span{224, 245}, span{280, 298}, cell(row, 263, 280))
				if err != nil {
					return Report{}, err
				}
				rep.Courses = append(rep.Courses, c)
			}
		}
		if !foundEnd {
			return Report{}, ErrUnsupported
		}
	} else {
		// Two columns of terms, the right one 285.12pt over
		for _, offset := range []float64{0, 285.12} {
			term := ""
			inColumn := func(c char) bool { return 20+offset <= c.x0 && c.x0 < 298+offset }
			for _, row := range rows(filter(p.chars, inColumn)) {
				label := cell(row, 20+offset, 298+offset)
				if m := headingPattern.FindStringSubmatch(label); m != nil {
					start, _ := strconv.Atoi(m[2])
					end, _ := strconv.Atoi(m[3])
					if end != start+1 {
						return Report{}, ErrUnsupported
					}
					year := m[3]
					if m[1] == "Fall" {
						year = m[2]
					}
					term = map[string]string{"Fall": "F", "Spring": "S", "Winter Study": "W"}[m[1]] + year[2:]
				} else if label == "Transfer Credit" || label == "Summary Information" {
					term = ""
				} else if term != "" && subjectPattern.MatchString(cell(row, 48+offset, 68+offset)) {
					c, err := record(row, span{91, 211.5}.shift(offset), span{48, 68}.shift(offset),
						span{69, 90}.shift(offset), span{268, 280}.shift(offset), term)
					if err != nil {
						return Report{}, err
					}
					rep.Courses = append(rep.Courses, c)
				}
			}
		}
	}

	if len(rep.Courses) == 0 || len(rep.Courses) > 200 {
		return Report{}, ErrUnsupported
	}
	seen := map[string]bool{}
	for _, c := range rep.Courses {
		key := fmt.Sprintf("%s/%s/%s", c.Subject, c.Number, c.Term)
		if seen[key] {
			return Report{}, ErrUnsupported
		}
		seen[key] = true
	}
	return rep, nil
}
