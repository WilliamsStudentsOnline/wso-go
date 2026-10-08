package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jinzhu/gorm"
)

//go:embed atlas/migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "atlas/migrations"

// LUCA (first atlas version)
// (we were previously on gorm)
const firstAtlasVersion = uint(20261007235756)

// MigrateMySQL applies embedded Atlas SQL via golang-migrate
func MigrateMySQL(db *gorm.DB) error {
	sqlDB := db.DB()
	if sqlDB == nil {
		return errors.New("mysql: underlying sql.DB is nil")
	}

	m, err := newMigrate(sqlDB)
	if err != nil {
		return err
	}
	// Don't Close — that closes GORM's sql.DB

	if err := stampLegacyIfNeeded(sqlDB, m); err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("atlas migrate up: %w", err)
	}
	return nil
}

// stampLegacyIfNeeded Force()s firstAtlasVersion when tables exist but
// schema_migrations does not (pre-Atlas / gormigrate DBs)
// Empty DBs skip this and run Up() instead
func stampLegacyIfNeeded(sqlDB *sql.DB, m *migrate.Migrate) error {
	_, _, err := m.Version()
	if err == nil {
		return nil
	}
	if !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("atlas migrate version: %w", err)
	}

	var n int
	err = sqlDB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = 'users'
	`).Scan(&n)
	if err != nil {
		return fmt.Errorf("atlas baseline probe: %w", err)
	}
	if n == 0 {
		return nil
	}

	if err := m.Force(int(firstAtlasVersion)); err != nil {
		return fmt.Errorf("atlas stamp baseline %d: %w", firstAtlasVersion, err)
	}
	return nil
}

func newMigrate(sqlDB *sql.DB) (*migrate.Migrate, error) {
	src, err := iofs.New(migrationsFS, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("atlas migrations source: %w", err)
	}

	// DSN needs multiStatements=true (SetupMySQLConfig)
	driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("atlas mysql driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("atlas migrate init: %w", err)
	}
	return m, nil
}

// MySQLMigrationsUpToDate is true when no pending Atlas migrations remain
func MySQLMigrationsUpToDate(db *gorm.DB) (bool, error) {
	sqlDB := db.DB()
	if sqlDB == nil {
		return false, errors.New("mysql: underlying sql.DB is nil")
	}

	m, err := newMigrate(sqlDB)
	if err != nil {
		return false, err
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return false, nil
		}
		return false, err
	}
	if dirty {
		return false, fmt.Errorf("mysql migrations dirty at version %d", version)
	}

	src, err := iofs.New(migrationsFS, migrationsDir)
	if err != nil {
		return false, err
	}
	_, err = src.Next(version)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
