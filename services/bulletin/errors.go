package bulletin

import "github.com/WilliamsStudentsOnline/wso-go/services"

// TODO: merge with errors in library branch
var (
	// 16** are Bulletin errors
	ErrorBulletinMissingParams = services.NewAPIError(1601, "missing bulletin parameters: title or body is empty")
	ErrorBulletinMissingDates  = services.NewAPIError(1602, "missing bulletin parameters: start or end dates")
	ErrorBulletinInvalidDates  = services.NewAPIError(1603, "start date cannot be after end date")
	ErrorBulletinUserNotFound  = services.NewAPIError(1604, "user not found")
	ErrorBulletinInvalidType   = services.NewAPIError(1605, "invalid  bulletin type")
)
