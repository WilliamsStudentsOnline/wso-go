package config

import (
	"bytes"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

func LoadDatabase(cfg *Config) *gorm.DB {
	// Database params passed by config
	db, err := gorm.Open(cfg.DatabaseType, cfg.DatabaseArgs)
	if err != nil {
		log.WithError(err).Fatal("failed to connect database")
	}

	if cfg.IsDevelopment() || cfg.IsTest() {
		db.LogMode(true)
	}

	db.SetLogger(DBLogger{})

	return db
}

// Safely closes DB when done
func CloseDatabase(db *gorm.DB) {
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}

type DBLogger struct {
	gorm.Logger
}

func (DBLogger) Print(v ...interface{}) {
	buf := new(bytes.Buffer)
	for argNum, arg := range v {
		if argNum > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(fmt.Sprint(arg))
	}
	log.Debug(buf.String())
}