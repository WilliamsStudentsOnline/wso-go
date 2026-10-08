// One-shot: copy db/development.db into MySQL (Atlas schema), then emit SQL dump
// Usage: go run ./db/seed/cmd/export_mysql_dump -sqlite db/development.db -dsn '...' -out db/seed/development.sql
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	sqlitePath := flag.String("sqlite", "db/development.db", "path to SQLite development.db")
	dsn := flag.String("dsn", "", "MySQL DSN for a throwaway DB")
	out := flag.String("out", "db/seed/development.sql", "output mysqldump path")
	flag.Parse()
	if *dsn == "" {
		fmt.Fprintln(os.Stderr, "usage: export_mysql_dump -dsn 'user:pass@tcp(host:port)/db?...'")
		os.Exit(2)
	}

	mysqlDB, err := gorm.Open("mysql", *dsn)
	if err != nil {
		fatalf("mysql open: %v", err)
	}
	defer mysqlDB.Close()

	if err := dropAll(mysqlDB); err != nil {
		fatalf("drop: %v", err)
	}
	if err := migrate.MigrateDB(mysqlDB); err != nil {
		fatalf("migrate: %v", err)
	}

	sqliteDB, err := gorm.Open("sqlite3", *sqlitePath)
	if err != nil {
		fatalf("sqlite open: %v", err)
	}
	defer sqliteDB.Close()

	tables, err := mysqlTables(mysqlDB)
	if err != nil {
		fatalf("list tables: %v", err)
	}

	if err := mysqlDB.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		fatalf("fk off: %v", err)
	}
	for _, table := range tables {
		if table == "schema_migrations" {
			continue
		}
		n, err := copyTable(sqliteDB, mysqlDB, table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", table, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "copied %s (%d rows)\n", table, n)
	}
	if err := mysqlDB.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		fatalf("fk on: %v", err)
	}

	if err := runMysqldump(*dsn, *out); err != nil {
		fatalf("mysqldump: %v", err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", *out)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func dropAll(db *gorm.DB) error {
	rows, err := db.Raw(`
		SELECT table_name FROM information_schema.tables
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

func mysqlTables(db *gorm.DB) ([]string, error) {
	rows, err := db.Raw(`
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func copyTable(src, dst *gorm.DB, table string) (int, error) {
	if !src.HasTable(table) {
		return 0, fmt.Errorf("missing in sqlite")
	}

	srcCols, err := columns(src, table, true)
	if err != nil {
		return 0, err
	}
	dstMeta, err := mysqlColumnMeta(dst, table)
	if err != nil {
		return 0, err
	}
	var common []string
	var metas []colMeta
	for _, c := range srcCols {
		if m, ok := dstMeta[strings.ToLower(c)]; ok {
			common = append(common, c)
			metas = append(metas, m)
		}
	}
	if len(common) == 0 {
		return 0, fmt.Errorf("no common columns")
	}

	if err := dst.Exec("DELETE FROM `" + table + "`").Error; err != nil {
		return 0, err
	}

	colList := quoteCols(common)
	q := fmt.Sprintf("SELECT %s FROM `%s`", colList, table)
	rows, err := src.Raw(q).Rows()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	placeholders := strings.TrimRight(strings.Repeat("?,", len(common)), ",")
	insert := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, colList, placeholders)

	n := 0
	for rows.Next() {
		raw := make([]interface{}, len(common))
		ptrs := make([]interface{}, len(common))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return n, err
		}
		vals := make([]interface{}, len(common))
		for i, v := range raw {
			vals[i] = coerce(v, metas[i])
		}
		if err := dst.Exec(insert, vals...).Error; err != nil {
			return n, err
		}
		n++
	}
	return n, rows.Err()
}

type colMeta struct {
	name     string
	notNull  bool
	dataType string
}

func mysqlColumnMeta(db *gorm.DB, table string) (map[string]colMeta, error) {
	rows, err := db.Raw(`
		SELECT column_name, is_nullable, data_type
		FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ?
	`, table).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]colMeta{}
	for rows.Next() {
		var name, nullable, dataType string
		if err := rows.Scan(&name, &nullable, &dataType); err != nil {
			return nil, err
		}
		out[strings.ToLower(name)] = colMeta{
			name:     name,
			notNull:  strings.EqualFold(nullable, "NO"),
			dataType: strings.ToLower(dataType),
		}
	}
	return out, nil
}

func columns(db *gorm.DB, table string, sqlite bool) ([]string, error) {
	if sqlite {
		rows, err := db.Raw(fmt.Sprintf("PRAGMA table_info(`%s`)", table)).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var cols []string
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dflt sql.NullString
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				return nil, err
			}
			cols = append(cols, name)
		}
		return cols, nil
	}

	rows, err := db.Raw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ?
		ORDER BY ordinal_position
	`, table).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, nil
}

func quoteCols(cols []string) string {
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = "`" + c + "`"
	}
	return strings.Join(parts, ", ")
}

func coerce(v interface{}, meta colMeta) interface{} {
	if v == nil {
		if !meta.notNull {
			return nil
		}
		return notNullDefault(meta.dataType)
	}

	switch x := v.(type) {
	case []byte:
		return coerceString(string(x), meta)
	case string:
		return coerceString(x, meta)
	default:
		return v
	}
}

func coerceString(s string, meta colMeta) interface{} {
	switch s {
	case "t", "true", "TRUE", "True":
		return 1
	case "f", "false", "FALSE", "False":
		return 0
	}
	if meta.notNull && s == "" {
		return notNullDefault(meta.dataType)
	}
	return s
}

func notNullDefault(dataType string) interface{} {
	switch {
	case strings.Contains(dataType, "int"),
		dataType == "tinyint",
		dataType == "bit",
		strings.Contains(dataType, "bool"),
		strings.Contains(dataType, "decimal"),
		strings.Contains(dataType, "float"),
		strings.Contains(dataType, "double"):
		return 0
	default:
		return ""
	}
}

func runMysqldump(dsn, outPath string) error {
	// Parse user:pass@tcp(host:port)/db?...
	user, pass, host, port, dbname, err := parseDSN(dsn)
	if err != nil {
		return err
	}

	// Dump table data only (no CREATE DATABASE); import into MYSQL_DATABASE (wso)
	cmd := exec.Command("docker", "exec", "wso-mysql",
		"mysqldump",
		"-u"+user,
		"-p"+pass,
		"--single-transaction",
		"--routines=false",
		"--triggers=false",
		"--set-gtid-purged=OFF",
		dbname,
	)
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	cmd.Stdout = f
	cmd.Stderr = os.Stderr
	_ = host
	_ = port
	return cmd.Run()
}

func parseDSN(dsn string) (user, pass, host, port, db string, err error) {
	// root:pass@tcp(127.0.0.1:3306)/wso?...
	at := strings.Index(dsn, "@")
	if at < 0 {
		return "", "", "", "", "", fmt.Errorf("bad dsn")
	}
	up := dsn[:at]
	rest := dsn[at+1:]
	colon := strings.Index(up, ":")
	if colon < 0 {
		return "", "", "", "", "", fmt.Errorf("bad user/pass")
	}
	user, pass = up[:colon], up[colon+1:]

	if !strings.HasPrefix(rest, "tcp(") {
		return "", "", "", "", "", fmt.Errorf("expected tcp(...) dsn")
	}
	rest = strings.TrimPrefix(rest, "tcp(")
	end := strings.Index(rest, ")")
	if end < 0 {
		return "", "", "", "", "", fmt.Errorf("bad host")
	}
	hp := rest[:end]
	rest = rest[end+1:]
	if i := strings.Index(hp, ":"); i >= 0 {
		host, port = hp[:i], hp[i+1:]
	} else {
		host, port = hp, "3306"
	}
	rest = strings.TrimPrefix(rest, "/")
	if i := strings.IndexAny(rest, "?"); i >= 0 {
		db = rest[:i]
	} else {
		db = rest
	}
	return user, pass, host, port, db, nil
}
