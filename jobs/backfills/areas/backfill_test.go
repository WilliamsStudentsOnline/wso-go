package areas_test

import (
	"testing"

	areasbackfill "github.com/WilliamsStudentsOnline/wso-go/jobs/backfills/areas"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAreasBackfill_DryRunApplyIdempotent(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()

	depts := map[string]*models.Department{}
	for _, name := range []string{"Asian American Studies", "Williams-Mystic Program"} {
		d := &models.Department{Name: name}
		assert.NoError(db.Create(d).Error)
		depts[name] = d
	}

	res, err := areasbackfill.Run(db, log, areasbackfill.Options{})
	assert.NoError(err)
	assert.Equal([]string{"AAS", "CAOS"}, res.Created)
	assert.Equal([]string{"EALC"}, res.MissingDepartment)
	var count int
	assert.NoError(db.Model(&models.AreaOfStudy{}).Count(&count).Error)
	assert.Equal(0, count)

	res, err = areasbackfill.Run(db, log, areasbackfill.Options{Apply: true})
	assert.NoError(err)
	assert.Equal([]string{"AAS", "CAOS"}, res.Created)
	assert.Empty(res.Existing)
	assert.Equal([]string{"EALC"}, res.MissingDepartment)

	var caos models.AreaOfStudy
	assert.NoError(db.Where("abbrev = ?", "CAOS").First(&caos).Error)
	assert.Equal("Coastal and Ocean Studies", caos.Name)
	assert.Equal("concentration", caos.CatalogKind)
	assert.Equal("https://catalog.williams.edu/caos/", caos.CatalogURL)
	assert.Equal("https://catalog.williams.edu/pdf/caos.pdf", caos.CurriculumPDFURL)
	if assert.NotNil(caos.DepartmentID) {
		assert.Equal(depts["Williams-Mystic Program"].ID, *caos.DepartmentID)
	}

	res, err = areasbackfill.Run(db, log, areasbackfill.Options{Apply: true})
	assert.NoError(err)
	assert.Empty(res.Created)
	assert.Equal([]string{"AAS", "CAOS"}, res.Existing)
	assert.NoError(db.Model(&models.AreaOfStudy{}).Count(&count).Error)
	assert.Equal(2, count)

	assert.NoError(db.Create(&models.Department{Name: "Asian Languages,Lit&Culture"}).Error)
	res, err = areasbackfill.Run(db, log, areasbackfill.Options{Apply: true})
	assert.NoError(err)
	assert.Equal([]string{"EALC"}, res.Created)
	assert.NoError(db.Model(&models.AreaOfStudy{}).Count(&count).Error)
	assert.Equal(3, count)
}
