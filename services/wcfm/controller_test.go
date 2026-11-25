package wcfm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

var fixtureNow = time.Date(2026, 11, 20, 6, 15, 0, 0, time.UTC)

var upstream = map[string]string{
	"/api/spins": `{"items":[
		{"id":3,"start":"2026-11-20T01:13:51-0500","end":"2026-11-20T01:17:22-0500","artist":"Sun Kil Moon","song":"Carry Me Ohio","release":"Ghosts of the Great Highway","released":2003},
		{"id":2,"start":"2026-11-20T01:09:00-0500","end":"2026-11-20T01:13:40-0500","artist":"Slowdive","song":"Catch the Breeze"},
		{"id":1,"start":"2026-11-20T01:05:00-0500","end":"2026-11-20T01:08:50-0500","artist":"Big Thief","song":"Pretty Things"}]}`,
	"/api/playlists": `{"items":[
		{"id":11,"show_id":102,"persona_id":1,"automation":1,"start":"2026-11-20T07:00:00+0000","end":"2026-11-20T08:00:00+0000","title":"91.9 WCFM Williamstown 11/20/26, 2:00 AM"},
		{"id":10,"show_id":101,"persona_id":2,"automation":0,"start":"2026-11-20T06:00:00+0000","end":"2026-11-20T07:00:00+0000","title":"Tunes and Turpentine","episode_name":"Night Shift"}]}`,
	"/api/shows": `{"_meta":{"pageCount":1},"items":[
		{"id":103,"start":"2026-11-20T12:00:00+0000","end":"2026-11-20T13:00:00+0000","title":"Unsupervised","_links":{"personas":[{"href":"https://spinitron.com/api/personas/1"}]}},
		{"id":101,"start":"2026-11-20T06:00:00+0000","end":"2026-11-20T07:00:00+0000","title":"Tunes and Turpentine","_links":{"personas":[{"href":"https://spinitron.com/api/personas/2"}]}}]}`,
	"/api/personas/1":  `{"id":1,"name":"WCFM"}`,
	"/api/personas/2":  `{"id":2,"name":"DJ Moss"}`,
	"/status-json.xsl": `{"icestats":{"source":{"listenurl":"http://x:8000/stream","listeners":14}}}`,
}

func fakeUpstream(t *testing.T, status int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status-json.xsl" && r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, ok := upstream[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
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

func TestBuildNow(t *testing.T) {
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
	// newest playlist hasn't started yet
	if assert.NotNil(now.OnAir) {
		assert.Equal("Tunes and Turpentine", now.OnAir.Title)
		assert.Equal("Night Shift", now.OnAir.EpisodeName)
		assert.Equal([]string{"DJ Moss"}, now.OnAir.DJs)
	}
	if assert.NotNil(now.Track) {
		assert.Equal("Carry Me Ohio", now.Track.Song)
		assert.Equal(time.Date(2026, 11, 20, 6, 13, 51, 0, time.UTC), now.Track.Start)
	}
	assert.Len(now.Recent, 2)
	if assert.NotNil(now.Next) {
		assert.Equal("Unsupervised", now.Next.Title)
		assert.Empty(now.Next.DJs)
	}
	assert.Equal(60, now.RefreshAfter)

	later := buildNow(fixtureNow.Add(time.Hour), spins, playlists, schedule, &ice)
	assert.Nil(later.OnAir)
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

func TestGetNow_UpstreamDown(t *testing.T) {
	assert := testify.New(t)
	fakeUpstream(t, http.StatusServiceUnavailable)
	router := utils.SetupRouter()
	router.GET("/now", newTestController(t).GetNow)

	w, err := utils.DoHTTPReq(router, http.MethodGet, "/now", nil)
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorWCFMUnavailable)
}

func TestCached_RedisHitThenStaleFallback(t *testing.T) {
	assert := testify.New(t)
	utils.RequireRedis(t, ConfigureControllerForTest())
	ctx := context.Background()
	log := zaptest.NewLogger(t).Sugar()
	key := redis_util.GetWCFMCacheKey("test")
	t.Cleanup(func() {
		_ = store.Client().Del(ctx, key, redis_util.GetWCFMCacheKey("stale:test"))
		store = nil
	})

	calls := 0
	good := func(context.Context) (string, error) {
		calls++
		return "last good", nil
	}
	for range 2 {
		got, err := cached(ctx, log, "test", time.Minute, good)
		assert.NoError(err)
		assert.Equal("last good", got)
	}
	assert.Equal(1, calls)

	assert.NoError(store.Client().Del(ctx, key))
	down := func(context.Context) (string, error) { return "", errors.New("spinitron down") }
	got, err := cached(ctx, log, "test", time.Minute, down)
	assert.NoError(err)
	assert.Equal("last good", got)
}
