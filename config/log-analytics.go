package config

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	pathRequest = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wso_backend_path_requested",
		Help: "The total number of times each endpoint is contacted",
	},
		[]string{"endpoint", "method"},
	)
	requestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_request_processed_total",
		Help: "The total number of responses",
	})
	successCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_success_processed_total",
		Help: "The total number of successful responses",
	})
	errorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_error_processed_total",
		Help: "The total number of error responses",
	})
	warnCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_warn_processed_total",
		Help: "The total number of warning responses",
	})

	latencyGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wso_backend_latency_measurement",
		Help: "Latency of requests/replies",
	})

	mobileCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_mobile_device_total",
		Help: "Number of requests from mobile browsers",
	})
	desktopCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_desktop_device_total",
		Help: "Number of requests from desktop browsers",
	})
	botCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_bot_device_total",
		Help: "Number of requests from bots and scrapers",
	})

	safariCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_safari_browser_total",
		Help: "Number of requests from Safari",
	})
	firefoxCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_firefox_browser_total",
		Help: "Number of requests from Firefox",
	})
	edgeCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_edge_browser_total",
		Help: "Number of requests from Edge",
	})
	chromeCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_chrome_browser_total",
		Help: "Number of requests from Chrome",
	})
	ieCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_ie_browser_total",
		Help: "Number of requests from Internet Explorer",
	})
	operaCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wso_backend_opera_browser_total",
		Help: "Number of requests from Opera",
	})
)
