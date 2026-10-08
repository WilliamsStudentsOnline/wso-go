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
var atlasMigrationsFS embed.FS

const atlasMigrationsPath = "atlas/migrations"

// baselineVersion is the first Atlas migration. Existing MySQL databases that
// were created via gormigrate/AutoMigrate are stamped at this version without
// re-running CREATE TABLE statements.
const baselineVersion = uint(20261007235756)

// MigrateMySQL applies versioned SQL migrations (Atlas-planned, golang-migrate format).
func MigrateMySQL(db *gorm.DB) error {
	sqlDB := db.DB()
	if sqlDB == nil {
		return errors.New("mysql: underlying sql.DB is nil")
	}

	m, err := newMySQLMigrate(sqlDB)
	if err != nil {
		return err
	}
	// Do not Close() — that closes the shared sql.DB owned by GORM.

	if err := maybeStampBaseline(sqlDB, m); err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("atlas migrate up: %w", err)
	}
	return nil
}

// maybeStampBaseline marks the baseline migration as applied when the DB
// already has application tables (legacy gormigrate/AutoMigrate) but no
// golang-migrate version yet. Fresh empty databases skip this and run Up().
func maybeStampBaseline(sqlDB *sql.DB, m *migrate.Migrate) error {
	_, _, err := m.Version()
	if err == nil {
		return nil
	}
	if !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("atlas migrate version: %w", err)
	}

	var usersCount int
	if err := sqlDB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = 'users'
	`).Scan(&usersCount); err != nil {
		return fmt.Errorf("atlas baseline probe: %w", err)
	}
	if usersCount == 0 {
		return nil
	}

	if err := m.Force(int(baselineVersion)); err != nil {
		return fmt.Errorf("atlas stamp baseline %d: %w", baselineVersion, err)
	}
	return nil
}

func newMySQLMigrate(sqlDB *sql.DB) (*migrate.Migrate, error) {
	source, err := iofs.New(atlasMigrationsFS, atlasMigrationsPath)
	if err != nil {
		return nil, fmt.Errorf("atlas migrations source: %w", err)
	}

	// Caller DSN must include multiStatements=true (see SetupMySQLConfig).
	driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("atlas mysql driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("atlas migrate init: %w", err)
	}
	return m, nil
}

// MySQLMigrationsUpToDate reports whether all embedded Atlas migrations are applied.
func MySQLMigrationsUpToDate(db *gorm.DB) (bool, error) {
	sqlDB := db.DB()
	if sqlDB == nil {
		return false, errors.New("mysql: underlying sql.DB is nil")
	}

	m, err := newMySQLMigrate(sqlDB)
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

	source, err := iofs.New(atlasMigrationsFS, atlasMigrationsPath)
	if err != nil {
		return false, err
	}
	_, err = source.Next(version)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
