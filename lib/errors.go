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

	// 11** are general errors
	ErrorMalformedRequestData = NewAPIErrorWithHTTP(1100, http.StatusBadRequest, "could not parse malformed request data")

	// 13** are authorization errors
	ErrorNoScopeAuthorization = NewAPIErrorWithHTTP(1330, http.StatusForbidden, "no scope authorization")
	ErrorMustBeSelf           = NewAPIErrorWithHTTP(1331, http.StatusForbidden, "must be self")

	// 14** are user errors
	ErrorUserMustBeStudent    = NewAPIError(1401, "user must be a student")
	ErrorUserCannotBePrefrosh = NewAPIError(1402, "user cannot be prefrosh")
	ErrorUserNotVisible       = NewAPIError(1403, "user not visible")
	ErrorUserNotAtWilliams    = NewAPIError(1404, "user not at williams")

	// 15** are factrak errors
	ErrorSurveyMissingCourseParams = NewAPIError(1531, "missing course parameters in create data: courseID or (areaOfStudyAbbreviation and courseNumber)")
	ErrorCommentTooSmall           = NewAPIError(1532, "comment must be 100 characters or more")
	ErrorSurveyStudentNotFound     = NewAPIError(1533, "user must be a student and could not be found")
	ErrorSurveyProfessorNotFound   = NewAPIError(1534, "passed professor must be a professor and could not be found")
	ErrorSurveyCourseNotFound      = NewAPIError(1535, "passed course could not be found")
	ErrorSurveyAreaOfStudyNotFound      = NewAPIError(1536, "passed area of study could not be found")
	ErrorSurveyAlreadyExists = NewAPIError(1537, "survey already exists with passed user ID, professor ID, and course ID")
)
