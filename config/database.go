package config

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

func LoadDatabase(cfg *Config) *gorm.DB {
	// Database params passed by config
	db, err := gorm.Open(cfg.DatabaseType, cfg.DatabaseArgs)
	if err != nil {
		log.Fatalln("failed to connect database")
	}

	if cfg.IsDevelopment() || cfg.IsTest() {
		db.LogMode(true)
	}

	db.SetLogger(log.StandardLogger())

	return db
}

// Safely closes DB when done
func CloseDatabase(db *gorm.DB) {
	if err := db.Close(); err != nil {
		log.Fatalln(err)
	}
}
