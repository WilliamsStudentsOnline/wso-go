package williams_record_update

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var blockTags = map[string]bool{
	"p": true, "div": true, "section": true, "article": true, "figure": true, "figcaption": true,
	"blockquote": true, "ul": true, "ol": true, "li": true, "pre": true, "table": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true, "hr": true,
}

var skipTags = map[string]bool{
	"script": true, "style": true, "noscript": true, "iframe": true, "svg": true,
	"form": true, "button": true, "video": true, "audio": true, "object": true, "embed": true,
}

func parseContent(content string) ([]ArticleBlock, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	p := &contentParser{}
	for _, n := range doc.Find("body").Nodes {
		p.walkChildren(n)
	}
	p.flush()
	return p.blocks, nil
}

type contentParser struct {
	blocks []ArticleBlock
	// loose inline content between block elements
	pending []*html.Node
}

func (p *contentParser) walkChildren(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.walk(c)
	}
}

func (p *contentParser) walk(n *html.Node) {
	if n.Type == html.TextNode {
		p.pending = append(p.pending, n)
		return
	}
	if n.Type != html.ElementNode || skipTags[n.Data] {
		return
	}

	switch {
	case n.Data == "img":
		p.flush()
		p.addImage(n, attr(n, "alt"))
	case n.Data == "figure" && containsTag(n, "img"):
		p.flush()
		p.addFigure(n)
	case blockTags[n.Data] && (hasBlockDescendant(n) || containsTag(n, "img")):
		p.flush()
		p.walkChildren(n)
		p.flush()
	case n.Data == "hr":
		p.flush()
	case blockTags[n.Data]:
		p.flush()
		p.addText(textOf(n))
	case containsTag(n, "img") || hasBlockDescendant(n):
		// e.g. <a><img></a>, <span><p>
		p.walkChildren(n)
	default:
		p.pending = append(p.pending, n)
	}
}

func (p *contentParser) flush() {
	if len(p.pending) == 0 {
		return
	}
	var b strings.Builder
	for _, n := range p.pending {
		writeText(&b, n)
	}
	p.pending = nil
	p.addText(b.String())
}

func (p *contentParser) addText(s string) {
	if s = normalizeText(s); s != "" {
		p.blocks = append(p.blocks, ArticleBlock{Type: ArticleBlockText, Text: s})
	}
}

func (p *contentParser) addImage(img *html.Node, caption string) {
	src := imageSrc(img)
	if src == "" {
		return
	}
	p.blocks = append(p.blocks, ArticleBlock{
		Type:    ArticleBlockImage,
		URL:     src,
		Caption: normalizeText(caption),
	})
}

// galleries nest a figure per image
func (p *contentParser) addFigure(fig *html.Node) {
	if containsTag(fig, "figure") {
		p.walkChildren(fig)
		p.flush()
		return
	}
	var caption string
	if c := findAll(fig, "figcaption"); len(c) > 0 {
		caption = textOf(c[0])
	}
	imgs := findAll(fig, "img")
	for i, img := range imgs {
		c := attr(img, "alt")
		if i == len(imgs)-1 && caption != "" {
			c = caption
		}
		p.addImage(img, c)
	}
}

func writeText(b *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		b.WriteString(n.Data)
		return
	case html.ElementNode:
		if skipTags[n.Data] || n.Data == "img" {
			return
		}
		if n.Data == "br" {
			b.WriteString("\n")
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		writeText(b, c)
	}
	if n.Type == html.ElementNode && n.Data == "li" {
		b.WriteString("\n")
	}
}

func textOf(n *html.Node) string {
	var b strings.Builder
	writeText(&b, n)
	return b.String()
}

var invisibleReplacer = strings.NewReplacer(
	"\u00a0", " ",
	"\u00ad", "",
	"\u200b", "",
	"\u200c", "",
	"\u200d", "",
	"\ufeff", "",
)

func normalizeText(s string) string {
	s = invisibleReplacer.Replace(s)
	lines := strings.Split(s, "\n")
	out := lines[:0]
	for _, line := range lines {
		if line = strings.Join(strings.Fields(line), " "); line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func imageSrc(img *html.Node) string {
	// lazy loaders put a placeholder in src
	for _, key := range []string{"data-src", "data-lazy-src", "src"} {
		v := strings.TrimSpace(attr(img, key))
		if v == "" || strings.HasPrefix(v, "data:") {
			continue
		}
		if strings.HasPrefix(v, "//") {
			v = "https:" + v
		}
		return v
	}
	return ""
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasBlockDescendant(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (blockTags[c.Data] || hasBlockDescendant(c)) {
			return true
		}
	}
	return false
}

func containsTag(n *html.Node, tag string) bool {
	return len(findAll(n, tag)) > 0
}

func findAll(n *html.Node, tag string) []*html.Node {
	var out []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if c.Data == tag {
			out = append(out, c)
		}
		out = append(out, findAll(c, tag)...)
	}
	return out
}
