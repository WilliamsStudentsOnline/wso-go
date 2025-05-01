package library_hours_update

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ExportLibraryServices struct {
	LibraryServices map[string]LibraryService `json:"libraryservices"`
	UpdateTime      string                    `json:"updateTime"`
}

type LibraryService struct {
	Name  string `json:"name"`
	Hours *Hours `json:"hours"`
}

type Hours struct {
	Open  []string `json:"open"`
	Close []string `json:"close"`
}

func UpdateLibraryHours(outPath string) error {
	var els ExportLibraryServices
	var err error
	els, err = loadLibraryHours()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(els)
}

func loadLibraryHours() (ExportLibraryServices, error) {
	// use custom transport to skip TLS verification if CA isn't trusted
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "https://libcal.williams.edu/api_hours_today.php", nil)
	if err != nil {
		return ExportLibraryServices{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; WSO-Backend/1.0; +https://github.com/WilliamsStudentsOnline/wso-go)")
	res, err := client.Do(req)

	if err != nil {
		return ExportLibraryServices{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ExportLibraryServices{}, fmt.Errorf("unexpected status: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return ExportLibraryServices{}, err
	}

	services := []string{
		"sawyer",
		"research",
		"collections",
		"schow",
	}

	els := ExportLibraryServices{
		LibraryServices: make(map[string]LibraryService, len(services)),
	}

	index := 0
	doc.Find("tr").Each(func(_ int, row *goquery.Selection) {
		if index >= len(services) {
			return
		}
		// extract text including possible multiple ranges
		raw := row.Find("td").Text()
		// split on newlines to get each interval
		lines := strings.Split(raw, "\n")

		var opens, closes []string
		for _, line := range lines {
			interval := strings.TrimSpace(line)
			if interval == "" {
				continue
			}
			// split on en-dash or hyphen
			parts := strings.SplitN(interval, "–", 2)
			if len(parts) == 1 {
				parts = strings.SplitN(interval, "-", 2)
			}
			if len(parts) < 2 {
				continue
			}
			opens = append(opens, strings.TrimSpace(parts[0]))
			closes = append(closes, strings.TrimSpace(parts[1]))
		}

		els.LibraryServices[services[index]] = LibraryService{
			Name: strings.Title(services[index]),
			Hours: &Hours{
				Open:  opens,
				Close: closes,
			},
		}
		index++
	})

	els.UpdateTime = time.Now().Format(time.RFC850)
	return els, nil
}
