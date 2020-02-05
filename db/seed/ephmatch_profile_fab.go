package seed

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/brianvoe/gofakeit"
)

func GenerateEphmatchProfile() *models.EphmatchProfile {
	desc := gofakeit.Sentence(10)
	matchMessage := gofakeit.Sentence(4)

	return &models.EphmatchProfile{
		Description:  &desc,
		MatchMessage: &matchMessage,
	}
}
