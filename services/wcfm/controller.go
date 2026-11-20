package wcfm

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	spinsTTL     = 5 * time.Second
	playlistsTTL = 15 * time.Second
	icecastTTL   = 5 * time.Second
	scheduleTTL  = 10 * time.Minute
	personaTTL   = 24 * time.Hour

	recentCount = 8
	// spinitron pre-creates automation playlists
	playlistCount    = 20
	scheduleDays     = 7
	maxSchedulePages = 5
	personaWorkers   = 8

	stationPersonaName = "WCFM"

	// let spinitron catch up
	boundarySlack = 3 * time.Second
	// DJs log spins late
	overdueRefresh = 8 * time.Second
	overdueWindow  = time.Minute
	staleRefresh   = time.Minute
	idleRefresh    = 30 * time.Second
	minRefresh     = 5 * time.Second
	maxRefresh     = 60 * time.Second
)

type Controller struct {
	services.BaseController
	token string
}

func NewController(cfg *config.Config, log *zap.SugaredLogger) *Controller {
	c := &Controller{BaseController: services.BaseController{Log: log}}
	if cfg.Secrets != nil {
		c.token = cfg.Secrets.SpinitronToken
	}
	return c
}

type Track struct {
	ID     int       `json:"id"`
	Song   string    `json:"song"`
	Artist string    `json:"artist,omitempty"`
	Album  string    `json:"album,omitempty"`
	Label  string    `json:"label,omitempty"`
	Year   int       `json:"year,omitempty"`
	Image  string    `json:"image,omitempty"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type Show struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	EpisodeName string    `json:"episodeName,omitempty"`
	Description string    `json:"description,omitempty"`
	Image       string    `json:"image,omitempty"`
	DJs         []string  `json:"djs"`
	Automated   bool      `json:"automated"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
}

type Now struct {
	StreamURL string  `json:"streamURL"`
	Live      bool    `json:"live"`
	Listeners *int    `json:"listeners"`
	OnAir     *Show   `json:"onAir"`
	Track     *Track  `json:"track"`
	Recent    []Track `json:"recent"`
	Next      *Show   `json:"next"`
	// seconds until the next poll
	RefreshAfter int `json:"refreshAfter"`
}

func (s Show) airing(at time.Time) bool {
	return !s.Start.After(at) && at.Before(s.End)
}

func (t *Controller) loadSpins(ctx context.Context) ([]Track, error) {
	return cached(ctx, t.Log, "spins", spinsTTL, func(ctx context.Context) ([]Track, error) {
		var body struct {
			Items []rawSpin `json:"items"`
		}
		q := url.Values{"count": {strconv.Itoa(recentCount + 1)}}
		if err := t.spinitronGet(ctx, "/spins", q, &body); err != nil {
			return nil, err
		}
		tracks := make([]Track, 0, len(body.Items))
		for _, s := range body.Items {
			tracks = append(tracks, Track{
				ID:     s.ID,
				Song:   s.Song,
				Artist: s.Artist,
				Album:  s.Release,
				Label:  s.Label,
				Year:   s.Released,
				Image:  s.Image,
				Start:  s.Start.Time,
				End:    s.End.Time,
			})
		}
		return tracks, nil
	})
}

func (t *Controller) loadPlaylists(ctx context.Context) ([]Show, error) {
	return cached(ctx, t.Log, "playlists", playlistsTTL, func(ctx context.Context) ([]Show, error) {
		var body struct {
			Items []rawPlaylist `json:"items"`
		}
		if err := t.spinitronGet(ctx, "/playlists", url.Values{"count": {strconv.Itoa(playlistCount)}}, &body); err != nil {
			return nil, err
		}
		var ids []int
		for _, p := range body.Items {
			if p.HideDJ == 0 {
				ids = append(ids, p.PersonaID)
			}
		}
		names, err := t.resolvePersonas(ctx, ids)
		if err != nil {
			return nil, err
		}

		shows := make([]Show, 0, len(body.Items))
		for _, p := range body.Items {
			var personas []int
			if p.HideDJ == 0 {
				personas = []int{p.PersonaID}
			}
			shows = append(shows, Show{
				ID:          p.ShowID,
				Title:       p.Title,
				EpisodeName: p.EpisodeName,
				Description: p.Description,
				Image:       p.Image,
				DJs:         djNames(names, personas),
				Automated:   p.Automation != 0,
				Start:       p.Start.Time,
				End:         p.End.Time,
			})
		}
		return shows, nil
	})
}

func (t *Controller) loadSchedule(ctx context.Context) ([]Show, error) {
	return cached(ctx, t.Log, "schedule", scheduleTTL, func(ctx context.Context) ([]Show, error) {
		now := time.Now().UTC()
		q := url.Values{
			"start": {now.Format(time.RFC3339)},
			"end":   {now.AddDate(0, 0, scheduleDays).Format(time.RFC3339)},
			"count": {"200"},
		}

		var raw []rawShow
		for page := 1; page <= maxSchedulePages; page++ {
			q.Set("page", strconv.Itoa(page))
			var body struct {
				Items []rawShow     `json:"items"`
				Meta  spinitronMeta `json:"_meta"`
			}
			if err := t.spinitronGet(ctx, "/shows", q, &body); err != nil {
				return nil, err
			}
			raw = append(raw, body.Items...)
			if page >= body.Meta.PageCount {
				break
			}
		}

		var ids []int
		for _, s := range raw {
			if s.HideDJ == 0 {
				ids = append(ids, s.personaIDs()...)
			}
		}
		names, err := t.resolvePersonas(ctx, ids)
		if err != nil {
			return nil, err
		}

		shows := make([]Show, 0, len(raw))
		for _, s := range raw {
			var personas []int
			if s.HideDJ == 0 {
				personas = s.personaIDs()
			}
			shows = append(shows, Show{
				ID:          s.ID,
				Title:       s.Title,
				Description: s.Description,
				Image:       s.Image,
				DJs:         djNames(names, personas),
				Start:       s.Start.Time,
				End:         s.End.Time,
			})
		}
		sort.Slice(shows, func(i, j int) bool { return shows[i].Start.Before(shows[j].Start) })
		return shows, nil
	})
}

func (t *Controller) personaName(ctx context.Context, id int) (string, error) {
	return cached(ctx, t.Log, "persona:"+strconv.Itoa(id), personaTTL, func(ctx context.Context) (string, error) {
		var p rawPersona
		if err := t.spinitronGet(ctx, "/personas/"+strconv.Itoa(id), nil, &p); err != nil {
			return "", err
		}
		return p.Name, nil
	})
}

// don't cache partial DJ lists
func (t *Controller) resolvePersonas(ctx context.Context, ids []int) (map[int]string, error) {
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		names    = map[int]string{}
		work     = make(chan int)
		firstErr error
	)
	for range personaWorkers {
		wg.Go(func() {
			for id := range work {
				name, err := t.personaName(ctx, id)
				mu.Lock()
				if err != nil && firstErr == nil {
					firstErr = fmt.Errorf("persona %d: %w", id, err)
				}
				names[id] = name
				mu.Unlock()
			}
		})
	}
	seen := map[int]bool{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			work <- id
		}
	}
	close(work)
	wg.Wait()
	return names, firstErr
}

// most shows use the station persona
func djNames(names map[int]string, ids []int) []string {
	djs := []string{}
	for _, id := range ids {
		if name := names[id]; name != "" && !strings.EqualFold(name, stationPersonaName) {
			djs = append(djs, name)
		}
	}
	return djs
}

func refreshAfter(now time.Time, track *Track, onAir, next *Show) time.Duration {
	wait := idleRefresh
	if track != nil {
		if untilEnd := track.End.Sub(now); untilEnd > 0 {
			wait = untilEnd + boundarySlack
		} else if -untilEnd < overdueWindow {
			wait = overdueRefresh
		} else {
			wait = staleRefresh
		}
	}

	var boundaries []time.Time
	if onAir != nil {
		boundaries = append(boundaries, onAir.End)
	}
	if next != nil {
		boundaries = append(boundaries, next.Start)
	}
	for _, b := range boundaries {
		if d := b.Sub(now) + boundarySlack; b.After(now) && d < wait {
			wait = d
		}
	}

	return min(max(wait, minRefresh), maxRefresh)
}

// skip automation placeholders
func scheduled(p Show, schedule []Show, now time.Time) bool {
	if !p.Automated {
		return true
	}
	for _, s := range schedule {
		if s.ID == p.ID && s.airing(now) {
			return true
		}
	}
	return false
}

func buildNow(now time.Time, spins []Track, playlists, schedule []Show, ice *icecastStatus) Now {
	resp := Now{StreamURL: StreamURL, Recent: []Track{}}
	if ice != nil {
		resp.Live = ice.Live
		resp.Listeners = &ice.Listeners
	}

	for i := range playlists {
		if playlists[i].airing(now) && scheduled(playlists[i], schedule, now) {
			resp.OnAir = &playlists[i]
			break
		}
	}
	for i := range schedule {
		if resp.OnAir == nil && schedule[i].airing(now) {
			resp.OnAir = &schedule[i]
		}
		if schedule[i].Start.After(now) {
			resp.Next = &schedule[i]
			break
		}
	}

	if len(spins) > 0 {
		resp.Track = &spins[0]
		spins = spins[1:]
	}
	if len(spins) > recentCount {
		spins = spins[:recentCount]
	}
	resp.Recent = append(resp.Recent, spins...)

	resp.RefreshAfter = int(refreshAfter(now, resp.Track, resp.OnAir, resp.Next).Seconds())
	return resp
}

// GetNow godoc
// @Summary Get what's on WCFM
// @Description get the current show, current and recent tracks, next show, and live listener count for WCFM 91.9. Poll again after refreshAfter seconds.
// @ID getWCFMNow
// @Tags wcfm
// @Produce json
// @Success 200 {object} wcfm.Now
// @Failure 502 {object} services.BaseErrorResponse "2700: unable to load WCFM data from spinitron"
// @Failure 500 {object} services.BaseErrorResponse
// @Router /wcfm/now [get]
func (t *Controller) GetNow(c *gin.Context) {
	var (
		wg                  sync.WaitGroup
		spins               []Track
		playlists, schedule []Show
		ice                 icecastStatus
		spinsErr, plErr     error
		schedErr, iceErr    error
	)
	wg.Go(func() { spins, spinsErr = t.loadSpins(c) })
	wg.Go(func() { playlists, plErr = t.loadPlaylists(c) })
	wg.Go(func() { schedule, schedErr = t.loadSchedule(c) })
	wg.Go(func() { ice, iceErr = cached(c, t.Log, "icecast", icecastTTL, fetchIcecast) })
	wg.Wait()

	if spinsErr != nil || plErr != nil {
		t.Log.Errorf("spinitron now: spins: %v, playlists: %v", spinsErr, plErr)
		t.RespondAPIError(c, lib.ErrorWCFMUnavailable)
		return
	}
	if schedErr != nil {
		t.Log.Warnf("spinitron schedule: %v", schedErr)
	}
	icePtr := &ice
	if iceErr != nil {
		t.Log.Warnf("icecast status: %v", iceErr)
		icePtr = nil
	}

	t.RespondOK(c, buildNow(time.Now().UTC(), spins, playlists, schedule, icePtr))
}

// GetSchedule godoc
// @Summary Get the WCFM schedule
// @Description get WCFM shows airing over the next week, including the one on now
// @ID getWCFMSchedule
// @Tags wcfm
// @Produce json
// @Success 200 {array} wcfm.Show
// @Failure 502 {object} services.BaseErrorResponse "2700: unable to load WCFM data from spinitron"
// @Failure 500 {object} services.BaseErrorResponse
// @Router /wcfm/schedule [get]
func (t *Controller) GetSchedule(c *gin.Context) {
	schedule, err := t.loadSchedule(c)
	if err != nil {
		t.Log.Errorf("spinitron schedule: %v", err)
		t.RespondAPIError(c, lib.ErrorWCFMUnavailable)
		return
	}

	now := time.Now()
	upcoming := []Show{}
	for _, s := range schedule {
		if s.End.After(now) {
			upcoming = append(upcoming, s)
		}
	}
	t.RespondOK(c, upcoming)
}
