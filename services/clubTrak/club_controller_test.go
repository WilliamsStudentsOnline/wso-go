func TestController_ListSurveys(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	c1 := models.Course{
		Number: "c1",
		AreaOfStudy: &models.AreaOfStudy{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &models.Department{
				Name: "Computer Science",
			},
		},
	}

	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student",
		UnixID: "s1",
	}
	assert.NoError(db.Create(&p1).Create(&s1).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Error)

	// Get test surveys
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/surveys", nil)
	assert.NoError(err)

	// Check if correct surveys (ordered by date)
	resp := GetSurveysFromResp(assert, w)
	assert.NoError(EqualSurveyIDs([]models.FactrakSurvey{fs2, fs1}, resp))

	// Make sure anonymous
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)
}