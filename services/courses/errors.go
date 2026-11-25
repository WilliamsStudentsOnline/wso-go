package courses

import "errors"

var (
	errBadStrm  = errors.New("strm must be a term code like 1271")
	errBadCache = errors.New("unexpected cached value")
)
