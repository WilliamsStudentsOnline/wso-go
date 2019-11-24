package config

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewDefaultProductionLog(cfg *Config) (*zap.SugaredLogger, error) {
	logCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:      false,
		Sampling:         nil,
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	if cfg != nil {
		logCfg.Level = zap.NewAtomicLevelAt(cfg.ParsedLogLevel())
	}

	fastLog, err := logCfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return fastLog.Sugar(), nil
}

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
		errorCode := c.GetInt("ErrorCodeKey")
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
		)

		if errorCode > 0 {
			entry.With("errorCode", errorCode)
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
		if statusCode >= 500 {
			entry.Error(msg)
		} else if statusCode >= 400 {
			entry.Warn(msg)
		} else {
			entry.Info(msg)
		}
	}
}
