package seed

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/brianvoe/gofakeit"
)

func GenerateDorm() *models.Dorm {
	keyOrCard := gofakeit.RandString([]string{"key", "card"})
	desc := gofakeit.Paragraph(1, 5, 200, "\n")
	built := gofakeit.Year()
	numSing := gofakeit.Number(0, 99)
	numDoub := gofakeit.Number(0, 49)
	numFlex := gofakeit.Number(0, 14)
	capacity := numSing + 2*numDoub + 2*numFlex
	numBathrooms := capacity / gofakeit.Number(1, capacity)
	numWashers := gofakeit.Number(1, 5)
	bathroomRatio := float64(numBathrooms) / float64(capacity)

	return &models.Dorm{
		Name:            gofakeit.City(),
		KeyOrCard:       &keyOrCard,
		Description:     &desc,
		Built:           &built,
		Capacity:        &capacity,
		NumberBathrooms: &numBathrooms,
		NumberSingles:   &numSing,
		NumberDoubles:   &numDoub,
		NumberFlex:      &numFlex,
		NumberWashers:   &numWashers,
		BathroomRatio:   &bathroomRatio,
	}
}
