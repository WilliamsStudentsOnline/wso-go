package wcfm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// Fixtures were captured from spinitron while this moment was on air
var fixtureNow = time.Date(2026, 11, 20, 6, 15, 0, 0, time.UTC)

const automationPersona = `{"id":182353,"name":"WCFM"}`

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func fakeUpstream(t *testing.T, status int) {
	t.Helper()
	routes := map[string][]byte{
		"/api/spins":       fixture(t, "spins.json"),
		"/api/playlists":   fixture(t, "playlists.json"),
		"/api/shows":       fixture(t, "shows.json"),
		"/status-json.xsl": fixture(t, "icecast.json"),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status-json.xsl" && r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, ok := routes[r.URL.Path]
		if strings.HasPrefix(r.URL.Path, "/api/personas/") {
			body, ok = []byte(automationPersona), true
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	oldBase, oldIce, oldStore := spinitronBaseURL, icecastStatusURL, store
	spinitronBaseURL, icecastStatusURL, store = srv.URL+"/api", srv.URL+"/status-json.xsl", nil
	t.Cleanup(func() { spinitronBaseURL, icecastStatusURL, store = oldBase, oldIce, oldStore })
}

func newTestController(t *testing.T) *Controller {
	cfg := &config.Config{Secrets: &config.Secrets{SpinitronToken: "test-token"}}
	return NewController(cfg, zaptest.NewLogger(t).Sugar())
}

func TestBuildNow_Fixtures(t *testing.T) {
	assert := testify.New(t)
	fakeUpstream(t, http.StatusOK)
	c := newTestController(t)
	ctx := context.Background()

	spins, err := c.loadSpins(ctx)
	assert.NoError(err)
	playlists, err := c.loadPlaylists(ctx)
	assert.NoError(err)
	schedule, err := c.loadSchedule(ctx)
	assert.NoError(err)
	ice, err := fetchIcecast(ctx)
	assert.NoError(err)

	now := buildNow(fixtureNow, spins, playlists, schedule, &ice)

	assert.True(now.Live)
	assert.Equal(14, *now.Listeners)
	assert.Equal(StreamURL, now.StreamURL)

	if assert.NotNil(now.OnAir) {
		assert.Equal("Tunes and Turpentine", now.OnAir.Title)
		assert.True(now.OnAir.Automated)
		assert.Empty(now.OnAir.DJs)
	}
	if assert.NotNil(now.Track) {
		assert.Equal("Sun Kil Moon", now.Track.Artist)
		assert.Equal(time.Date(2026, 11, 20, 6, 13, 51, 0, time.UTC), now.Track.Start)
	}
	assert.Len(now.Recent, recentCount)
	if assert.NotNil(now.Next) {
		assert.Equal("Unsupervised", now.Next.Title)
	}
	assert.Equal(60, now.RefreshAfter)

	// the newest playlist is pre-created automation that hasn't started yet
	earlier := buildNow(fixtureNow.Add(-30*time.Minute), spins, playlists, []Show{playlists[1]}, &ice)
	if assert.NotNil(earlier.OnAir) {
		assert.Equal("Sailing Lubitz", earlier.OnAir.Title)
	}
}

func TestBuildNow_StaleTrackStaysUp(t *testing.T) {
	assert := testify.New(t)
	start := fixtureNow.Add(-20 * time.Minute)
	spins := []Track{{ID: 1, Start: start, End: start.Add(4 * time.Minute)}}

	now := buildNow(fixtureNow, spins, nil, nil, nil)

	if assert.NotNil(now.Track) {
		assert.Equal(1, now.Track.ID)
	}
	assert.Nil(now.Listeners)
	assert.Empty(now.Recent)
	assert.Equal(int(staleRefresh.Seconds()), now.RefreshAfter)
}

func TestRefreshAfter(t *testing.T) {
	now := fixtureNow
	track := func(endsIn time.Duration) *Track { return &Track{End: now.Add(endsIn)} }
	show := func(at time.Duration) *Show { return &Show{Start: now.Add(at), End: now.Add(at)} }

	tests := []struct {
		name  string
		track *Track
		onAir *Show
		next  *Show
		want  time.Duration
	}{
		{"no data", nil, nil, nil, idleRefresh},
		{"song ending soon", track(20 * time.Second), nil, nil, 23 * time.Second},
		{"song ending later is capped", track(4 * time.Minute), nil, nil, maxRefresh},
		{"song overdue", track(-30 * time.Second), nil, nil, overdueRefresh},
		{"song long overdue", track(-5 * time.Minute), nil, nil, staleRefresh},
		{"song ending right now is floored", track(0), nil, nil, overdueRefresh},
		{"show ends before song", track(50 * time.Second), show(10 * time.Second), nil, 13 * time.Second},
		{"next show starts before song ends", track(50 * time.Second), nil, show(20 * time.Second), 23 * time.Second},
		{"past boundaries ignored", track(40 * time.Second), show(-time.Minute), nil, 43 * time.Second},
		{"floor", track(time.Second), nil, nil, minRefresh},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testify.Equal(t, tt.want, refreshAfter(now, tt.track, tt.onAir, tt.next))
		})
	}
}

func TestFetchIcecast_SourceShapes(t *testing.T) {
	tests := []struct {
		name string
		body string
		want icecastStatus
	}{
		{"single mount", `{"icestats":{"source":{"listenurl":"http://x:8000/stream","listeners":4}}}`, icecastStatus{Live: true, Listeners: 4}},
		{"several mounts", `{"icestats":{"source":[{"listenurl":"http://x:8000/other","listeners":9},{"listenurl":"http://x:8000/stream","listeners":2}]}}`, icecastStatus{Live: true, Listeners: 2}},
		{"off air", `{"icestats":{"admin":"x"}}`, icecastStatus{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			old := icecastStatusURL
			icecastStatusURL = srv.URL
			defer func() { icecastStatusURL = old }()

			got, err := fetchIcecast(context.Background())
			testify.NoError(t, err)
			testify.Equal(t, tt.want, got)
		})
	}
}

func TestEndpoints_Public(t *testing.T) {
	assert := testify.New(t)
	fakeUpstream(t, http.StatusOK)
	router := utils.SetupRouter()
	c := newTestController(t)
	router.GET("/now", c.GetNow)
	router.GET("/schedule", c.GetSchedule)

	w, err := utils.DoHTTPReq(router, http.MethodGet, "/now", nil)
	assert.NoError(err)
	var now Now
	assert.NoError(json.Unmarshal(utils.GetGoodResp(assert, w).Data, &now))
	assert.Equal(StreamURL, now.StreamURL)
	assert.Positive(now.RefreshAfter)

	// fixture shows are in the past by the time tests run, so only the shape is checked
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/schedule", nil)
	assert.NoError(err)
	var schedule []Show
	assert.NoError(json.Unmarshal(utils.GetGoodResp(assert, w).Data, &schedule))
	assert.NotNil(schedule)
}

func TestGetNow_UpstreamDown(t *testing.T) {
	assert := testify.New(t)
	fakeUpstream(t, http.StatusServiceUnavailable)
	router := utils.SetupRouter()
	router.GET("/now", newTestController(t).GetNow)

	w, err := utils.DoHTTPReq(router, http.MethodGet, "/now", nil)
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorWCFMUnavailable)
}

func TestCached_UsesRedis(t *testing.T) {
	assert := testify.New(t)
	if err := ConfigureControllerForTest(); err != nil {
		if os.Getenv("GITHUB_ACTIONS") == "" {
			t.Skipf("redis not reachable: %v", err)
		}
		t.Fatal(err)
	}
	ctx := context.Background()
	log := zaptest.NewLogger(t).Sugar()
	t.Cleanup(func() {
		_ = store.Client().Del(ctx, redis_util.GetWCFMCacheKey("test"), redis_util.GetWCFMCacheKey("stale:test"))
		store = nil
	})

	calls := 0
	fetch := func(context.Context) ([]Track, error) {
		calls++
		return []Track{{ID: calls, Song: "song"}}, nil
	}
	first, err := cached(ctx, log, "test", time.Minute, fetch)
	assert.NoError(err)
	second, err := cached(ctx, log, "test", time.Minute, fetch)
	assert.NoError(err)

	assert.Equal(1, calls)
	assert.Equal(first, second)
}

func TestDJNames_DropsStationPersona(t *testing.T) {
	names := map[int]string{1: "WCFM", 2: "DJ Moss", 3: ""}
	testify.Equal(t, []string{"DJ Moss"}, djNames(names, []int{1, 2, 3, 4}))
}

func TestBuildNow_UnscheduledAutomationIsNotAShow(t *testing.T) {
	assert := testify.New(t)
	start := fixtureNow.Add(-10 * time.Minute)
	auto := Show{ID: 1, Title: "91.9 WCFM Williamstown 11/20/26, 2:00AM", Automated: true, Start: start, End: start.Add(time.Hour)}
	scheduledAuto := Show{ID: 2, Title: "Tunes and Turpentine", Automated: true, Start: start, End: start.Add(time.Hour)}
	next := Show{ID: 3, Title: "Unsupervised", Start: start.Add(5 * time.Hour), End: start.Add(6 * time.Hour)}

	now := buildNow(fixtureNow, nil, []Show{auto}, []Show{next}, nil)
	assert.Nil(now.OnAir)
	assert.Equal("Unsupervised", now.Next.Title)

	now = buildNow(fixtureNow, nil, []Show{scheduledAuto}, []Show{scheduledAuto, next}, nil)
	if assert.NotNil(now.OnAir) {
		assert.Equal("Tunes and Turpentine", now.OnAir.Title)
	}
}

func TestCached_ServesStaleWhenFetchFails(t *testing.T) {
	assert := testify.New(t)
	if err := ConfigureControllerForTest(); err != nil {
		if os.Getenv("GITHUB_ACTIONS") == "" {
			t.Skipf("redis not reachable: %v", err)
		}
		t.Fatal(err)
	}
	ctx := context.Background()
	log := zaptest.NewLogger(t).Sugar()
	t.Cleanup(func() {
		_ = store.Client().Del(ctx, redis_util.GetWCFMCacheKey("stale-test"), redis_util.GetWCFMCacheKey("stale:stale-test"))
		store = nil
	})

	good := func(context.Context) (string, error) { return "last good", nil }
	_, err := cached(ctx, log, "stale-test", time.Minute, good)
	assert.NoError(err)
	assert.NoError(store.Client().Del(ctx, redis_util.GetWCFMCacheKey("stale-test")))

	down := func(context.Context) (string, error) { return "", errors.New("spinitron down") }
	got, err := cached(ctx, log, "stale-test", time.Minute, down)
	assert.NoError(err)
	assert.Equal("last good", got)
}

func TestCached_SharesConcurrentMisses(t *testing.T) {
	assert := testify.New(t)
	store = nil
	log := zaptest.NewLogger(t).Sugar()

	var calls atomic.Int32
	release := make(chan struct{})
	fetch := func(context.Context) (int, error) {
		calls.Add(1)
		<-release
		return 7, nil
	}

	var wg sync.WaitGroup
	results := make([]int, 10)
	for i := range results {
		wg.Go(func() {
			results[i], _ = cached(context.Background(), log, "shared-test", time.Minute, fetch)
		})
	}
	waiting := func() int {
		flightsMu.Lock()
		defer flightsMu.Unlock()
		if f, ok := flights[redis_util.GetWCFMCacheKey("shared-test")]; ok {
			return f.waiters
		}
		return 0
	}
	assert.Eventually(func() bool { return waiting() == len(results)-1 }, time.Second, time.Millisecond)
	close(release)
	wg.Wait()

	assert.Equal(int32(1), calls.Load())
	for _, r := range results {
		assert.Equal(7, r)
	}
}

func TestShared_PanicReleasesWaiters(t *testing.T) {
	assert := testify.New(t)
	func() {
		defer func() { _ = recover() }()
		_, _ = shared("panic-test", func() ([]byte, error) { panic("boom") })
	}()

	done := make(chan struct{})
	go func() {
		payload, err := shared("panic-test", func() ([]byte, error) { return []byte("ok"), nil })
		assert.NoError(err)
		assert.Equal("ok", string(payload))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("key stayed locked after a panicking fetch")
	}
}
