package logging

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/coreos/go-systemd/v22/journal"
	"go.uber.org/zap/zapcore"
)

func TestZapLevelToJournal(t *testing.T) {
	cases := []struct {
		level zapcore.Level
		want  journal.Priority
	}{
		{zapcore.DebugLevel, journal.PriDebug},
		{zapcore.InfoLevel, journal.PriInfo},
		{zapcore.WarnLevel, journal.PriWarning},
		{zapcore.ErrorLevel, journal.PriErr},
		{zapcore.DPanicLevel, journal.PriCrit},
		{zapcore.PanicLevel, journal.PriCrit},
		{zapcore.FatalLevel, journal.PriCrit},
	}
	for _, tc := range cases {
		if got := zapLevelToJournal(tc.level); got != tc.want {
			t.Fatalf("zapLevelToJournal(%v) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

func TestSetupLogJournalUnavailable(t *testing.T) {
	if journal.Enabled() {
		t.Skip("journald is available; cannot assert hard-fail path")
	}

	cfg := &config.Config{
		LogLevel:   "info",
		LogFormats: []string{"journal"},
	}
	_, err := SetupLog(cfg, "wso-backend")
	if err == nil {
		t.Fatal("expected error when journal format is requested without journald")
	}
}

func TestSetupLogConsoleStillWorks(t *testing.T) {
	cfg := &config.Config{
		LogLevel:   "info",
		LogFormats: []string{"console"},
	}
	log, err := SetupLog(cfg, "wso-backend")
	if err != nil {
		t.Fatalf("SetupLog: %v", err)
	}
	log.Sync()
}
