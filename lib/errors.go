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
	ErrorUnableToSavePicture         = NewAPIErrorWithHTTP(1160, http.StatusInternalServerError, "unable to save uploaded picture")
	ErrorUnableToDeletePicture       = NewAPIErrorWithHTTP(1161, http.StatusInternalServerError, "unable to delete uploaded picture")

	// 13** are authorization errors
	ErrorNoScopeAuthorization         = NewAPIErrorWithHTTP(1330, http.StatusForbidden, "no scope authorization")
	ErrorMustBeSelf                   = NewAPIErrorWithHTTP(1331, http.StatusForbidden, "must be self")
	ErrorAuthedUserNotFound           = NewAPIErrorWithHTTP(1332, http.StatusForbidden, "authenticated user not found")
	ErrorMissingIdentityCredentials   = NewAPIErrorWithHTTP(1351, http.StatusBadRequest, "missing identity credentials (unix id or password)")
	ErrorFailedIdentityAuthentication = NewAPIErrorWithHTTP(1351, http.StatusBadRequest, "failed to authenticate identity (incorrect unix id or password)")

	// 14** are user service errors
	ErrorUserMustBeStudent       = NewAPIError(1401, "user must be a student")
	ErrorUserCannotBePrefrosh    = NewAPIError(1402, "user cannot be prefrosh")
	ErrorUserNotVisible          = NewAPIError(1403, "user not visible")
	ErrorUserNotAtWilliams       = NewAPIError(1404, "user not at williams")
	ErrorUserIDNoParse           = NewAPIError(1405, "user id could not be parsed")
	ErrorInvalidUserTag          = NewAPIError(1406, "invalid user tag")
	ErrorUserInvalidCampusStatus = NewAPIError(1407, "user campus status is invalid")

	// 15** are factrak errors
	// Create/Update errors
	ErrorSurveyMissingCourseParams = NewAPIError(1531, "missing course parameters in create data: courseID or (areaOfStudyAbbreviation and courseNumber)")
	ErrorSurveyCommentTooSmall     = NewAPIError(1532, "comment must be 100 characters or more")
	ErrorSurveyStudentNotFound     = NewAPIError(1533, "user must be a student and could not be found")
	ErrorSurveyProfessorNotFound   = NewAPIError(1534, "passed professor must be a professor and could not be found")
	ErrorSurveyCourseNotFound      = NewAPIError(1535, "passed course could not be found")
	ErrorSurveyAreaOfStudyNotFound = NewAPIError(1536, "passed area of study could not be found")
	ErrorSurveyAlreadyExists       = NewAPIError(1537, "survey already exists with passed user ID, professor ID, and course ID")
	ErrorSurveyCourseSemesterBad   = NewAPIError(1538, "survey course semester has an incorrect year or season")
	ErrorSurveyCourseFormatBad     = NewAPIError(1539, "survey course format is invalid")
	ErrorSurveyCourseYearFuture    = NewAPIError(1540, "survey course year is in the future")
	// Agreement errors
	ErrorSurveyAgreementNotFound      = NewAPIErrorWithHTTP(1551, http.StatusNotFound, "survey agreement could not be found")
	ErrorSurveyAgreementAlreadyExists = NewAPIError(1552, "survey agreement already exists for this user and survey")
	ErrorSurveyAgreementNoSelf        = NewAPIError(1553, "cannot create survey agreement with your own survey")
	// Ranking errors
	ErrorInvalidRankingMetric = NewAPIErrorWithHTTP(1570, http.StatusBadRequest, "cannot rank by this metric")

	// 16** are dormtrak errors
	// Create/Update errors
	ErrorReviewStudentNotFound = NewAPIError(1633, "user must be a student and could not be found")
	ErrorReviewMissingDorm     = NewAPIError(1634, "user is missing dorm field")
	ErrorReviewDormNotOwner    = NewAPIError(1635, "user does not own this dorm room")
	ErrorReviewAlreadyExists   = NewAPIError(1536, "review already exists with passed user ID and dorm room ID")
	ErrorDormtrakTooManyPhotos = NewAPIErrorWithHTTP(1640, http.StatusBadRequest, "too many photos already uploaded to this review")

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
	ErrorEphmatchLikeNoSelf               = NewAPIError(1930, "cannot ephmatch-like yourself")
	ErrorEphmatchProfileNotFound          = NewAPIError(1931, "ephmatch profile could not be found")
	ErrorEphmatchRelationAlreadyExists    = NewAPIError(1932, "ephmatch relation already exists with user ID and passed ephmatch profile user ID")
	ErrorEphmatchDoesNotExist             = NewAPIError(1933, "ephmatch does not exist with user ID and passed ephmatch profile user ID")
	ErrorEphmatchInvalidMessagingPlatform = NewAPIError(1934, "ephmatch messaging platform is invalid")
	ErrorEphmatchEmptyMessagingUsername   = NewAPIError(1935, "ephmatch messaging username is empty")
	ErrorEphmatchDescriptionTooLong       = NewAPIError(1936, "ephmatch description is too long")
	ErrorEphmatchInvalidRelation          = NewAPIError(1937, "ephmatch relation is invalid")
	ErrorEphmatchInvalidLookingFor        = NewAPIError(1938, "ephmatch looking for is invalid")

	// 20** are notification errors
	ErrorNotificationInvalidTokenType = NewAPIError(2030, "notification token type is invalid")
	ErrorNotificationEmptyToken       = NewAPIError(2031, "notification token is empty")

	// 21** are Goodrich errors
	ErrorGoodrichInvalidPaymentMethod = NewAPIError(2130, "goodrich payment method is invalid")
	ErrorGoodrichTimeSlotInvalid      = NewAPIError(2131, "goodrich time slot is not valid")
	ErrorGoodrichMissingWilliamsID    = NewAPIError(2132, "missing williams id number for payment swipe or points")
	ErrorGoodrichUnknownMenuItem      = NewAPIError(2133, "unknown menu item in order")
	ErrorGoodrichUnavailableMenuItem  = NewAPIError(2134, "unavailable menu item in order")
	ErrorGoodrichTimeBadDay           = NewAPIError(2135, "goodrich time is on wrong day")
	ErrorGoodrichTimeFilled           = NewAPIError(2136, "goodrich time is filled")
	ErrorGoodrichComboDealInvalid     = NewAPIError(2137, "cannot get combo pricing with this selection of items")
	ErrorGoodrichSwipeMaxedOut        = NewAPIError(2138, "cannot use a swipe on more than $5")
	ErrorGoodrichOrderNoItems         = NewAPIError(2139, "order has no items")
	ErrorGoodrichDateClosed           = NewAPIError(2140, "goodrich closed on this date")
	ErrorGoodrichOutOfMenuItem        = NewAPIError(2141, "out of a menu item in order")
	ErrorGoodrichInvalidOrderStatus   = NewAPIError(2150, "goodrich order status is invalid")
	ErrorGoodrichNoTimesAvailable     = NewAPIError(2170, "goodrich has no time slots available")
	ErrorGoodrichNoLeasesAvailable    = NewAPIError(2171, "cannot acquire a lease as no leases available")
	ErrorGoodrichLeaseMissing         = NewAPIError(2172, "goodrich lease missing: reload this page and try again")
	ErrorGoodrichLeaseExpired         = NewAPIError(2173, "goodrich order lease has expired")

	// 22** are course scheduler errors
	ErrorCourseSchedulerUnauthorizedUser = NewAPIError(2230, "attempted to access selection of unauthorized user")
	ErrorCourseSchedulerInvalidCourseID  = NewAPIError(2231, "course id not found")
)
