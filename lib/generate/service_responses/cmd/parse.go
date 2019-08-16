package main

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/iancoleman/strcase"
)

var TypeMapping = map[string]string{
	"boolean": "bool",
	"number":  "float64",
	"integer": "int",
	"string":  "string",
	"null":    "interface{/*nil*/}",
	"array":   "[]interface{}",
	"object":  "map[string]interface{}",
}

type schemaRoot struct {
	root     *Schema
	prefix   string
	packages map[string]bool
}

func ParseSchema(r io.Reader) (*Schema, error) {
	var root Schema
	err := json.NewDecoder(r).Decode(&root)
	if err != nil {
		return nil, err
	}

	return &root, nil
}

func GenerateStructs(root *Schema, prefix string) ([]StructType, []string, Entry) {
	r := schemaRoot{
		root:     root,
		prefix:   prefix,
		packages: make(map[string]bool),
	}

	st, entry := r.ParseRoot()
	var pkgs []string
	for pkg := range r.packages {
		pkgs = append(pkgs, pkg)
	}

	return st, pkgs, entry
}

func (r *schemaRoot) ParseRoot() ([]StructType, Entry) {
	var st []StructType
	for defName, def := range r.root.Definitions {
		st = append(st, r.ParseDef(def, defName))
	}

	entry := Entry{}

	if r.root.Ref != "" {
		name := getRefName(r.root.Ref)
		def := r.ParseDef(r.root.Definitions[name], name)
		entry.IsSlice = false
		entry.Name = def.Name
		entry.GoConvName = def.GoConvName
	} else if r.root.Type.Type == "array" {
		entry.IsSlice = true
		ref := r.root.Items.Ref
		name := getRefName(ref)
		def := r.ParseDef(r.root.Definitions[name], name)
		entry.Name = def.Name
		entry.GoConvName = def.GoConvName
	} else {
		panic("unsupported root type")
	}

	// Sort
	sort.Slice(st, func(i int, j int) bool {
		return st[i].Name < st[j].Name
	})

	return st, entry
}

func (r *schemaRoot) ParseDef(t *Type, name string) StructType {
	structType := StructType{}
	structType.Name = r.prefix + strcase.ToCamel(name)

	// Only support object definitions currently
	if t.Type == "object" {
		var structFields []StructField
		for propName, prop := range t.Properties {
			sf := r.ParseProperty(prop, propName)
			if !lib.StringsContains(t.Required, propName) && prop.GoPtr == nil {
				sf.IsPointer = true
			}
			structFields = append(structFields, sf)
		}
		structType.Fields = structFields
	} else {
		panic("only supports definitions of type object")
	}

	structType.GoConvName = name
	if t.GoOtherName != "" {
		structType.GoConvName = t.GoOtherName
	}
	if t.GoOtherPackage != "" {
		splts := strings.Split(t.GoOtherPackage, "/")
		structType.GoConvName = splts[len(splts)-1] + "." + structType.GoConvName
		r.packages[t.GoOtherPackage] = true
	}

	// Sort
	sort.Slice(structType.Fields, func(i int, j int) bool {
		return structType.Fields[i].Name < structType.Fields[j].Name
	})

	return structType
}

func (r *schemaRoot) ParseProperty(t *Type, name string) StructField {
	structField := StructField{}
	structField.Name = strcase.ToCamel(name)
	if t.GoName != "" {
		structField.Name = t.GoName
	}
	structField.JSONTag = name

	if t.GoType != "" {
		// If type is specified in as a goType
		structField.Type = t.GoType
	} else if t.Ref != "" {
		// If type is reference to a global definition (aka struct)
		structField.IsStruct = true
		structField.Type = strcase.ToCamel(getRefName(t.Ref))
		structField.Type = r.prefix + structField.Type
		structField.JSONTag += ",omitempty"
	} else if t.Type == "array" {
		// If type is an array
		structField.IsSlice = true
		arrField := r.ParseProperty(t.Items, name)
		structField.IsStruct = arrField.IsStruct
		structField.Type = arrField.Type
		structField.JSONTag += ",omitempty"
	} else {
		// Otherwise try to cast it
		structField.Type = TypeMapping[t.Type]
	}

	if t.GoPtr != nil {
		structField.IsPointer = *t.GoPtr
	}

	structField.GoConvName = structField.Name
	if t.GoOtherName != "" {
		structField.GoConvName = t.GoOtherName
	}
	if t.GoOtherPackage != "" {
		splts := strings.Split(t.GoOtherPackage, "/")
		structField.GoConvName = splts[len(splts)-1] + "." + structField.GoConvName
		r.packages[t.GoOtherPackage] = true
	}

	return structField
}

func getRefName(ref string) string {
	return ref[len("#/definitions/"):]
}
