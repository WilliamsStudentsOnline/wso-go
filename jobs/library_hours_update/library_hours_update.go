package library_hours_update

import (
    "encoding/json"
    "log"
    "os"
    "net/http"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
)

type ExportLibraryServices struct {
    LibraryServices map[string]LibraryService `json:"libraryservices"`
    UpdateTime string			      `json:"updateTime"`
}

type LibraryService struct {
    Name string  `json:"name"`
    Hours *Hours `json:"hours"`
}

type Hours struct {
    Open string `json:"open"`
    Close string `json:"close"`
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
    res, err := http.Get("https://libcal.williams.edu/api_hours_today.php")
    if err != nil {
	return ExportLibraryServices{}, err
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
	// TODO: return error
	log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
    }

    doc, err := goquery.NewDocumentFromReader(res.Body)
    if err != nil {
	return ExportLibraryServices{}, err
    }

    // the php script this is parsing lists the hours in this order:
    //  - Sawyer Library
    //  - Research Services
    //  - Special Collections
    //  - Schow Science Library

    var libraries = []LibraryService {
	LibraryService { Name: "Sawyer", Hours: &Hours{}, },
	LibraryService { Name: "Research Services", Hours: &Hours{}, },
	LibraryService { Name: "Special Collections", Hours: &Hours{}, },
	LibraryService { Name: "Schow", Hours: &Hours{}, },
    }

    index := 0

    doc.Find("tr").Each(func(i int, s *goquery.Selection) {
	// TODO: could make the order not hardcoded by uncommenting
	// and using the service variable below

	/*
	service := s.Find("th").Text()
	service = strings.TrimSpace(service)
	*/

	hours := s.Find("td").Text()

	h := strings.Split(hours, "–")
	// the page uses two different dash symbols, so splitting with
	// the above doesn't work for every library listing
	if len(h) == 1 {
	    h = strings.Split(hours, "-")
	}
	open := strings.TrimSpace(h[0])
	close := strings.TrimSpace(h[1])

	libraries[index].Hours.Open = open
	libraries[index].Hours.Close = close

	index = index + 1
    })

    els := ExportLibraryServices{
	LibraryServices: make(map[string]LibraryService),
    }
    els.LibraryServices["sawyer"] = libraries[0]
    els.LibraryServices["research"] = libraries[1]
    els.LibraryServices["collections"] = libraries[2]
    els.LibraryServices["schow"] = libraries[3]
    els.UpdateTime = time.Now().Format(time.RFC850)

    return els, err
}
