package daily_messages_update

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testNow = time.Date(2026, 11, 20, 7, 15, 0, 0, eastern)

func TestUpdateWritesFiles(t *testing.T) {
	feed, err := os.ReadFile("testdata/feed.json")
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(feed)
	}))
	defer srv.Close()

	dir := t.TempDir()
	today := filepath.Join(dir, "today.json")
	require.NoError(t, Update(srv.Client(), srv.URL, dir, today, testNow))
	require.NoError(t, Update(srv.Client(), srv.URL, dir, today, testNow))

	for _, p := range []string{filepath.Join(dir, "2026-11-20.json"), today} {
		b, err := os.ReadFile(p)
		require.NoError(t, err)
		var out Export
		require.NoError(t, json.Unmarshal(b, &out))
		assert.Equal(t, "2026-11-20", out.Date)
		assert.Len(t, out.Categories, 8)
	}
	target, err := os.Readlink(today)
	require.NoError(t, err)
	assert.Equal(t, "2026-11-20.json", target)
}

func TestUpdateKeepsFileOnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	dir := t.TempDir()
	existing := filepath.Join(dir, "2026-11-20.json")
	require.NoError(t, os.WriteFile(existing, []byte("old"), 0644))

	assert.Error(t, Update(srv.Client(), srv.URL, dir, "", testNow))
	b, err := os.ReadFile(existing)
	require.NoError(t, err)
	assert.Equal(t, "old", string(b))
}

func TestUpdateNamesFileByFeedDate(t *testing.T) {
	feed, err := os.ReadFile("testdata/feed.json")
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(feed)
	}))
	defer srv.Close()

	dir := t.TempDir()
	afterMidnight := time.Date(2026, 11, 21, 0, 5, 0, 0, eastern)
	require.NoError(t, Update(srv.Client(), srv.URL, dir, "", afterMidnight))
	_, err = os.Stat(filepath.Join(dir, "2026-11-20.json"))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "2026-11-21.json"))
	assert.True(t, os.IsNotExist(err))
}

func TestUpdateSkipsWithoutFailing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"A": [{"ID": 1, "type": "announcement"}, {"ID": {}}]}`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	require.NoError(t, Update(srv.Client(), srv.URL, dir, "", testNow))
	_, err := os.Stat(filepath.Join(dir, "2026-11-20.json"))
	assert.NoError(t, err)
}
