package logging

import (
	"fmt"

	"github.com/coreos/go-systemd/v22/journal"
	"go.uber.org/zap/zapcore"
)

func newJournalCore(enab zapcore.LevelEnabler, ident string) (zapcore.Core, error) {
	if !journal.Enabled() {
		return nil, fmt.Errorf("journald is not available on this host")
	}

	// journald already records timestamp and priority; keep the message readable
	encCfg := DefaultEncoderCfg()
	encCfg.TimeKey = ""
	encCfg.LevelKey = ""

	return &journalCore{
		LevelEnabler: enab,
		encoder:      zapcore.NewConsoleEncoder(encCfg),
		ident:        ident,
	}, nil
}

type journalCore struct {
	zapcore.LevelEnabler
	encoder zapcore.Encoder
	ident   string
}

func (c *journalCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.encoder = c.encoder.Clone()
	for i := range fields {
		fields[i].AddTo(clone.encoder)
	}
	return &clone
}

func (c *journalCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *journalCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.encoder.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	msg := buf.String()
	buf.Free()

	return journal.Send(msg, zapLevelToJournal(ent.Level), map[string]string{
		"SYSLOG_IDENTIFIER": c.ident,
	})
}

func (c *journalCore) Sync() error {
	return nil
}

func zapLevelToJournal(level zapcore.Level) journal.Priority {
	switch level {
	case zapcore.DebugLevel:
		return journal.PriDebug
	case zapcore.InfoLevel:
		return journal.PriInfo
	case zapcore.WarnLevel:
		return journal.PriWarning
	case zapcore.ErrorLevel:
		return journal.PriErr
	case zapcore.DPanicLevel, zapcore.PanicLevel:
		return journal.PriCrit
	case zapcore.FatalLevel:
		return journal.PriCrit
	default:
		return journal.PriInfo
	}
}
