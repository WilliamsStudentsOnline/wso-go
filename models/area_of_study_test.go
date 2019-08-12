package models_test

import (
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
)

func TestAreaOfStudyModel_GetAllAreasOfStudy(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewAreaOfStudyModel(db)

	areasOfStudy := []AreaOfStudy{
		{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &Department{
				Name: "Computer Science",
			},
		},
		{
			Name:         "Economics",
			Abbreviation: "ECON",
			Department: &Department{
				Name: "Economics",
			},
		},
	}

	for i := range areasOfStudy {
		assert.NoError(db.Create(&areasOfStudy[i]).Error)
	}

	var res []AreaOfStudy
	assert.NoError(m.GetAllAreasOfStudy(&res))

	for i := range areasOfStudy {
		assert.Equal(areasOfStudy[i].ID, res[i].ID)
		assert.Equal(areasOfStudy[i].Name, res[i].Name)
		assert.Equal(areasOfStudy[i].Abbreviation, res[i].Abbreviation)
	}
}

func TestAreaOfStudyModel_GetAreaOfStudyByID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewAreaOfStudyModel(db)

	areaOfStudy := AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department: &Department{
			Name: "Computer Science",
		},
	}

	assert.NoError(db.Create(&areaOfStudy).Error)

	var res AreaOfStudy
	assert.NoError(m.GetAreaOfStudyByID(areaOfStudy.ID, &res))

	assert.Equal(areaOfStudy.ID, res.ID)
	assert.Equal(areaOfStudy.Name, res.Name)
	assert.Equal(areaOfStudy.Abbreviation, res.Abbreviation)
	assert.Equal(areaOfStudy.Department.ID, res.Department.ID)
	assert.Equal(areaOfStudy.Department.Name, res.Department.Name)
}

func TestAreaOfStudyModel_DoesAreaOfStudyExist(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewAreaOfStudyModel(db)

	areaOfStudy := AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department: &Department{
			Name: "Computer Science",
		},
	}

	assert.NoError(db.Create(&areaOfStudy).Error)

	exists, err := m.DoesAreaOfStudyExist(areaOfStudy.ID)
	assert.NoError(err)
	assert.True(exists)

	exists, err = m.DoesAreaOfStudyExist(48)
	assert.NoError(err)
	assert.False(exists)
}

func TestAreaOfStudyModel_GetAreaOfStudyByAbbreviation(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewAreaOfStudyModel(db)

	areasOfStudy := []AreaOfStudy{
		{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &Department{
				Name: "Computer Science",
			},
		},
		{
			Name:         "Economics",
			Abbreviation: "ECON",
			Department: &Department{
				Name: "Economics",
			},
		},
	}

	for i := range areasOfStudy {
		assert.NoError(db.Create(&areasOfStudy[i]).Error)
	}

	// Test uppercase
	var res AreaOfStudy
	assert.NoError(m.GetAreaOfStudyByAbbreviation("CSCI", &res))

	assert.Equal(areasOfStudy[0].ID, res.ID)
	assert.Equal(areasOfStudy[0].Name, res.Name)
	assert.Equal(areasOfStudy[0].Abbreviation, res.Abbreviation)

	// Test lowercase
	res = AreaOfStudy{}
	assert.NoError(m.GetAreaOfStudyByAbbreviation("econ", &res))

	assert.Equal(areasOfStudy[1].ID, res.ID)
	assert.Equal(areasOfStudy[1].Name, res.Name)
	assert.Equal(areasOfStudy[1].Abbreviation, res.Abbreviation)
}
