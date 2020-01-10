package seed

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func GenerateEphmatch() *models.Ephmatch {
	return &models.Ephmatch{}
}

func GenerateEphmatchWithProfile(profile *models.EphmatchProfile) *models.Ephmatch {
	match := GenerateEphmatch()
	match.UserID = profile.UserID
	match.User = profile.User
	return match
}
