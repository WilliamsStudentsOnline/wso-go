package williams_record_update

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serve(t *testing.T, status int, body []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func readArticles(t *testing.T, path string) []Article {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	var articles []Article
	require.NoError(t, json.Unmarshal(b, &articles))
	return articles
}

func TestUpdateWritesJSON(t *testing.T) {
	assert := testify.New(t)
	feed, err := os.ReadFile("testdata/feed.xml")
	require.NoError(t, err)
	srv := serve(t, http.StatusOK, feed)
	out := filepath.Join(t.TempDir(), "mobile", "williams_record.json")

	n, err := Update(srv.Client(), srv.URL, out)
	assert.NoError(err)
	assert.Equal(10, n)

	articles := readArticles(t, out)
	require.Len(t, articles, 10)
	assert.Equal("https://williamsrecord.com/?p=474112", articles[0].GUID)
	assert.Equal(ArticleBlockImage, articles[0].Content[0].Type)

	info, err := os.Stat(out)
	assert.NoError(err)
	assert.Equal(os.FileMode(0o644), info.Mode().Perm())
	entries, _ := os.ReadDir(filepath.Dir(out))
	assert.Len(entries, 1)
}

func TestUpdateSkipsBadItems(t *testing.T) {
	assert := testify.New(t)
	srv := serve(t, http.StatusOK, []byte(`<?xml version="1.0"?>
<rss xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:dc="http://purl.org/dc/elements/1.1/"><channel>
	<item>
		<title>Tea &#038; Boba</title>
		<link>https://williamsrecord.com/1/</link>
		<guid>https://williamsrecord.com/?p=1</guid>
		<pubDate>Wed, 11 Nov 2026 06:00:00 +0000</pubDate>
		<description><![CDATA[Summary [&#8230;]]]></description>
	</item>
	<item>
		<title>Bad date</title>
		<guid>https://williamsrecord.com/?p=2</guid>
		<pubDate>yesterday</pubDate>
	</item>
</channel></rss>`))
	out := filepath.Join(t.TempDir(), "out.json")

	n, err := Update(srv.Client(), srv.URL, out)
	assert.NoError(err)
	assert.Equal(1, n)

	articles := readArticles(t, out)
	require.Len(t, articles, 1)
	assert.Equal("Tea & Boba", articles[0].Title)
	assert.Equal("Summary […]", articles[0].Description)
	assert.NotNil(articles[0].Categories)
	assert.NotNil(articles[0].Content)
}

func TestUpdateKeepsExistingFileOnFailure(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.json")
	require.NoError(t, os.WriteFile(out, []byte("previous"), 0o644))

	for name, srv := range map[string]*httptest.Server{
		"blocked": serve(t, http.StatusForbidden, []byte("cloudflare")),
		"empty":   serve(t, http.StatusOK, []byte(`<rss><channel></channel></rss>`)),
		"garbage": serve(t, http.StatusOK, []byte(`<html>not a feed`)),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Update(srv.Client(), srv.URL, out)
			testify.Error(t, err)
			b, _ := os.ReadFile(out)
			testify.Equal(t, "previous", string(b))
		})
	}
}
