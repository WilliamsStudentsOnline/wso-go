package models_test

import (
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestClubModel_GetAllClubs(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewClubModel(db, zaptest.NewLogger(t).Sugar())

	clubs := []Club{
		{
			Name:     "Kusika",
			Category: CategoryDance,
			//ClubAdminID: 5,
		},
		{
			Name:        "Purple Rain",
			Category:    CategoryArtsEntertainment,
			ClubAdminID: 5,
		},
	}

	for i := range clubs {
		assert.NoError(db.Create(&clubs[i]).Error)
	}

	var res []*Club
	assert.NoError(m.GetAllClubs(&res, &GetAllClubsOptions{}))

	for i := range clubs {
		assert.Equal(clubs[i].ID, res[i].ID)
		assert.Equal(clubs[i].Name, res[i].Name)
	}
}

func TestClubModel_CreateClub(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewClubModel(db, zaptest.NewLogger(t).Sugar())

	clubs := []Club{
		{
			Name:     "Kusika",
			Category: CategoryDance,
		},
		{
			Name:     "Purple Rain",
			Category: CategoryArtsEntertainment,
		},
		{
			Name:        "Spring Streeters",
			Category:    CategoryArtsEntertainment,
			ClubAdminID: 5,
		},
	}
	//Not suppossed to throw errors
	assert.NoError(m.CreateClub(&clubs[2]))

	//Supposed to throw errors-null fields
	assert.NoError(m.CreateClub(&clubs[0]))
	assert.NoError(db.Create(&clubs[1]).Error)

	//
	assert.Equal(uint(5), clubs[2].ClubAdminID)
	assert.Equal(CategoryArtsEntertainment, clubs[2].Category)
}
