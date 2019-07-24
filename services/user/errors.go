package user

import "github.com/WilliamsStudentsOnline/wso-go/services"

var (
	ErrorNotVisible    = services.NewAPIError(1403, "user not visible")
	ErrorNotAtWilliams = services.NewAPIError(1404, "user not at williams")
)
