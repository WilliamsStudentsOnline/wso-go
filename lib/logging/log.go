package logging

import (
	"os"
	"path"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func DefaultEncoderCfg() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
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
	}
}

func SetupLog(cfg *config.Config, logName string) (*zap.SugaredLogger, error) {
	var cores []zapcore.Core
	for _, format := range cfg.LogFormats {
		switch format {
		case "log":
			logFile, err := os.OpenFile(
				path.Join(cfg.LogDirectory, logName+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return nil, err
			}
			cores = append(cores,
				zapcore.NewCore(zapcore.NewConsoleEncoder(DefaultEncoderCfg()), logFile, cfg.ParsedLogLevel()))
		case "json":
			jsonFile, err := os.OpenFile(
				path.Join(cfg.LogDirectory, logName+".json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return nil, err
			}
			cores = append(cores,
				zapcore.NewCore(zapcore.NewJSONEncoder(DefaultEncoderCfg()), jsonFile, cfg.ParsedLogLevel()))
		case "console":
			cores = append(cores,
				zapcore.NewCore(zapcore.NewConsoleEncoder(DefaultEncoderCfg()), os.Stdout, cfg.ParsedLogLevel()))
		}
	}

	// Default to stdout
	if len(cores) == 0 {
		cores = append(cores,
			zapcore.NewCore(zapcore.NewConsoleEncoder(DefaultEncoderCfg()), os.Stdout, cfg.ParsedLogLevel()))
	}

	logCore := zapcore.NewTee(cores...)
	fastLog := zap.New(logCore, zap.AddStacktrace(zapcore.ErrorLevel))

	if cfg.IsDevelopment() {
		fastLog.WithOptions(zap.Development())
	}

	log := fastLog.Sugar()
	return log, nil
}
