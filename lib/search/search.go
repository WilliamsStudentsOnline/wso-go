package search

import (
	"strings"
)

const (
	SearchBackendSQL = "sql"
)

func IsInvalidTokenError(err error) bool {
	return strings.Contains(err.Error(), "invalid token")
}
