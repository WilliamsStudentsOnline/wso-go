package williams_record_update

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultFeedURL = "https://williamsrecord.com/feed/"

type rssFeed struct {
	Items []rssItem `xml:"channel>item"`
}

type rssItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	GUID        string   `xml:"guid"`
	Creator     string   `xml:"http://purl.org/dc/elements/1.1/ creator"`
	PubDate     string   `xml:"pubDate"`
	Categories  []string `xml:"category"`
	Description string   `xml:"description"`
	Content     string   `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
}

func fetchFeed(client *http.Client, feedURL string) ([]rssItem, error) {
	resp, err := client.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", feedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get %s: status %d", feedURL, resp.StatusCode)
	}
	return parseFeed(resp.Body)
}

func parseFeed(r io.Reader) ([]rssItem, error) {
	var feed rssFeed
	if err := xml.NewDecoder(r).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode rss: %w", err)
	}
	return feed.Items, nil
}

func parsePubDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{time.RFC1123Z, time.RFC1123} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized pubDate %q", s)
}
