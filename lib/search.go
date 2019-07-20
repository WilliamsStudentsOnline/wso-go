package lib

import (
	"regexp"
	"strings"
)

type SearchQuery struct {
	query string

}

var (
	SymbolAnd = regexp.MustCompilePOSIX(`AND`)
	SymbolOr = regexp.MustCompilePOSIX(`OR`)
	SymbolField = regexp.MustCompilePOSIX(`.+:.+`)
)

type QueryField struct {
	Field string
	Value string
}

func NewSearchQuery(query string) *SearchQuery {
	return &SearchQuery{
		query: query,
	}
}

func (q *SearchQuery) Parse() {
	strings.Split(q.query, " ")
}