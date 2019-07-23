package lib

type APIError struct {
	Code int
	Message string
}

func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code: code,
		Message: message,
	}
}

func (e *APIError) Error() string {
	return e.Message
}

var (
	// 14** are user errors
	ErrorUserMustBeStudent    = NewAPIError(1401, "user must be a student")
	ErrorUserCannotBePrefrosh = NewAPIError(1402, "user cannot be prefrosh")
	ErrorUserNotVisible       = NewAPIError(1403, "user not visible")
	ErrorUserNotAtWilliams    = NewAPIError(1404, "user not at williams")


)
