package search

// I'm convinced everything will break, so keep this here.
/*
var queryLexer = lexer.Must(lexer.Regexp(
	`(\s+)` +
		`|(?P<Keyword>AND|OR)` +
		`|(?P<Word>[a-zA-Z][\w-]*)` +
		`|(?P<String>"(?:\\.|[^"])*")` +
		`|(?P<Punct>[\(:\)])`,
))


type Query struct {
	Fields []*Field `(@@`
	//Values      []*Value `| @@)*`
	Values         []*Value         `| @@`
	AndExpressions []*AndExpression `| "AND" @@`
	OrExpressions []*OrExpression `| "OR" @@`
	SubQueries []*Query `| "(" @@ ")")*`
}

type Expression struct {
	And []*AndOtherExpression `@@ { "OR" @@ }`
}

type OrExpression struct {
	And []*Condition `@@ { "AND" @@ }`
}

type AndExpression struct {
	Or []*Condition `@@ { "OR" @@ }`
}

type Condition struct {
	Field    *Field `(@@`
	Value    *Value `| @@`
	SubQuery *Query `| "(" @@ ")")!`
}

type Field struct {
	Key   string `@Word`
	Value *Value `":"@@`
}

type Value struct {
	Value  *string `  @Word`
	String *string `| @String`
}
*/
