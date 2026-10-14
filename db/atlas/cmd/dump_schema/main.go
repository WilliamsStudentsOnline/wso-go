// used by `make atlas-schema`
package main

import (
	"flag"
	"fmt"
	"os"

	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	dsn := flag.String("dsn", "", "MySQL DSN, e.g. root:pass@tcp(127.0.0.1:3306)/wso_atlas?parseTime=true&charset=utf8mb4")
	flag.Parse()
	if *dsn == "" {
		fmt.Fprintln(os.Stderr, "usage: dump_schema -dsn 'user:pass@tcp(host:port)/db?parseTime=true'")
		os.Exit(2)
	}

	db, err := gorm.Open("mysql", *dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open mysql: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// fresh-start db
	if err := dropAllTables(db); err != nil {
		fmt.Fprintf(os.Stderr, "drop tables: %v\n", err)
		os.Exit(1)
	}

	if err := migrate.AutoMigrateModels(db); err != nil {
		fmt.Fprintf(os.Stderr, "automigrate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("ok")
}

func dropAllTables(db *gorm.DB) error {
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
	if len(tables) == 0 {
		return nil
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
