package lib

import (
	"fmt"
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

// This may have multiple errors, so pass them.
func NewErrorMalformedRequestData(err error) *APIError {
	return &APIError{
		Code:     ErrorMalformedRequestData.Code,
		Message:  ErrorMalformedRequestData.Message,
		HTTPCode: ErrorMalformedRequestData.HTTPCode,
		Errors:   []error{err},
	}
}

func NewErrorInvalidSearchToken(err error) *APIError {
	errStr := err.Error()
	return &APIError{
		Code:     ErrorInvalidSearchToken.Code,
		Message:  fmt.Sprintf("invalid search token %q", errStr[len(errStr)-2]),
		HTTPCode: ErrorInvalidSearchToken.HTTPCode,
	}
}

func NewErrorUnknownSearchField(field string) *APIError {
	return &APIError{
		Code:     ErrorUnknownSearchField.Code,
		Message:  fmt.Sprintf("unknown search field %q", field),
		HTTPCode: ErrorUnknownSearchField.HTTPCode,
	}
}

func NewErrorInvalidSearchQuery(err error) *APIError {
	return &APIError{
		Code:     ErrorInvalidSearchQuery.Code,
		Message:  fmt.Sprintf("invalid search query %q", err.Error()),
		HTTPCode: ErrorInvalidSearchQuery.HTTPCode,
	}
}

var (
	// Standard HTTP error 0404
	ErrorRecordNotFound = NewAPIErrorWithHTTP(404, http.StatusNotFound, "record not found")
	// HTTP 500
	ErrorInternalServerError = NewAPIErrorWithHTTP(500, http.StatusInternalServerError, "internal server error")

	// 11** are general errors
	ErrorMalformedRequestData        = NewAPIErrorWithHTTP(1100, http.StatusBadRequest, "could not parse malformed request data")
	ErrorRequestDataValidationFailed = NewAPIErrorWithHTTP(1101, http.StatusBadRequest, "request data validation failed")
	ErrorInvalidSearchToken          = NewAPIErrorWithHTTP(1150, http.StatusBadRequest, "invalid search token")
	ErrorInvalidSearchQuery          = NewAPIErrorWithHTTP(1151, http.StatusBadRequest, "invalid search query")
	ErrorUnknownSearchField          = NewAPIErrorWithHTTP(1152, http.StatusBadRequest, "unknown search field")

	// 13** are authorization errors
	ErrorNoScopeAuthorization = NewAPIErrorWithHTTP(1330, http.StatusForbidden, "no scope authorization")
	ErrorMustBeSelf           = NewAPIErrorWithHTTP(1331, http.StatusForbidden, "must be self")
	ErrorAuthedUserNotFound   = NewAPIErrorWithHTTP(1332, http.StatusForbidden, "authenticated user not found")

	// 14** are user service errors
	ErrorUserMustBeStudent    = NewAPIError(1401, "user must be a student")
	ErrorUserCannotBePrefrosh = NewAPIError(1402, "user cannot be prefrosh")
	ErrorUserNotVisible       = NewAPIError(1403, "user not visible")
	ErrorUserNotAtWilliams    = NewAPIError(1404, "user not at williams")
	ErrorUserIDNoParse        = NewAPIError(1405, "user id could not be parsed")
	ErrorInvalidUserTag       = NewAPIError(1406, "invalid user tag")
	ErrorUnableToSavePicture  = NewAPIErrorWithHTTP(1420, http.StatusInternalServerError, "unable to save uploaded picture")

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

	// 17** are ephcatch errors
	ErrorEphcatchLikeNoSelf    = NewAPIError(1730, "cannot ephcatch-like yourself")
	ErrorEphcatcherNotFound    = NewAPIError(1731, "ephcatcher could not be found")
	ErrorEphcatchAlreadyExists = NewAPIError(1732, "ephcatch already exists with user ID and passed ephcatcher ID")
	ErrorEphcatchDoesNotExist  = NewAPIError(1732, "ephcatch does not exist with user ID and passed ephcatcher ID")

	// 18** are bulletin errors
	ErrorBulletinInvalidDates   = NewAPIError(1830, "start date cannot be after end date")
	ErrorBulletinInvalidType    = NewAPIError(1831, "invalid bulletin type")
	ErrorBulletinRideDateInPast = NewAPIError(1841, "date cannot be in past")
	ErrorDiscussionNotFound     = NewAPIErrorWithHTTP(1850, http.StatusNotFound, "discussion cannot be found")

	// 19** are ephmatch errors
	ErrorEphmatchLikeNoSelf      = NewAPIError(1930, "cannot ephmatch-like yourself")
	ErrorEphmatchProfileNotFound = NewAPIError(1931, "ephmatch profile could not be found")
	ErrorEphmatchAlreadyExists   = NewAPIError(1932, "ephmatch already exists with user ID and passed ephmatch profile user ID")
	ErrorEphmatchDoesNotExist    = NewAPIError(1932, "ephmatch does not exist with user ID and passed ephmatch profile user ID")
)
