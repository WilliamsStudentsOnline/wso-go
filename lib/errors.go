package lib

import (
	"net/http"
)

type APIError struct {
	Code     int
	Message  string
	HTTPCode int
}

// Create a new API error. Default HTTP code is 400 Bad Request.
func NewAPIError(code int, message string) *APIError {
	return NewAPIErrorWithHTTP(code, http.StatusBadRequest, message)
}

// Create a new API error with a HTTP code.
func NewAPIErrorWithHTTP(code int, httpCode int, message string) *APIError {
	return &APIError{
		Code:     code,
		Message:  message,
		HTTPCode: httpCode,
	}
}

func (e *APIError) Error() string {
	return e.Message
}

var (
	// Standard HTTP error 0404
	ErrorRecordNotFound = NewAPIErrorWithHTTP(0404, http.StatusNotFound, "record not found")
	// HTTP 500
	ErrorInternalServerError = NewAPIErrorWithHTTP(0500, http.StatusInternalServerError, "internal server error")

	// 13** are authorization errors
	ErrorNoScopeAuthorization = NewAPIErrorWithHTTP(1330, http.StatusForbidden, "no scope authorization")
	ErrorMustBeSelf           = NewAPIErrorWithHTTP(1331, http.StatusForbidden, "must be self")

	// 14** are user errors
	ErrorUserMustBeStudent    = NewAPIError(1401, "user must be a student")
	ErrorUserCannotBePrefrosh = NewAPIError(1402, "user cannot be prefrosh")
	ErrorUserNotVisible       = NewAPIError(1403, "user not visible")
	ErrorUserNotAtWilliams    = NewAPIError(1404, "user not at williams")
)
