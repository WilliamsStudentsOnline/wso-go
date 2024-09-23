package config

import (
	"bytes"
	"fmt"
	url "net/url"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go.uber.org/zap"
)

func LoadDatabase(cfg *Config, log *zap.SugaredLogger) *gorm.DB {
	// Database params passed by config
	db, err := gorm.Open(cfg.DatabaseType, cfg.DatabaseArgs)
	if err != nil {
		log.With(zap.Error(err)).Fatal("failed to connect database")
	}

	if cfg.IsDevelopment() || cfg.IsTest() {
		db.LogMode(true)
	}

	db.SetLogger(DBLogger{
		log: log,
	})

	return db
}

// Safely closes db when done
func CloseDatabase(db *gorm.DB, log *zap.SugaredLogger) {
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}

// This function sets up the database arguments for MySQL in the config file
func SetupMySQLConfig(cfg *Config) {
	// If database arguments already set, use those
	if cfg.DatabaseArgs != "" {
		return
	}

	cfg.MySQLUser = setDefaultStr(cfg.MySQLUser, "root")
	if cfg.MySQLPort == 0 {
		cfg.MySQLPort = 3306
	}

	// Setup arguments
	qs := url.Values{}
	for key, val := range cfg.MySQLArgs {
		qs.Add(key, val)
	}

	var mysqlUrl string
	if cfg.MySQLUnix {
		mysqlUrl = fmt.Sprintf("unix(%s)", cfg.MySQLHost)
	} else {
		mysqlUrl = fmt.Sprintf("tcp(%s:%d)", cfg.MySQLHost, cfg.MySQLPort)
	}

	cfg.DatabaseArgs = fmt.Sprintf("%s:%s@%s/%s?%s",
		cfg.MySQLUser,
		cfg.Secrets.MySQLPassword,
		mysqlUrl,
		cfg.MySQLDatabase,
		qs.Encode(),
	)
}

// This function sets up the database arguments for MySQL in the config file
func SetupSQLiteConfig(cfg *Config) {
	// If database arguments already set, use those
	if cfg.DatabaseArgs != "" {
		return
	}

	cfg.DatabaseArgs = cfg.SQLiteFile
}

type DBLogger struct {
	gorm.Logger
	log *zap.SugaredLogger
}

func (l DBLogger) Print(v ...interface{}) {
	buf := new(bytes.Buffer)
	for argNum, arg := range v {
		if argNum > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(fmt.Sprint(arg))
	}
	l.log.Debug(buf.String())
}
