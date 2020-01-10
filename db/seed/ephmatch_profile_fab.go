package seed

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/brianvoe/gofakeit"
)

func GenerateEphmatchProfile() *models.EphmatchProfile {
	gender := gofakeit.RandString([]string{models.EphmatchProfileGenderMale, models.EphmatchProfileGenderFemale, models.EphmatchProfileGenderNB, "other gender"})
	desc := gofakeit.Sentence(10)

	return &models.EphmatchProfile{
		Gender:      &gender,
		Description: &desc,
	}
}
