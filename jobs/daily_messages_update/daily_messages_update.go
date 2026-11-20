package daily_messages_update

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	FeedURL         = "https://events.williams.edu/wp-json/wms/events/v1/list/dm"
	userAgent       = "Mozilla/5.0 (compatible; WSO-Backend/1.0; +https://github.com/WilliamsStudentsOnline/wso-go)"
	maxFeedBytes    = 10 << 20
	outputFileMode  = 0644
	maxErrBodyBytes = 512
)

func Update(client *http.Client, url, outDir, todayLink string, now time.Time) error {
	raw, err := fetch(client, url)
	if err != nil {
		return err
	}

	date := now.In(eastern).Format("2006-01-02")
	out, skipped, err := Normalize(raw, date, now)
	if err != nil {
		return fmt.Errorf("normalize: %w", err)
	}
	b, err := json.Marshal(out)
	if err != nil {
		return err
	}

	dest := filepath.Join(outDir, out.Date+".json")
	if err := atomicWrite(dest, b); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	if todayLink != "" {
		target, err := filepath.Rel(filepath.Dir(todayLink), dest)
		if err != nil {
			return err
		}
		if err := atomicSymlink(target, todayLink); err != nil {
			return fmt.Errorf("link %s: %w", todayLink, err)
		}
	}
	for _, err := range skipped {
		log.Printf("skipped: %v", err)
	}
	return nil
}

func fetch(client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(res.Body, maxErrBodyBytes))
		return nil, fmt.Errorf("unexpected status %d: %s", res.StatusCode, snippet)
	}
	return io.ReadAll(io.LimitReader(res.Body, maxFeedBytes))
}

func atomicWrite(dest string, b []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".tmp.*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(outputFileMode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dest)
}

func atomicSymlink(target, link string) error {
	tmp := fmt.Sprintf("%s.tmp.%d", link, os.Getpid())
	_ = os.Remove(tmp)
	if err := os.Symlink(target, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, link)
}
