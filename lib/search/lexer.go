/*
Search query parsing (mostly for Users).

Query Parsing Notes:
 - By default, terms are connected together by AND

Changes from Rails:
 - Operators are "AND" and "OR", not "&" and "|"
 - Removed the "," operator; just use "OR"
 - Removed ability to do "name: (Ephraim, Hopkins)", that is: removed shorthand to search multiple field values.
   Instead, do this "name: Ephraim OR name: Hopkins"
*/
package search

import (
	"strconv"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

var queryLexer = lexer.Must(lexer.Regexp(
	`(\s+)` +
		`|(?P<Keyword>AND|OR)` +
		`|(?P<Word>[a-zA-Z][\w-]*)` +
		`|(?P<String>"(?:\\.|[^"])*")` +
		`|(?P<Int>\d+)` +
		`|(?P<Punct>[\(:\)])`,
))

var parser *participle.Parser

func init() {
	var err error
	parser, err = participle.Build(&Query{},
		participle.Lexer(queryLexer),
		participle.Unquote("String"))
	if err != nil {
		panic(err)
	}
}

type Query struct {
	// We must specify OR for it to work
	Or []*Expression `(@@ { "OR" @@ })*`
}

type Expression struct {
	// Either we can have "AND" or blank for it to be an AND
	And []*Condition `@@ { "AND"? @@ }`
}

type Condition struct {
	Field *Field `(@@`
	Value *Value `| @@`
	//SubQuery *Query `| "(" @@ ")")!`
	Or []*Expression `| "(" @@ { "OR" @@ } ")")!`
	// To add NOT, do add this to the middle/end: (and add NOT as a keyword)
	//Not *Condition `| "NOT" @@`
}

type Field struct {
	Key   string `@Word`
	Value *Value `":"@@`
}

type Value struct {
	Word   *string `  @Word`
	String *string `| @String`
	Int    *int    `| @Int`
}

func (v *Value) ToString() string {
	if v.String != nil {
		return *v.String
	}
	if v.Word != nil {
		return *v.Word
	}
	if v.Int != nil {
		return strconv.Itoa(*v.Int)
	}

	return ""
}

func ParseSearchQuery(q string) (*Query, error) {
	ast := &Query{}
	err := parser.ParseString(q, ast)
	if err != nil {
		return nil, err
	}

	return ast, nil
}
