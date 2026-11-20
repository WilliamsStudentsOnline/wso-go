package williams_record_update

import (
	"testing"

	testify "github.com/stretchr/testify/assert"
)

func text(s string) ArticleBlock {
	return ArticleBlock{Type: ArticleBlockText, Text: s}
}

func image(url, caption string) ArticleBlock {
	return ArticleBlock{Type: ArticleBlockImage, URL: url, Caption: caption}
}

func TestParseContent(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []ArticleBlock
	}{
		{
			name: "paragraphs with inline formatting",
			in:   `<p class="p1"><span class="s1">The <i>Record</i>  reported   it.</span></p><p><b>Q:</b> Why?</p>`,
			want: []ArticleBlock{text("The Record reported it."), text("Q: Why?")},
		},
		{
			name: "empty and whitespace paragraphs dropped",
			in:   "<p></p><p>&nbsp;</p><p><span> </span></p><p>kept</p>",
			want: []ArticleBlock{text("kept")},
		},
		{
			name: "figure with caption in document order",
			in: `<p>before</p><figure class="wp-caption"><img src="https://x/a.png" alt="alt" srcset="https://x/a-300.png 300w"/>` +
				`<figcaption>Name/<em>The Williams Record</em></figcaption></figure><p>after</p>`,
			want: []ArticleBlock{text("before"), image("https://x/a.png", "Name/The Williams Record"), text("after")},
		},
		{
			name: "multiple figures",
			in:   `<figure><img src="https://x/1.png"/><figcaption>one</figcaption></figure><p>mid</p><figure><img src="https://x/2.png"/><figcaption>two</figcaption></figure>`,
			want: []ArticleBlock{image("https://x/1.png", "one"), text("mid"), image("https://x/2.png", "two")},
		},
		{
			name: "image inside paragraph uses alt",
			in:   `<p><a href="https://x/full.png"><img src="https://x/a.png" alt="a cat"/></a></p>`,
			want: []ArticleBlock{image("https://x/a.png", "a cat")},
		},
		{
			name: "image inside paragraph with text",
			in:   `<p>lead <img src="https://x/a.png"/> tail</p>`,
			want: []ArticleBlock{text("lead"), image("https://x/a.png", ""), text("tail")},
		},
		{
			name: "lazy loaded image prefers data-src",
			in:   `<figure><img src="data:image/gif;base64,R0lGOD" data-src="//x/real.jpg"/></figure>`,
			want: []ArticleBlock{image("https://x/real.jpg", "")},
		},
		{
			name: "image without src skipped",
			in:   `<p><img alt="nothing"/></p><p>text</p>`,
			want: []ArticleBlock{text("text")},
		},
		{
			name: "gallery of nested figures",
			in: `<figure class="wp-block-gallery"><figure class="wp-block-image"><img src="https://x/1.png"/><figcaption>one</figcaption></figure>` +
				`<figure class="wp-block-image"><img src="https://x/2.png"/></figure><figcaption>gallery caption</figcaption></figure>`,
			want: []ArticleBlock{image("https://x/1.png", "one"), image("https://x/2.png", ""), text("gallery caption")},
		},
		{
			name: "line breaks preserved within a paragraph",
			in:   "<p>Team: Soccer<br/>Hometown:   Newton<br><br/>Major: History</p>",
			want: []ArticleBlock{text("Team: Soccer\nHometown: Newton\nMajor: History")},
		},
		{
			name: "nested containers flattened",
			in:   `<div class="wrap"><div><p>one</p><blockquote><p>quoted</p></blockquote></div></div><p>two</p>`,
			want: []ArticleBlock{text("one"), text("quoted"), text("two")},
		},
		{
			name: "leaf blockquote and headings",
			in:   `<h2>Heading</h2><blockquote>Just a quote</blockquote>`,
			want: []ArticleBlock{text("Heading"), text("Just a quote")},
		},
		{
			name: "lists keep items on separate lines",
			in:   `<ul><li>a</li><li>b <i>c</i></li></ul>`,
			want: []ArticleBlock{text("a"), text("b c")},
		},
		{
			name: "embeds and scripts dropped",
			in:   `<p>keep</p><iframe src="https://youtube.com/embed/x"></iframe><script>alert(1)</script><style>p{}</style><p>also</p>`,
			want: []ArticleBlock{text("keep"), text("also")},
		},
		{
			name: "loose inline text between blocks",
			in:   `stray <b>bold</b> text<p>para</p>more <a href="#">stray</a>`,
			want: []ArticleBlock{text("stray bold text"), text("para"), text("more stray")},
		},
		{
			name: "block nested in inline span",
			in:   `<span><p>inside</p><p>span</p></span>`,
			want: []ArticleBlock{text("inside"), text("span")},
		},
		{
			name: "invisible characters stripped",
			in:   "<p>Not in Our Name­ ​– No</p>",
			want: []ArticleBlock{text("Not in Our Name – No")},
		},
		{
			name: "entities decoded",
			in:   "<p>Tea &amp; Boba &#8230; &#8220;hi&#8221;</p>",
			want: []ArticleBlock{text("Tea & Boba … “hi”")},
		},
		{
			name: "empty content",
			in:   "",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseContent(tt.in)
			testify.NoError(t, err)
			testify.Equal(t, tt.want, got)
		})
	}
}
