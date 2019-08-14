package models_test

import (
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
)

func TestCourseModel_GetAllCourses(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseModel(db)

	courses := []Course{
		{
			Number: "256",
			AreaOfStudy: &AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &Department{
					Name: "Computer Science",
				},
			},
		},
		{
			Number: "120",
			AreaOfStudy: &AreaOfStudy{
				Name:         "Economics",
				Abbreviation: "ECON",
				Department: &Department{
					Name: "Economics",
				},
			},
		},
	}

	for i := range courses {
		assert.NoError(db.Create(&courses[i]).Error)
	}

	var res []*Course
	preloads := []string{"areaOfStudy"}
	assert.NoError(m.GetAllCourses(&res, &GetAllCoursesOptions{Preload: preloads}))

	for i := range courses {
		assert.Equal(courses[i].ID, res[i].ID)
		assert.Equal(courses[i].Number, res[i].Number)
		// Check preload
		assert.Equal(courses[i].AreaOfStudy.ID, res[i].AreaOfStudy.ID)
		assert.Equal(courses[i].AreaOfStudy.Name, res[i].AreaOfStudy.Name)
		assert.Equal(courses[i].AreaOfStudy.Abbreviation, res[i].AreaOfStudy.Abbreviation)
	}
}

func TestCourseModel_GetCourseByID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseModel(db)

	course := Course{
		Number: "256",
		AreaOfStudy: &AreaOfStudy{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &Department{
				Name: "Computer Science",
			},
		},
	}

	assert.NoError(db.Create(&course).Error)

	var res Course
	assert.NoError(m.GetCourseByID(course.ID, &res))

	assert.Equal(course.ID, res.ID)
	assert.Equal(course.Number, res.Number)
	// Does not preload
	assert.Nil(res.AreaOfStudy)
}

func TestCourseModel_FindOrCreate(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseModel(db)

	course := Course{
		Number: "256",
		AreaOfStudy: &AreaOfStudy{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &Department{
				Name: "Computer Science",
			},
		},
	}

	assert.NoError(db.Create(&course).Error)

	// Find
	c1 := Course{
		Number:        "256",
		AreaOfStudyID: course.AreaOfStudyID,
	}
	assert.NoError(m.FindOrCreate(&c1))

	assert.Equal(course.ID, c1.ID)
	assert.Equal(course.Number, c1.Number)

	// Create
	c2 := Course{
		Number:        "236",
		AreaOfStudyID: course.AreaOfStudyID,
	}
	assert.NoError(m.FindOrCreate(&c2))

	assert.Equal(uint(2), c2.ID)
	assert.Equal("236", c2.Number)
}

func TestCourseModel_FindByAbbrevAndNumber(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseModel(db)

	a1 := AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department: &Department{
			Name: "Computer Science",
		},
	}

	// Populate
	c1 := Course{
		Number:      "136",
		AreaOfStudy: &a1,
	}
	c2 := Course{
		Number: "136", // Same number
		AreaOfStudy: &AreaOfStudy{
			Name:         "Economics",
			Abbreviation: "ECON",
			Department: &Department{
				Name: "Economics",
			},
		},
	}
	c3 := Course{
		Number:      "256",
		AreaOfStudy: &a1, // Same dept
	}

	err := db.Create(&c1).Create(&c2).Create(&c3).Error
	assert.NoError(err)

	var resC Course
	err = m.FindByAbbrevAndNumber("CSCI", "136", &resC)
	assert.NoError(err)

	assert.Equal(c1.ID, resC.ID)
	assert.Equal("136", resC.Number)
	assert.Equal("CSCI", resC.AreaOfStudy.Abbreviation)
	assert.Equal("Computer Science", resC.AreaOfStudy.Department.Name)
}

func TestCourseModel_GetCoursesByAreaOfStudyAndProfessors(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseModel(db)

	a1 := AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department: &Department{
			Name: "Computer Science",
		},
	}

	p1 := User{
		UnixID: "p1",
		Name:   "Professor 1",
		Type:   UserTypeProfessor,
	}
	p2 := User{
		UnixID: "p2",
		Name:   "Professor 2",
		Type:   UserTypeProfessor,
	}
	s1 := User{
		UnixID: "s1",
		Name:   "Student 1",
		Type:   UserTypeStudent,
	}

	// Populate
	c1 := Course{
		Number:      "136",
		AreaOfStudy: &a1,
	}
	c2 := Course{
		Number:      "256",
		AreaOfStudy: &a1, // Same dept
	}

	fsC1 := []*FactrakSurvey{
		{
			Course:    &c1,
			User:      &s1,
			Professor: &p1,
		},
		{
			Course:    &c1,
			User:      &s1,
			Professor: &p2,
		},
		{
			Course:    &c2,
			User:      &s1,
			Professor: &p1,
		},
		{
			Course:    &c1,
			User:      &s1,
			Professor: &p1,
		},
	}

	err := db.Create(&a1).Create(&c1).Create(&c2).Create(&p1).Create(&p2).Create(&s1).Error
	assert.NoError(err)

	for _, fs := range fsC1 {
		assert.NoError(db.Create(fs).Error)
	}

	var resC []*Course
	err = m.GetCoursesByAreaOfStudyAndProfessors(a1.ID, &resC)
	assert.NoError(err)

	assert.Len(resC, 2)
	assert.Len(resC[0].Professors, 2)
	assert.Len(resC[1].Professors, 1)

	profIDsC1 := []uint{resC[0].Professors[0].ID, resC[0].Professors[1].ID}

	assert.ElementsMatch([]uint{p1.ID, p2.ID}, profIDsC1)
}
