package lib

import (
	"net/http"
)

type APIError struct {
	Code     int
	Message  string
	HTTPCode int
	Errors   []error
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

// This may have multiple errors, so pass them.
func NewErrorRequestDataValidationFailed(errs []error) *APIError {
	return &APIError{
		Code:     ErrorRequestDataValidationFailed.Code,
		Message:  ErrorRequestDataValidationFailed.Message,
		HTTPCode: ErrorRequestDataValidationFailed.HTTPCode,
		Errors:   errs,
	}
}

var (
	// Standard HTTP error 0404
	ErrorRecordNotFound = NewAPIErrorWithHTTP(0404, http.StatusNotFound, "record not found")
	// HTTP 500
	ErrorInternalServerError = NewAPIErrorWithHTTP(0500, http.StatusInternalServerError, "internal server error")

	// 11** are general errors
	ErrorMalformedRequestData        = NewAPIErrorWithHTTP(1100, http.StatusBadRequest, "could not parse malformed request data")
	ErrorRequestDataValidationFailed = NewAPIErrorWithHTTP(1101, http.StatusBadRequest, "request data validation failed")

	// 13** are authorization errors
	ErrorNoScopeAuthorization = NewAPIErrorWithHTTP(1330, http.StatusForbidden, "no scope authorization")
	ErrorMustBeSelf           = NewAPIErrorWithHTTP(1331, http.StatusForbidden, "must be self")

	// 14** are user service errors
	ErrorUserMustBeStudent    = NewAPIError(1401, "user must be a student")
	ErrorUserCannotBePrefrosh = NewAPIError(1402, "user cannot be prefrosh")
	ErrorUserNotVisible       = NewAPIError(1403, "user not visible")
	ErrorUserNotAtWilliams    = NewAPIError(1404, "user not at williams")
	ErrorUserIDNoParse        = NewAPIError(1405, "user id could not be parsed")
	ErrorInvalidUserTag       = NewAPIError(1406, "invalid user tag")

	// 15** are factrak errors
	// Create/Update errors
	ErrorSurveyMissingCourseParams = NewAPIError(1531, "missing course parameters in create data: courseID or (areaOfStudyAbbreviation and courseNumber)")
	ErrorSurveyCommentTooSmall     = NewAPIError(1532, "comment must be 100 characters or more")
	ErrorSurveyStudentNotFound     = NewAPIError(1533, "user must be a student and could not be found")
	ErrorSurveyProfessorNotFound   = NewAPIError(1534, "passed professor must be a professor and could not be found")
	ErrorSurveyCourseNotFound      = NewAPIError(1535, "passed course could not be found")
	ErrorSurveyAreaOfStudyNotFound = NewAPIError(1536, "passed area of study could not be found")
	ErrorSurveyAlreadyExists       = NewAPIError(1537, "survey already exists with passed user ID, professor ID, and course ID")
	// Agreement errors
	ErrorSurveyAgreementNotFound      = NewAPIErrorWithHTTP(1551, http.StatusNotFound, "survey agreement could not be found")
	ErrorSurveyAgreementAlreadyExists = NewAPIError(1552, "survey agreement already exists for this user and survey")
	ErrorSurveyAgreementNoSelf        = NewAPIError(1553, "cannot create survey agreement with your own survey")

	// 16** are dormtrak errors
	// Create/Update errors
	ErrorReviewStudentNotFound = NewAPIError(1633, "user must be a student and could not be found")
	ErrorReviewMissingDorm     = NewAPIError(1634, "user is missing dorm field")
	ErrorReviewDormNotOwner    = NewAPIError(1635, "user does not own this dorm room")
	ErrorReviewAlreadyExists   = NewAPIError(1536, "review already exists with passed user ID and dorm room ID")
)
