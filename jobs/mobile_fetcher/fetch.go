package mobile_fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultHTTPTimeout = 20 * time.Second
	maxErrorBodyBytes  = 512
)

type Fetcher struct {
	client    *http.Client
	outputDir string
}

func NewFetcher(outputDir string, client *http.Client) *Fetcher {
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return &Fetcher{
		client:    client,
		outputDir: outputDir,
	}
}

func (f *Fetcher) FetchAll(jobs []Job) {
	for _, job := range jobs {
		if err := f.Fetch(job); err != nil {
			log.Printf("[%s] fetch failed: %v", job.Name, err)
			continue
		}
		log.Printf("[%s] wrote %s", job.Name, filepath.Join(f.outputDir, job.Output))
	}
}

func (f *Fetcher) Fetch(job Job) error {
	req, err := http.NewRequest(http.MethodGet, job.URL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	for k, v := range job.Headers {
		req.Header.Set(k, v)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
		return fmt.Errorf("non-2xx status %d: %s", resp.StatusCode, truncateForLog(snippet))
	}

	dest := filepath.Join(f.outputDir, job.Output)
	if err := atomicWrite(dest, resp.Body); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	return nil
}

func atomicWrite(dest string, r io.Reader) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(dest)+".tmp.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(tmp, r); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("copy body: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	cleanup = false
	return nil
}

func truncateForLog(b []byte) string {
	s := string(b)
	if len(s) == 0 {
		return "(empty body)"
	}
	return s
}
