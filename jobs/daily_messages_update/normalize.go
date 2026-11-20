package daily_messages_update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strconv"
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

func Normalize(raw []byte, date string, now time.Time) (Export, []error, error) {
	out := Export{
		Date:       date,
		UpdateTime: now.In(eastern).Format(time.RFC3339),
		Categories: []Category{},
	}

	trimmed := bytes.TrimSpace(raw)
	if bytes.HasPrefix(trimmed, []byte("[")) {
		var empty []json.RawMessage
		if err := json.Unmarshal(trimmed, &empty); err != nil {
			return Export{}, nil, err
		}
		if len(empty) != 0 {
			return Export{}, nil, fmt.Errorf("expected object or empty array, got array of %d", len(empty))
		}
		return out, nil, nil
	}

	var skipped []error
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	if err := expectDelim(dec, '{'); err != nil {
		return Export{}, nil, err
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return Export{}, nil, err
		}
		name, ok := tok.(string)
		if !ok {
			return Export{}, nil, fmt.Errorf("unexpected category key %v", tok)
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return Export{}, nil, fmt.Errorf("category %q: %w", name, err)
		}
		items, err := categoryItems(value)
		if err != nil {
			skipped = append(skipped, fmt.Errorf("category %q: %w", name, err))
			continue
		}
		cat := Category{Name: cleanText(name), Messages: make([]Message, 0, len(items))}
		for i, item := range items {
			var r rawMessage
			if err := json.Unmarshal(item, &r); err != nil {
				skipped = append(skipped, fmt.Errorf("category %q item %d: %w", name, i, err))
				continue
			}
			cat.Messages = append(cat.Messages, normalizeMessage(r))
		}
		out.Categories = append(out.Categories, cat)
	}
	if err := expectDelim(dec, '}'); err != nil {
		return Export{}, nil, err
	}
	out.Date = feedDate(out, date)
	return out, skipped, nil
}

// the date most messages run on, since upstream may flip days later than midnight
func feedDate(out Export, fallback string) string {
	counts := map[string]int{}
	for _, c := range out.Categories {
		for _, m := range c.Messages {
			for _, d := range m.RunDates {
				counts[d]++
			}
		}
	}
	best := fallback
	for d, n := range counts {
		if n > counts[best] || (n == counts[best] && best != fallback && d < best) {
			best = d
		}
	}
	return best
}

// PHP serializes a filtered list as an object keyed by index
func categoryItems(value json.RawMessage) ([]json.RawMessage, error) {
	if !isObject(value) {
		var items []json.RawMessage
		err := json.Unmarshal(value, &items)
		return items, err
	}
	var byKey map[string]json.RawMessage
	if err := json.Unmarshal(value, &byKey); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, errA := strconv.Atoi(keys[i])
		b, errB := strconv.Atoi(keys[j])
		if errA == nil && errB == nil {
			return a < b
		}
		return keys[i] < keys[j]
	})
	items := make([]json.RawMessage, 0, len(keys))
	for _, k := range keys {
		items = append(items, byKey[k])
	}
	return items, nil
}

func expectDelim(dec *json.Decoder, want json.Delim) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != want {
		return fmt.Errorf("expected %q, got %v", want, tok)
	}
	return nil
}

func normalizeMessage(r rawMessage) Message {
	m := Message{
		ID:        r.ID,
		Type:      string(r.Type),
		Title:     cleanText(string(r.Title)),
		Summary:   htmlToText(string(r.DmText)),
		Body:      htmlToText(string(r.Content)),
		BodyHTML:  strings.TrimSpace(string(r.Content)),
		URL:       stripQuery(string(r.Permalink)),
		MoreURL:   strings.TrimSpace(string(r.WebsiteURL)),
		TicketURL: strings.TrimSpace(string(r.TicketURL)),
		ImageURL:  strings.TrimSpace(string(r.ThumbURLFull)),
		Author: Author{
			Name:       cleanText(string(r.Author)),
			Email:      strings.TrimSpace(string(r.AuthorEmail)),
			Department: cleanText(string(r.LdapDepartment)),
		},
		Organizations: cleanText(string(r.Orgs)),
		Audiences:     append([]string{}, r.Terms.Audience...),
		RunDates:      []string{},
		Departments:   []Department{},
	}
	sort.Strings(m.Audiences)

	for i := 0; ; i++ {
		v, ok := r.Meta[fmt.Sprintf("acf_daily_message_run_dates_%d", i)]
		if !ok {
			break
		}
		if d, err := time.Parse("20060102", v); err == nil {
			m.RunDates = append(m.RunDates, d.Format("2006-01-02"))
		}
	}

	for slug, name := range r.DeptsArr {
		m.Departments = append(m.Departments, Department{Slug: slug, Name: cleanText(name)})
	}
	sort.Slice(m.Departments, func(i, j int) bool { return m.Departments[i].Slug < m.Departments[j].Slug })

	if r.Type != "announcement" {
		m.Event = normalizeEvent(r)
	}
	return m
}

func normalizeEvent(r rawMessage) *Event {
	allDay := strings.Trim(string(r.AllDay), `"`)
	return &Event{
		Start:         parseEventTime(r.Meta["_EventStartDate"]),
		End:           parseEventTime(r.Meta["_EventEndDate"]),
		AllDay:        allDay == "1" || allDay == "true",
		TimeFormatted: cleanText(string(r.TimeFormatted)),
		Venue:         cleanText(string(r.Venue)),
		Room:          cleanText(string(r.VenueRoom)),
	}
}

func parseEventTime(v string) string {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", v, eastern)
	if err != nil || t.Year() < 1900 {
		return ""
	}
	return t.Format(time.RFC3339)
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
