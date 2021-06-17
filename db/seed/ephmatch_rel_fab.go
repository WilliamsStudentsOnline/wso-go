package seed

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func GenerateEphmatchRelation() *models.EphmatchRelation {
	return &models.EphmatchRelation{}
}

func GenerateEphmatchRelationWithProfile(profile *models.EphmatchProfile) *models.EphmatchRelation {
	match := GenerateEphmatchRelation()
	match.UserID = profile.UserID
	match.User = profile.User
	return match
}
