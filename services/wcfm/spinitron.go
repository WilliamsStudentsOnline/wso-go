package wcfm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const StreamURL = "https://wcfm-streaming.williams.edu:8443/stream"

var (
	spinitronBaseURL = "https://spinitron.com/api"
	// 8443 only has TLS_RSA ciphers
	icecastStatusURL = "http://wcfm-streaming.williams.edu:8000/status-json.xsl"
	httpClient       = &http.Client{Timeout: 5 * time.Second}
)

// no colon in the offset, so not RFC3339
const spinitronTimeLayout = "2006-01-02T15:04:05-0700"

type spinTime struct{ time.Time }

func (t *spinTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil || s == "" {
		return err
	}
	parsed, err := time.Parse(spinitronTimeLayout, s)
	if err != nil {
		return err
	}
	t.Time = parsed.UTC()
	return nil
}

type spinitronMeta struct {
	PageCount int `json:"pageCount"`
}

type rawSpin struct {
	ID       int      `json:"id"`
	Start    spinTime `json:"start"`
	End      spinTime `json:"end"`
	Artist   string   `json:"artist"`
	Song     string   `json:"song"`
	Release  string   `json:"release"`
	Label    string   `json:"label"`
	Released int      `json:"released"`
	Image    string   `json:"image"`
}

type rawPlaylist struct {
	ID          int      `json:"id"`
	PersonaID   int      `json:"persona_id"`
	ShowID      int      `json:"show_id"`
	Start       spinTime `json:"start"`
	End         spinTime `json:"end"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	EpisodeName string   `json:"episode_name"`
	Image       string   `json:"image"`
	Automation  int      `json:"automation"`
	HideDJ      int      `json:"hide_dj"`
}

type rawShow struct {
	ID          int      `json:"id"`
	Start       spinTime `json:"start"`
	End         spinTime `json:"end"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	HideDJ      int      `json:"hide_dj"`
	Links       struct {
		Personas []struct {
			Href string `json:"href"`
		} `json:"personas"`
	} `json:"_links"`
}

func (s rawShow) personaIDs() []int {
	var ids []int
	for _, p := range s.Links.Personas {
		if id, err := strconv.Atoi(path.Base(p.Href)); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

type rawPersona struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (t *Controller) spinitronGet(ctx context.Context, endpoint string, query url.Values, out any) error {
	u := spinitronBaseURL + endpoint
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+t.token)
	return doJSON(req, out)
}

func doJSON(req *http.Request, out any) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("%s: status %d: %s", req.URL.Path, resp.StatusCode, snippet)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type icecastStatus struct {
	Live      bool `json:"live"`
	Listeners int  `json:"listeners"`
}

type icecastSource struct {
	ListenURL string `json:"listenurl"`
	Listeners int    `json:"listeners"`
}

func fetchIcecast(ctx context.Context) (icecastStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, icecastStatusURL, nil)
	if err != nil {
		return icecastStatus{}, err
	}
	var body struct {
		Icestats struct {
			Source json.RawMessage `json:"source"`
		} `json:"icestats"`
	}
	if err := doJSON(req, &body); err != nil {
		return icecastStatus{}, err
	}

	// object, array, or missing depending on mount count
	raw := bytes.TrimSpace(body.Icestats.Source)
	var sources []icecastSource
	switch {
	case len(raw) == 0:
	case raw[0] == '[':
		err = json.Unmarshal(raw, &sources)
	default:
		var s icecastSource
		err = json.Unmarshal(raw, &s)
		sources = append(sources, s)
	}
	if err != nil {
		return icecastStatus{}, err
	}

	for _, s := range sources {
		if path.Base(s.ListenURL) == path.Base(StreamURL) {
			return icecastStatus{Live: true, Listeners: s.Listeners}, nil
		}
	}
	return icecastStatus{}, nil
}
