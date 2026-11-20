package williams_record_update

import "time"

const (
	ArticleBlockText  = "text"
	ArticleBlockImage = "image"
)

type ArticleBlock struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	URL     string `json:"url,omitempty"`
	Caption string `json:"caption,omitempty"`
}

type Article struct {
	GUID        string         `json:"guid"`
	Title       string         `json:"title"`
	Link        string         `json:"link"`
	Author      string         `json:"author"`
	Description string         `json:"description"`
	Categories  []string       `json:"categories"`
	PublishedAt time.Time      `json:"publishedAt"`
	Content     []ArticleBlock `json:"content"`
}
