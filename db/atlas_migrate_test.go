package db

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Opt-in MySQL smoke test (needs a real server; default `go test` stays on SQLite)
// ATLAS_TEST_DSN='root:secret-mysql-password@tcp(127.0.0.1:3306)/wso_test?parseTime=true&multiStatements=true'
func TestMigrateMySQLFreshAndStamp(t *testing.T) {
	dsn := os.Getenv("ATLAS_TEST_DSN")
	if dsn == "" {
		t.Skip("ATLAS_TEST_DSN not set")
	}

	db, err := gorm.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, dropAllMySQLTables(db))

	require.NoError(t, MigrateDB(db))
	up, err := MigrationUpToDate(MigrationGormOptions, db)
	require.NoError(t, err)
	assert.True(t, up)

	// idempotent
	require.NoError(t, MigrateDB(db))

	// pre-Atlas shape: tables present, no schema_migrations → stamp
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS schema_migrations").Error)
	require.NoError(t, MigrateDB(db))
	up, err = MigrationUpToDate(MigrationGormOptions, db)
	require.NoError(t, err)
	assert.True(t, up)
}

func dropAllMySQLTables(db *gorm.DB) error {
	rows, err := db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'
	`).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return err
	}
	for _, t := range tables {
		if err := db.Exec("DROP TABLE IF EXISTS `" + t + "`").Error; err != nil {
			return err
		}
	}
	return db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error
}
