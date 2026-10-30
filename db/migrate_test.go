package db_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/assert"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// use shared db while avoiding circular import error
func TestMigrateMySQLUpToDate(t *testing.T) {
	assert := assert.New(t)

	host := envOr("WSO_TEST_MYSQL_HOST", "127.0.0.1")
	user := envOr("WSO_TEST_MYSQL_USER", "root")
	pass := envOr("WSO_TEST_MYSQL_PASSWORD", "secret-mysql-password")
	port, err := strconv.Atoi(envOr("WSO_TEST_MYSQL_PORT", "3307"))
	assert.NoError(err)

	dbName := fmt.Sprintf("wso_migrate_test_%d", os.Getpid())
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&charset=utf8mb4&multiStatements=true",
		user, pass, host, port)

	root, err := gorm.Open("mysql", rootDSN)
	if err != nil {
		t.Fatalf("test MariaDB unreachable at %s:%d (run `make itest`): %v", host, port, err)
	}
	defer root.Close()

	assert.NoError(root.Exec(fmt.Sprintf("CREATE DATABASE `%s`", dbName)).Error)
	defer root.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&multiStatements=true",
		user, pass, host, port, dbName)
	db, err := gorm.Open("mysql", dsn)
	assert.NoError(err)
	defer db.Close()

	assert.NoError(migrate.MigrateDB(db))

	ok, err := migrate.MySQLMigrationsUpToDate(db)
	assert.NoError(err)
	assert.True(ok)
}
