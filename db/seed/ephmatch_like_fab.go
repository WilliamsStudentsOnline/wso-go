package seed

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func GenerateEphmatchLike() *models.EphmatchLike {
	return &models.EphmatchLike{}
}

func GenerateEphmatchLikeWithProfile(profile *models.EphmatchProfile) *models.EphmatchLike {
	match := GenerateEphmatchLike()
	match.UserID = profile.UserID
	match.User = profile.User
	return match
}
