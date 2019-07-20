package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "Migration Helper"
	app.HideVersion = true
	app.Action = runApp
	app.Usage = "a gorm migrations generator"
	app.HelpName = "migrate"

	app.UsageText = "migrate [options] migration_title"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "path, p",
			Value: filepath.Join("db", "migrations"),
			Usage: "path to migrations folder",
		},
		cli.StringFlag{
			Name:  "model, m",
			Value: "model",
			Usage: "go model name that we are migrating",
		},
		cli.StringFlag{
			Name:  "table, t",
			Value: "table",
			Usage: "sql table name of model that we are migrating",
		},
		cli.StringFlag{
			Name:  "action, a",
			Usage: "migration action (eg create_table, etc)",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("Error:", err.Error())
	}

}

func runApp(c *cli.Context) error {
	title := c.Args().First()
	path := c.String("path")

	if title == "" {
		return errors.New("a title is required")
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// This is year, month, day, hour, minute, and second in UTC time
	migrationTime := time.Now().UTC().Format("20060102150405")
	migrationName := migrationTime + "_" + title

	// To camel case: replace _ with " ", titleize every word, replace " " with nothing
	camelTitle := strings.ReplaceAll(strings.Title(strings.ReplaceAll(title, "_", " ")), " ", "")

	// Migration path
	path = filepath.Join(path, migrationName+".go")

	// Template the new file
	tmpl, err := template.New("migration").Parse(migrationTmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, map[string]interface{}{
		"package":       filepath.Base(filepath.Dir(path)),
		"name":          migrationName,
		"model":         strings.Title(c.String("model")),
		"table":         strings.ToLower(c.String("table")),
		"camelTitle":    camelTitle,
		"migrationTime": migrationTime,
	})
	if err != nil {
		panic(err)
	}

	return nil
}

var migrationTmpl = `package {{.package}}

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var {{.camelTitle}}{{.migrationTime}} = &gormigrate.Migration{
	ID: "{{.name}}",
	Migrate: func(tx *gorm.DB) error {
		// It's a good pratice to copy the struct inside the function,
		// so side effects are prevented if the original struct changes during the time.
		// But, when the table already exists, it just adds new fields as columns, so just have a struct
		// with those fields.
		type {{.model}} struct {
		}
		return tx.AutoMigrate(&{{.model}}{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.DropTable("{{.table}}").Error
	},
}
`
