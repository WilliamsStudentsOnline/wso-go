package config

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
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
)
