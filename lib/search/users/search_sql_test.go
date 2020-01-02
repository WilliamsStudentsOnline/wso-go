package users_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	. "github.com/WilliamsStudentsOnline/wso-go/lib/search/users"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSearchUsersMySQL(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	s := NewSearchUsersMySQL(db)

	u1 := models.User{
		Type:          models.UserTypeStudent,
		Name:          "Aidan Lloyd-Tucker",
		UnixID:        "al15",
		WilliamsEmail: "al15@williams.edu",
		Title:         lib.StrToPtr("Research Assistant"),
		ClassYear:     lib.IntToPtr(2022),
		HomeZip:       lib.StrToPtr("94028"),
		HomeTown:      lib.StrToPtr("Palo Alto"),
		HomeState:     lib.StrToPtr("California"),
		HomeCountry:   lib.StrToPtr("United States"),
		Major:         lib.StrToPtr("Computer Science"),
		SUBox:         lib.StrToPtr("2885"),
		Entry:         lib.StrToPtr("DAMP"),
		DormRoom: &models.DormRoom{
			Dorm: &models.Dorm{
				Neighborhood: &models.Neighborhood{
					Name: "Currier",
				},
				Name: "East",
			},
			Number: "103",
		},
		Pronoun: lib.StrToPtr("he/him/his"),
		Tags: []*models.Tag{
			{
				Name: "WSO",
			},
			{
				Name: "WOC",
			},
		},
	}

	u2 := models.User{
		Type:          models.UserTypeStudent,
		Name:          "Foo Bar",
		UnixID:        "fb1",
		WilliamsEmail: "fb1@williams.edu",
		Title:         lib.StrToPtr("Test Dummy"),
		ClassYear:     lib.IntToPtr(2021),
		HomeZip:       lib.StrToPtr("01267"),
		HomeTown:      lib.StrToPtr("London"),
		HomeState:     lib.StrToPtr("England"),
		HomeCountry:   lib.StrToPtr("United Kingdom"),
		Major:         lib.StrToPtr("Economics"),
		SUBox:         lib.StrToPtr("1234"),
		Entry:         lib.StrToPtr("Sage AB"),
		DormRoom: &models.DormRoom{
			Dorm: &models.Dorm{
				Neighborhood: &models.Neighborhood{
					Name: "Dodd",
				},
				Name: "Hubble",
			},
			Number: "304",
		},
		Pronoun: lib.StrToPtr("they/them"),
		Tags: []*models.Tag{
			{
				Name: "ABCD",
			},
		},
	}

	u3 := models.User{
		Type:           models.UserTypeProfessor,
		Name:           "Bobby Tables",
		CampusPhoneExt: lib.StrToPtr("413 987 6543"),
		UnixID:         "bot3",
		WilliamsEmail:  "bot3@williams.edu",
		Title:          lib.StrToPtr("Professor of Computer Science"),
		Department: &models.Department{
			Name:  "Computer Science",
			Users: nil,
			AreasOfStudy: []*models.AreaOfStudy{
				{
					Name:         "Computer Science",
					Abbreviation: "CSCI",
				},
			},
		},
		Office: &models.Office{
			Number: "TCL 304",
		},
	}

	assert.NoError(db.Create(&u1).Create(&u2).Create(&u3).Error)

	userModel := models.NewUserModel(db, zap.S())

	// Populate search fields
	for _, u := range []models.User{u1, u2, u3} {
		assert.NoError(userModel.PopulateSearchFields(u.ID))
	}

	testCases := []struct {
		name     string
		query    string
		expected []*models.User
	}{
		{
			"by search_field name",
			"Aidan",
			[]*models.User{&u1},
		},
		{
			"by search_field unknown",
			"helloworld",
			[]*models.User{},
		},
		{
			"by search_field full name",
			"Aidan Lloyd-Tucker",
			[]*models.User{&u1},
		},
		{
			"by search_field unix",
			"al15",
			[]*models.User{&u1},
		},
		{
			"by field unix",
			"unix: al15",
			[]*models.User{&u1},
		},
		{
			"by field unixes",
			"unix: al15 OR unix: fb1",
			[]*models.User{&u1, &u2},
		},
		{
			"by field hometown",
			"town: Palo Alto",
			[]*models.User{&u1},
		},
		{
			"by search_fields hometown",
			"Palo Alto",
			[]*models.User{&u1},
		},
		{
			"by field building",
			"building: TCL",
			[]*models.User{&u3},
		},
		{
			"by field dorm",
			"dorm: East",
			[]*models.User{&u1},
		},
		{
			"by field dorms",
			"dorm: East OR dorm: Hubble",
			[]*models.User{&u1, &u2},
		},
		{
			"by field room",
			"room: 103",
			[]*models.User{&u1},
		},
		{
			"by field room, dorm",
			"room: 103 dorm: East",
			[]*models.User{&u1},
		},
		{
			"by field room, dorm bad",
			"room: 103 dorm: Hubble",
			[]*models.User{},
		},
		{
			"by field room or dorm",
			"room: 103 OR dorm: Hubble",
			[]*models.User{&u1, &u2},
		},
		{
			"by field department",
			"department: Computer science",
			[]*models.User{&u3},
		},
		{
			"by field department abbreviation",
			"department: csCi",
			[]*models.User{&u3},
		},
		{
			"by field name partial",
			"name: ida",
			[]*models.User{&u1},
		},
		{
			"by field neighborhood",
			"neighborhood: currier",
			[]*models.User{&u1},
		},
		{
			"by field neighborhoods",
			"neighborhood: currier OR neighborhood: dodd",
			[]*models.User{&u1, &u2},
		},
		{
			"by field tag",
			"tags: woc",
			[]*models.User{&u1},
		},
		{
			"by field tags",
			"tags: wso OR tags: abcd",
			[]*models.User{&u1, &u2},
		},
		{
			"complex query",
			"(neighborhood: currier) OR (class_year: 2021 AND tags: abcd) OR (department: CSCI AND (building: TCL OR building: HELLO) OR unix: al15)",
			[]*models.User{&u1, &u2, &u3},
		},
		{
			"complex query with search_fields",
			"aidan ((neighborhood: dodd) OR (class_year: 2021 AND tags: abcd) OR (department: CSCI AND (building: TCL OR building: HELLO)))",
			[]*models.User{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, _, err := s.Search(tc.query, nil)
			testify.NoError(t, err)

			testify.Len(t, res, len(tc.expected))
			for i := range tc.expected {
				testify.Equal(t, tc.expected[i].ID, res[i].ID)
			}
		})
	}

	errorCases := []struct {
		name     string
		query    string
		expected *lib.APIError
	}{
		{
			"bad field",
			"aidan foobar:hi",
			lib.NewErrorUnknownSearchField("foobar"),
		},
		{
			"bad field 2",
			"Location:md3",
			lib.NewErrorUnknownSearchField("Location"),
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := s.Search(tc.query, nil)
			testify.Error(t, err)
			testify.IsType(t, &lib.APIError{}, err)
			testify.Equal(t, tc.expected, err.(*lib.APIError))
		})
	}

}
