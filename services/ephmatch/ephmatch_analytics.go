package ephmatch

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ephmatchOpenGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wso_backend_ephmatch_open",
		Help: "The current status of Ephmatch (1 for open, 0 for closed)",
	})
	ephmatchSeniorsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wso_backend_ephmatch_seniors_open",
		Help: "The current status of seniors Ephmatch (1 for open, 0 for closed)",
	})
	ephmatchProfileMakeCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_ephmatch_profile_creations",
		Help: "The number of new profiles created on Ephmatch",
	})
	ephmatchProfileDeleteCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_ephmatch_profile_deletions",
		Help: "The number of profiles deleted on Ephmatch",
	})
	ephmatchMatchesUnmatchCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_ephmatch_matches_unmatches",
		Help: "The number of times a user has unmatched another",
	})
	ephmatchMatchesMatchSummaryTotal = promauto.NewSummary(prometheus.SummaryOpts{
		Name:       "wso_backend_ephmatch_matches_summary_total",
		Help:       "The 50th, 90th, and 99th percentile of total match numbers amongst users",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
	ephmatchMatchesMatchSummaryUnseen = promauto.NewSummary(prometheus.SummaryOpts{
		Name:       "wso_backend_ephmatch_matches_summary_unseen",
		Help:       "The 50th, 90th, and 99th percentile of unseen match numbers amongst users",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
)
