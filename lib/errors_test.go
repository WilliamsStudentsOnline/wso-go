package lib

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// parsed rather than reflected: package vars can't be enumerated
func TestAPIErrorCodesUnique(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	seen := map[int][]string{}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			fn, ok := call.Fun.(*ast.Ident)
			if !ok || (fn.Name != "NewAPIError" && fn.Name != "NewAPIErrorWithHTTP") {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.INT {
				return true
			}
			code, err := strconv.Atoi(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			seen[code] = append(seen[code], fset.Position(call.Pos()).String())
			return true
		})
	}

	if len(seen) == 0 {
		t.Fatal("found no API error codes")
	}
	codes := make([]int, 0, len(seen))
	for code := range seen {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		if locs := seen[code]; len(locs) > 1 {
			t.Errorf("error code %d used %d times: %s", code, len(locs), strings.Join(locs, ", "))
		}
	}
}
