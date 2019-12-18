package search

import (
	"strings"

	"github.com/alecthomas/participle/lexer"
)

const (
	SearchBackendSQL = "sql"
)

func IsInvalidTokenError(err error) bool {
	return strings.Contains(err.Error(), "invalid token")
}

func IsQueryError(err error) bool {
	_, ok := err.(*lexer.Error)
	return ok
	//return strings.Contains(err.Error(), "<source>")
}
