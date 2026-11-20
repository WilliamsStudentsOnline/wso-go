package daily_messages_update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	nethtml "golang.org/x/net/html"
)

var eastern = mustLoadLocation("America/New_York")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

type Export struct {
	Date       string     `json:"date"`
	UpdateTime string     `json:"updateTime"`
	Categories []Category `json:"categories"`
}

type Category struct {
	Name     string    `json:"name"`
	Messages []Message `json:"messages"`
}

type Message struct {
	ID            int          `json:"id"`
	Type          string       `json:"type"`
	Title         string       `json:"title"`
	Summary       string       `json:"summary"`
	Body          string       `json:"body"`
	BodyHTML      string       `json:"bodyHtml"`
	URL           string       `json:"url"`
	MoreURL       string       `json:"moreUrl"`
	TicketURL     string       `json:"ticketUrl"`
	ImageURL      string       `json:"imageUrl"`
	Audiences     []string     `json:"audiences"`
	RunDates      []string     `json:"runDates"`
	Author        Author       `json:"author"`
	Departments   []Department `json:"departments"`
	Organizations string       `json:"organizations"`
	Event         *Event       `json:"event"`
}

type Author struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
}

type Department struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Event struct {
	Start         string `json:"start"`
	End           string `json:"end"`
	AllDay        bool   `json:"allDay"`
	TimeFormatted string `json:"timeFormatted"`
	Venue         string `json:"venue"`
	Room          string `json:"room"`
}

func Normalize(raw []byte, date string, now time.Time) (Export, error) {
	out := Export{
		Date:       date,
		UpdateTime: now.In(eastern).Format(time.RFC3339),
		Categories: []Category{},
	}

	trimmed := bytes.TrimSpace(raw)
	if bytes.HasPrefix(trimmed, []byte("[")) {
		var empty []json.RawMessage
		if err := json.Unmarshal(trimmed, &empty); err != nil {
			return Export{}, err
		}
		if len(empty) != 0 {
			return Export{}, fmt.Errorf("expected object or empty array, got array of %d", len(empty))
		}
		return out, nil
	}

	dec := json.NewDecoder(bytes.NewReader(trimmed))
	if _, err := expectDelim(dec, '{'); err != nil {
		return Export{}, err
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return Export{}, err
		}
		name, ok := tok.(string)
		if !ok {
			return Export{}, fmt.Errorf("unexpected category key %v", tok)
		}
		var raws []rawMessage
		if err := dec.Decode(&raws); err != nil {
			return Export{}, fmt.Errorf("category %q: %w", name, err)
		}
		cat := Category{Name: cleanText(name), Messages: make([]Message, 0, len(raws))}
		for _, r := range raws {
			m, err := normalizeMessage(r)
			if err != nil {
				return Export{}, fmt.Errorf("message %d: %w", r.ID, err)
			}
			cat.Messages = append(cat.Messages, m)
		}
		out.Categories = append(out.Categories, cat)
	}
	if _, err := expectDelim(dec, '}'); err != nil {
		return Export{}, err
	}
	return out, nil
}

func expectDelim(dec *json.Decoder, want json.Delim) (json.Token, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != want {
		return nil, fmt.Errorf("expected %q, got %v", want, tok)
	}
	return tok, nil
}

func normalizeMessage(r rawMessage) (Message, error) {
	m := Message{
		ID:        r.ID,
		Type:      r.Type,
		Title:     cleanText(r.Title),
		Summary:   htmlToText(r.DmText),
		Body:      htmlToText(r.Content),
		BodyHTML:  strings.TrimSpace(r.Content),
		URL:       stripQuery(r.Permalink),
		MoreURL:   strings.TrimSpace(string(r.WebsiteURL)),
		TicketURL: strings.TrimSpace(string(r.TicketURL)),
		ImageURL:  strings.TrimSpace(string(r.ThumbURLFull)),
		Author: Author{
			Name:       cleanText(r.Author),
			Email:      strings.TrimSpace(r.AuthorEmail),
			Department: cleanText(r.LdapDepartment),
		},
		Organizations: cleanText(r.Orgs),
		Audiences:     []string{},
		RunDates:      []string{},
		Departments:   []Department{},
	}

	for _, a := range r.Terms.Audience {
		m.Audiences = append(m.Audiences, a.Slug)
	}
	sort.Strings(m.Audiences)

	for i := 0; ; i++ {
		key := fmt.Sprintf("acf_daily_message_run_dates_%d", i)
		if _, ok := r.Meta[key]; !ok {
			break
		}
		v := r.metaValue(key)
		if v == "" {
			continue
		}
		d, err := time.Parse("20060102", v)
		if err != nil {
			return Message{}, fmt.Errorf("run date %q: %w", v, err)
		}
		m.RunDates = append(m.RunDates, d.Format("2006-01-02"))
	}

	for slug, name := range r.DeptsArr {
		m.Departments = append(m.Departments, Department{Slug: slug, Name: cleanText(name)})
	}
	sort.Slice(m.Departments, func(i, j int) bool { return m.Departments[i].Slug < m.Departments[j].Slug })

	if r.Type != "announcement" {
		ev, err := normalizeEvent(r)
		if err != nil {
			return Message{}, err
		}
		m.Event = ev
	}
	return m, nil
}

func normalizeEvent(r rawMessage) (*Event, error) {
	start, err := parseEventTime(r.metaValue("_EventStartDate"))
	if err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}
	end, err := parseEventTime(r.metaValue("_EventEndDate"))
	if err != nil {
		return nil, fmt.Errorf("end: %w", err)
	}
	allDay := strings.Trim(string(r.AllDay), `"`)
	return &Event{
		Start:         start,
		End:           end,
		AllDay:        allDay == "1" || allDay == "true",
		TimeFormatted: cleanText(r.TimeFormatted),
		Venue:         cleanText(string(r.Venue)),
		Room:          cleanText(string(r.VenueRoom)),
	}, nil
}

func parseEventTime(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", v, eastern)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC3339), nil
}

func stripQuery(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

var spaceRun = regexp.MustCompile(`[ \t\x{a0}]+`)

func cleanText(s string) string {
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "​", "")
	return strings.TrimSpace(spaceRun.ReplaceAllString(s, " "))
}

var blockTags = map[string]bool{
	"p": true, "div": true, "li": true, "ul": true, "ol": true, "br": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"blockquote": true, "tr": true,
}

func htmlToText(s string) string {
	root, err := nethtml.Parse(strings.NewReader(s))
	if err != nil {
		return cleanText(s)
	}
	var b strings.Builder
	var walk func(n *nethtml.Node)
	walk = func(n *nethtml.Node) {
		if isMoreLink(n) {
			return
		}
		if n.Type == nethtml.TextNode {
			b.WriteString(strings.ReplaceAll(n.Data, "\n", " "))
		}
		block := n.Type == nethtml.ElementNode && blockTags[n.Data]
		if block {
			b.WriteString("\n")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if block && n.Data != "br" {
			b.WriteString("\n\n")
		}
	}
	walk(root)

	var paras []string
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			paras = append(paras, strings.Join(cur, "\n"))
			cur = nil
		}
	}
	for _, line := range strings.Split(b.String(), "\n") {
		line = cleanText(line)
		if line == "" {
			flush()
			continue
		}
		cur = append(cur, line)
	}
	flush()
	return strings.Join(paras, "\n\n")
}

func isMoreLink(n *nethtml.Node) bool {
	if n.Type != nethtml.ElementNode || n.Data != "a" {
		return false
	}
	for _, a := range n.Attr {
		if a.Key == "class" && strings.Contains(a.Val, "more-link") {
			return true
		}
	}
	return false
}
