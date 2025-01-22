package config

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
	"go.uber.org/zap"
)

// Logger is the zap logger handler
func Logger(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		start := time.Now()

		c.Next()

		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		errorCode := c.GetInt(services.ErrorCodeKey)
		userID, _ := c.Get("id")

		entry := log.With(
			"statusCode", statusCode,
			"latency", latency,
			"clientIP", clientIP,
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"userAgent", clientUserAgent,
			"userID", userID,
			"handler", c.HandlerName(),
		)

		if errorCode > 0 {
			entry = entry.With("errorCode", errorCode)
		}

		queryUrl := query
		if queryUrl != "" {
			queryUrl = "?" + queryUrl
		}

		// Print errors if it is an error
		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		}

		msg := fmt.Sprintf("%s %s %d (%dms)",
			c.Request.Method,
			path+queryUrl,
			statusCode,
			latency)

		// Prometheus hooks and error messages
		requestCounter.Inc()
		pathRequest.WithLabelValues(path, query)

		// Log error code of messages
		if statusCode >= 500 {
			errorCounter.Inc()
			entry.Error(msg)
		} else if statusCode >= 400 {
			warnCounter.Inc()
			entry.Warn(msg)
		} else {
			successCounter.Inc()
			entry.Info(msg)
		}

		// Device type
		ua := useragent.Parse(clientUserAgent)
		if ua.Mobile {
			mobileCounter.Inc()
		} else if ua.Bot {
			botCounter.Inc()
		} else if ua.Desktop {
			desktopCounter.Inc()
		} else {
			if strings.Contains(ua.Name, "curl") || strings.Contains(ua.Name, "wget") {
				botCounter.Inc()
			} else {
				log.Warn("Unknown user agent detected")
			}
		}

		// Browser type
		switch ua.Name {
		case "Safari":
			safariCounter.Inc()
		case "Internet Explorer":
			ieCounter.Inc()
		case "Edge":
			edgeCounter.Inc()
		case "Firefox":
			firefoxCounter.Inc()
		case "Chrome":
			chromeCounter.Inc()
		case "Opera":
			operaCounter.Inc()
		}

		latencyGauge.Set(float64(latency))
		// TODO: histogram for the request URL maybe?
	}
}
