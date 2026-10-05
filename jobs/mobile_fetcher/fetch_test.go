package mobile_fetcher

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFetchSuccessWritesFile(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	outDir := t.TempDir()
	f := NewFetcher(outDir, srv.Client())

	err := f.Fetch(Job{
		Name:   "spins",
		URL:    srv.URL,
		Output: "spins.json",
		Headers: map[string]string{
			"Authorization": "Bearer abc",
		},
	})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if sawAuth != "Bearer abc" {
		t.Fatalf("Authorization header = %q", sawAuth)
	}

	got, err := os.ReadFile(filepath.Join(outDir, "spins.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"ok":true}` {
		t.Fatalf("file contents = %q", got)
	}
}

func TestFetchNon2xxKeepsExistingFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "boom")
	}))
	defer srv.Close()

	outDir := t.TempDir()
	path := filepath.Join(outDir, "spins.json")
	if err := os.WriteFile(path, []byte("good-data"), 0644); err != nil {
		t.Fatal(err)
	}

	f := NewFetcher(outDir, srv.Client())
	err := f.Fetch(Job{Name: "spins", URL: srv.URL, Output: "spins.json"})
	if err == nil {
		t.Fatal("expected error for non-2xx")
	}
	if !strings.Contains(err.Error(), "non-2xx") {
		t.Fatalf("error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "good-data" {
		t.Fatalf("existing file overwritten: %q", got)
	}
}

func TestFetchIncompleteDownloadKeepsExistingFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "partial")
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("server does not support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		_ = conn.Close()
	}))
	defer srv.Close()

	outDir := t.TempDir()
	path := filepath.Join(outDir, "spins.json")
	if err := os.WriteFile(path, []byte("good-data"), 0644); err != nil {
		t.Fatal(err)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	f := NewFetcher(outDir, client)
	err := f.Fetch(Job{Name: "spins", URL: srv.URL, Output: "spins.json"})
	if err == nil {
		t.Fatal("expected error for incomplete download")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "good-data" {
		t.Fatalf("existing file overwritten: %q", got)
	}
}

func TestFetchAllContinuesAfterFailure(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if r.URL.Path == "/bad" {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	outDir := t.TempDir()
	f := NewFetcher(outDir, srv.Client())
	f.FetchAll([]Job{
		{Name: "bad", URL: srv.URL + "/bad", Output: "bad.json"},
		{Name: "good", URL: srv.URL + "/good", Output: "good.json"},
	})

	if _, err := os.Stat(filepath.Join(outDir, "bad.json")); !os.IsNotExist(err) {
		t.Fatalf("bad.json should not exist, err=%v", err)
	}
	got, err := os.ReadFile(filepath.Join(outDir, "good.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "ok" {
		t.Fatalf("good.json = %q", got)
	}
	if n != 2 {
		t.Fatalf("expected 2 requests, got %d", n)
	}
}
