package main

import (
	"bytes"
	"flag"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/iancoleman/strcase"
)

func main() {
	var inPath string
	var outPath string

	flag.StringVar(&inPath, "in", "", "JSON Schema file as input")
	flag.StringVar(&outPath, "out", "", "Generated file as output")
	flag.Parse()

	if inPath == "" {
		panic("missing in path")
	}
	if outPath == "" {
		panic("missing out path")
	}

	inPath, err := filepath.Abs(inPath)
	if err != nil {
		panic(err)
	}

	inFile, err := os.Open(inPath)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	schema, err := ParseSchema(inFile)
	if err != nil {
		panic(err)
	}

	prefix := strcase.ToCamel(schema.Title)
	if prefix == "" {
		panic("prefix is empty in JSON Schema")
	}

	structs, packages, entry := GenerateStructs(schema, prefix)

	buf := new(bytes.Buffer)

	t := template.Must(template.New("").Parse(StructTemplate))
	err = t.Execute(buf, struct {
		Timestamp time.Time
		Prefix    string
		Structs   []StructType
		Packages  []string
		Entry     Entry
	}{
		Timestamp: time.Now(),
		Prefix:    prefix,
		Structs:   structs,
		Packages:  packages,
		Entry:     entry,
	})
	if err != nil {
		panic(err)
	}

	p, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	_, err = outFile.Write(p)
	if err != nil {
		panic(err)
	}
}
