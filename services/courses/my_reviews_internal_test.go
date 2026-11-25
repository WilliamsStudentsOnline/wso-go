package courses

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
)

func TestOneCopyPerSection(t *testing.T) {
	assert := testify.New(t)
	sec := func(strm, classNbr int, subject string) *models.CourseOffering {
		return &models.CourseOffering{Strm: strm, ClassNbr: classNbr, Subject: subject, Section: "01", Component: "LEC"}
	}
	got := oneCopyPerSection([]*models.CourseOffering{
		sec(1271, 1263, "MATH"), sec(1271, 1264, "PHYS"), // cross-listed copies
		sec(1253, 3001, "MATH"), sec(1253, 3002, "PHYS"),
		sec(1243, 4001, "MATH"),
	}, map[string]bool{"PHYS": true})
	nbrs := []int{}
	for _, o := range got {
		nbrs = append(nbrs, o.ClassNbr)
	}
	assert.Equal([]int{1264, 3002, 4001}, nbrs) // the PHYS copy wins; order kept

	got = oneCopyPerSection([]*models.CourseOffering{sec(1271, 1263, "MATH"), sec(1271, 1264, "PHYS")}, map[string]bool{"CSCI": true})
	assert.Len(got, 1)
	assert.Equal(1263, got[0].ClassNbr)
}
