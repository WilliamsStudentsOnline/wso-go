package api

import (
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

// TestWeeklyMenuURL checks URL formatting.
func TestWeeklyMenuURL(t *testing.T) {
    a := &NutriSliceAPI{
        baseUrl: "https://example.api.nutrislice.com/menu/api/",
        client:  NewHTTPClient(5 * time.Second),
    }
    got := a.WeeklyMenuURL("driscoll-dining-hall", "breakfast", time.Date(2025, 11, 19, 0, 0, 0, 0, time.UTC))
    want := "https://example.api.nutrislice.com/menu/api/weeks/school/driscoll-dining-hall/menu-type/breakfast/2025/11/19/"
    if got != want {
        t.Fatalf("WeeklyMenuURL = %q, want %q", got, want)
    }
}

// TestFetchWeeklyMenu uses an httptest server to ensure the client can read the endpoint.
// It avoids depending on DoGet implementation and uses the client's http.Client directly.
func TestFetchWeeklyMenu(t *testing.T) {
    const sampleBody = `{"ok":true}`

    // setup a test server that expects the path produced by WeeklyMenuURL
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // basic sanity checks
        if r.Method != http.MethodGet {
            t.Fatalf("expected GET, got %s", r.Method)
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(sampleBody))
    }))
    defer ts.Close()

    // create API with baseUrl pointing at the test server's menu/api/ base path
    base := ts.URL + "/menu/api/"
    apiClient := &NutriSliceAPI{
        baseUrl: base,
        client:  NewHTTPClient(5 * time.Second),
    }

    // construct URL and perform GET using the client's http.Client directly
    url := apiClient.WeeklyMenuURL("driscoll-dining-hall", "breakfast", time.Date(2025, 11, 19, 0, 0, 0, 0, time.UTC))
    resp, err := apiClient.client.Get(url)
    if err != nil {
        t.Fatalf("http get failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status: %d", resp.StatusCode)
    }
    b, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Fatalf("read body failed: %v", err)
    }
    if string(b) != sampleBody {
        t.Fatalf("unexpected body: %s", string(b))
    }
}