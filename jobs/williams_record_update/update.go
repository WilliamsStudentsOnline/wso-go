package williams_record_update

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Update(client *http.Client, feedURL, out string) (int, error) {
	items, err := fetchFeed(client, feedURL)
	if err != nil {
		return 0, err
	}

	articles := []Article{}
	for _, item := range items {
		article, err := toArticle(item)
		if err != nil {
			log.Printf("skipping %q: %v", item.Link, err)
			continue
		}
		articles = append(articles, article)
	}
	if len(articles) == 0 {
		return 0, errors.New("feed had no usable articles")
	}

	if err := writeJSON(out, articles); err != nil {
		return 0, err
	}
	return len(articles), nil
}

func toArticle(item rssItem) (Article, error) {
	guid := strings.TrimSpace(item.GUID)
	if guid == "" {
		guid = strings.TrimSpace(item.Link)
	}
	if guid == "" {
		return Article{}, errors.New("missing guid and link")
	}
	published, err := parsePubDate(item.PubDate)
	if err != nil {
		return Article{}, err
	}
	content, err := parseContent(item.Content)
	if err != nil {
		return Article{}, fmt.Errorf("parse content: %w", err)
	}
	description, err := parseContent(item.Description)
	if err != nil {
		return Article{}, fmt.Errorf("parse description: %w", err)
	}

	categories := []string{}
	for _, c := range item.Categories {
		if c = strings.TrimSpace(html.UnescapeString(c)); c != "" {
			categories = append(categories, c)
		}
	}
	if content == nil {
		content = []ArticleBlock{}
	}

	return Article{
		GUID:        guid,
		Title:       normalizeText(html.UnescapeString(item.Title)),
		Link:        strings.TrimSpace(item.Link),
		Author:      normalizeText(html.UnescapeString(item.Creator)),
		Description: joinText(description),
		Categories:  categories,
		PublishedAt: published,
		Content:     content,
	}, nil
}

func joinText(blocks []ArticleBlock) string {
	var parts []string
	for _, b := range blocks {
		if b.Type == ArticleBlockText {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n\n")
}

func writeJSON(dest string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".tmp.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := json.NewEncoder(tmp).Encode(v); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("encode %s: %w", dest, err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), dest); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
