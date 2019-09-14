package autocomplete_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/lib/autocomplete"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/autocomplete"
	testify "github.com/stretchr/testify/assert"
)

func TestController_AreaOfStudy(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	d1 := models.Department{
		Name: "Computer Science",
	}
	assert.NoError(db.Create(&d1).Error)

	// Insert test objects into db
	areas := []*models.AreaOfStudy{
		{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
		},
		{
			Name:         "Comparative Literature",
			Abbreviation: "COMP",
		},
		{
			Name:         "Economics",
			Abbreviation: "ECON",
		},
		{
			Name:         "Physics",
			Abbreviation: "PHYS",
		},
		{
			Name:         "Philosophy",
			Abbreviation: "PHIL",
		},
	}
	for i := range areas {
		areas[i].Department = &d1
		assert.NoError(db.Create(areas[i]).Error)
	}

	testCases := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			"empty",
			"",
			[]string{},
		},
		{
			"abbrev",
			"csc",
			[]string{"CSCI"},
		},
		{
			"name word 1",
			"Compute",
			[]string{"Computer Science"},
		},
		{
			"name word 2",
			"Computer Sci",
			[]string{"Computer Science"},
		},
		{
			"mixed name and mixed abbrev",
			"Comp",
			[]string{"COMP", "Computer Science", "Comparative Literature"},
		},
		{
			"mixed abbrev",
			"ph",
			[]string{"PHYS", "PHIL", "Physics", "Philosophy"},
		},
		{
			"mixed abbrev econ",
			"eco",
			[]string{"ECON", "Economics"},
		},
		{
			"full",
			"csci",
			[]string{"CSCI"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)

			w, err := utils.DoHTTPReq(router, http.MethodGet, "/area-of-study?q="+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			resp := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(resp.Error)
			var respEntries []autocomplete.ACEntry
			err = json.Unmarshal(resp.Data, &respEntries)
			a.NoError(err)

			// Check if correct response
			respStrs := getRespStrs(respEntries)
			a.Len(respStrs, len(tc.expected))
			a.Equal(tc.expected, respStrs)
		})
	}
}

func TestController_Course(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	a1 := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department: &models.Department{
			Name: "Computer Science",
		},
	}
	assert.NoError(db.Create(&a1).Error)

	// Insert test objects into db
	courses := []*models.Course{
		{
			Number:      "134",
			AreaOfStudy: &a1,
		},
		{
			Number:      "136",
			AreaOfStudy: &a1,
		},
		{
			Number:      "237",
			AreaOfStudy: &a1,
		},
		{
			Number:      "256",
			AreaOfStudy: &a1,
		},
		{
			Number: "120",
			AreaOfStudy: &models.AreaOfStudy{
				Name:         "Economics",
				Abbreviation: "ECON",
				Department: &models.Department{
					Name: "Economics",
				},
			},
		},
		{
			Number: "101",
			AreaOfStudy: &models.AreaOfStudy{
				Name:         "Environmental Studies",
				Abbreviation: "ENVI",
				Department: &models.Department{
					Name: "Environmental Studies",
				},
			},
		},
	}
	for i := range courses {
		assert.NoError(db.Create(courses[i]).Error)
	}

	testCases := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			"empty",
			"",
			[]string{},
		},
		{
			"abbrev",
			"e",
			[]string{"ECON", "ENVI"},
		},
		{
			"abbrev got one but not finished",
			"csc",
			[]string{"CSCI"},
		},
		{
			"abbrev finished",
			"csci",
			[]string{"CSCI 134", "CSCI 136", "CSCI 237", "CSCI 256"},
		},
		{
			"number 2 digit",
			"csci 2",
			[]string{"CSCI 237", "CSCI 256"},
		},
		{
			"number 2 digit",
			"csci 13",
			[]string{"CSCI 134", "CSCI 136"},
		},
		{
			"full",
			"csci 256",
			[]string{"CSCI 256"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)

			w, err := utils.DoHTTPReq(router, http.MethodGet, "/course?q="+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			resp := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(resp.Error)
			var respEntries []autocomplete.ACEntry
			err = json.Unmarshal(resp.Data, &respEntries)
			a.NoError(err)

			// Check if correct response
			respStrs := getRespStrs(respEntries)
			a.Len(respStrs, len(tc.expected))
			a.Equal(tc.expected, respStrs)
		})
	}
}

func TestController_Professor(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	// Insert test objects into db
	profs := []*models.User{
		{
			Name: "Aidan Lloyd-Tucker",
		},
		{
			Name: "Bob Smith",
		},
		{
			Name: "Maud Mandel",
		},
		{
			Name: "Adam Falk",
		},
	}
	for i := range profs {
		profs[i].Type = models.UserTypeProfessor
		profs[i].UnixID = "p" + strconv.Itoa(i)
		assert.NoError(db.Create(profs[i]).Error)
	}

	testCases := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			"empty",
			"",
			[]string{},
		},
		{
			"fn: a",
			"a",
			[]string{"Adam Falk", "Maud Mandel", "Aidan Lloyd-Tucker"},
		},
		{
			"fn: ad",
			"ad",
			[]string{"Adam Falk"},
		},
		{
			"last name",
			"man",
			[]string{"Maud Mandel"},
		},
		{
			"multiple names",
			"mandel falk",
			[]string{"Adam Falk", "Maud Mandel"},
		},
		{
			"full",
			"adam falk",
			[]string{"Adam Falk"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)

			w, err := utils.DoHTTPReq(router, http.MethodGet, "/professor?q="+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			resp := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(resp.Error)
			var respEntries []autocomplete.ACEntry
			err = json.Unmarshal(resp.Data, &respEntries)
			a.NoError(err)

			// Check if correct response
			respStrs := getRespStrs(respEntries)
			a.Len(respStrs, len(tc.expected))
			a.Equal(tc.expected, respStrs)
		})
	}
}

func TestController_Tag(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	// Insert test objects into db
	tags := []*models.Tag{
		{
			Name: "WSO",
		},
		{
			Name: "WOC",
		},
		{
			Name: "Cycling",
		},
		{
			Name: "Computer Science",
		},
	}
	for i := range tags {
		assert.NoError(db.Create(tags[i]).Error)
	}

	testCases := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			"empty",
			"",
			[]string{},
		},
		{
			"tag",
			"w",
			[]string{"WSO", "WOC"},
		},
		{
			"tag middle",
			"o",
			[]string{"WSO", "WOC", "Computer Science"},
		},
		{
			"tag single",
			"cycl",
			[]string{"Cycling"},
		},
		{
			"full",
			"wso",
			[]string{"WSO"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)

			w, err := utils.DoHTTPReq(router, http.MethodGet, "/tag?q="+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			resp := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(resp.Error)
			var respEntries []autocomplete.ACEntry
			err = json.Unmarshal(resp.Data, &respEntries)
			a.NoError(err)

			// Check if correct response
			respStrs := getRespStrs(respEntries)
			a.Len(respStrs, len(tc.expected))
			a.Equal(tc.expected, respStrs)
		})
	}
}

func getRespStrs(entries []autocomplete.ACEntry) []string {
	strs := make([]string, len(entries))
	for i := range entries {
		strs[i] = entries[i].Value
	}
	return strs
}
