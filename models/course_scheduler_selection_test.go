package models_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestCourseSchedulerSelectionModel_GetAllCourseSchedulerSelections(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseSchedulerSelectionModel(db, zaptest.NewLogger(t).Sugar())

	var userID1 uint = 128
	var userID2 uint = 400
	var courseID1 uint = 256
	var courseID2 uint = 101
	var year uint = 2024
	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Foo",
				UnixID:      "foo1",
				Title:       lib.StrToPtr("Research Assistant"),
				ClassYear:   lib.IntToPtr(2022),
				Major:       lib.StrToPtr("Computer Science"),
				SUBox:       lib.StrToPtr("2885"),
				Entry:       lib.StrToPtr("AP3"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Currier",
						},
						Name: "East",
					},
					Number: "103",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("Palo Alto"),
				HomeState:   lib.StrToPtr("California"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID1,
			Course: &Course{
				Number: "256",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Computer Science",
					Abbreviation: "CSCI",
					Department: &Department{
						Name: "Computer Science",
					},
				},
			},
			CourseID: &courseID1,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Bar",
				UnixID:      "bar1",
				Title:       lib.StrToPtr("Research Assistant"),
				ClassYear:   lib.IntToPtr(2021),
				Major:       lib.StrToPtr("Psychology"),
				SUBox:       lib.StrToPtr("1024"),
				Entry:       lib.StrToPtr("MD3"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Greylock",
						},
						Name: "Greylock",
					},
					Number: "104",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("New York"),
				HomeState:   lib.StrToPtr("New York"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID2,
			Course: &Course{
				Number: "101",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Psychology",
					Abbreviation: "101",
					Department: &Department{
						Name: "Psychology",
					},
				},
			},
			CourseID: &courseID2,
			Hidden:   false,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	for i := range courseSchedulerSelections {
		assert.NoError(db.Create(&courseSchedulerSelections[i]).Error)
	}

	var res []*CourseSchedulerSelection
	assert.NoError(m.GetAllCourseSchedulerSelections(&res, &GetAllCourseSchedulerSelectionsOptions{Preload: []string{"user", "course"}}))

	for i := range courseSchedulerSelections {
		assert.Equal(courseSchedulerSelections[i].UserID, res[i].UserID)
		assert.Equal(courseSchedulerSelections[i].CourseID, res[i].CourseID)
		assert.Equal(courseSchedulerSelections[i].Hidden, res[i].Hidden)
		assert.Equal(courseSchedulerSelections[i].Semester, res[i].Semester)
		assert.Equal(courseSchedulerSelections[i].Year, res[i].Year)

		// Check preloading
		assert.Equal(courseSchedulerSelections[i].User.UnixID, res[i].User.UnixID)
		assert.Equal(courseSchedulerSelections[i].Course.Number, res[i].Course.Number)
	}

}

func TestCourseSchedulerSelectionModel_AddSelection(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseSchedulerSelectionModel(db, zaptest.NewLogger(t).Sugar())

	user := &User{
		Type:        UserTypeStudent,
		Name:        "Foo",
		UnixID:      "foo1",
		Title:       lib.StrToPtr("Research Assistant"),
		ClassYear:   lib.IntToPtr(2022),
		Major:       lib.StrToPtr("Computer Science"),
		SUBox:       lib.StrToPtr("2885"),
		Entry:       lib.StrToPtr("AP3"),
		DormVisible: lib.BoolToPtr(true),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{
					Name: "Currier",
				},
				Name: "East",
			},
			Number: "103",
		},
		HomeVisible: lib.BoolToPtr(true),
		HomeTown:    lib.StrToPtr("Palo Alto"),
		HomeState:   lib.StrToPtr("California"),
		HomeCountry: lib.StrToPtr("United States"),
	}
	course := &Course{
		Number: "256",
		AreaOfStudy: &AreaOfStudy{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &Department{
				Name: "Computer Science",
			},
		},
	}

	assert.NoError(m.AddSelection(user, course, true, SemesterFall, 2023))

	// depends on GetAllCourseSchedulerSelections working, see test above
	var res []*CourseSchedulerSelection
	assert.NoError(m.GetAllCourseSchedulerSelections(&res, &GetAllCourseSchedulerSelectionsOptions{Preload: []string{"user", "course"}}))

	for i := range res {
		assert.Equal(true, res[i].Hidden)
		assert.Equal("SPRING", res[i].Semester)
		assert.Equal(2023, res[i].Year)

		// Check preloaded fields
		assert.Equal("Foo", res[i].User.Name)
		assert.Equal("256", res[i].Course.Number)

		// Check auto-generated IDs
		assert.Equal(0, res[i].User.ID)
		assert.Equal(0, res[i].Course.ID)
	}

}

func TestCourseSchedulerSelectionModel_DeleteSelectionsByUserIDAndCourseID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseSchedulerSelectionModel(db, zaptest.NewLogger(t).Sugar())

	var userID1 uint = 128
	var userID2 uint = 400
	var courseID1 uint = 256
	var courseID2 uint = 101
	var year uint = 2024
	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Foo",
				UnixID:      "foo1",
				Title:       lib.StrToPtr("Research Assistant"),
				ClassYear:   lib.IntToPtr(2022),
				Major:       lib.StrToPtr("Computer Science"),
				SUBox:       lib.StrToPtr("2885"),
				Entry:       lib.StrToPtr("AP3"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Currier",
						},
						Name: "East",
					},
					Number: "103",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("Palo Alto"),
				HomeState:   lib.StrToPtr("California"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID1,
			Course: &Course{
				Number: "256",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Computer Science",
					Abbreviation: "CSCI",
					Department: &Department{
						Name: "Computer Science",
					},
				},
			},
			CourseID: &courseID1,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Bar",
				UnixID:      "bar1",
				Title:       lib.StrToPtr("Research Assistant"),
				ClassYear:   lib.IntToPtr(2021),
				Major:       lib.StrToPtr("Psychology"),
				SUBox:       lib.StrToPtr("1024"),
				Entry:       lib.StrToPtr("MD3"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Greylock",
						},
						Name: "Greylock",
					},
					Number: "104",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("New York"),
				HomeState:   lib.StrToPtr("New York"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID2,
			Course: &Course{
				Number: "101",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Psychology",
					Abbreviation: "101",
					Department: &Department{
						Name: "Psychology",
					},
				},
			},
			CourseID: &courseID2,
			Hidden:   false,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	for i := range courseSchedulerSelections {
		assert.NoError(db.Create(&courseSchedulerSelections[i]).Error)
	}

	assert.NoError(m.DeleteSelectionsByUserIDAndCourseID(userID1, courseID1))

	// depends on GetAllCourseSchedulerSelections working, see test above
	var res []*CourseSchedulerSelection
	assert.NoError(m.GetAllCourseSchedulerSelections(&res, &GetAllCourseSchedulerSelectionsOptions{}))

	assert.Equal(1, len(res))
	assert.Equal(courseID2, *res[0].CourseID)
	assert.Equal(false, res[0].Hidden)

}

func TestCourseSchedulerSelectionModel_GetSelectionsByUserID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseSchedulerSelectionModel(db, zaptest.NewLogger(t).Sugar())

	var userID1 uint = 128
	var userID2 uint = 400
	var courseID1 uint = 256
	var courseID2 uint = 101
	var courseID3 uint = 151
	var year uint = 2024
	User1 := &User{
		Type:        UserTypeStudent,
		Name:        "Foo",
		UnixID:      "foo1",
		Title:       lib.StrToPtr("Research Assistant"),
		ClassYear:   lib.IntToPtr(2022),
		Major:       lib.StrToPtr("Computer Science"),
		SUBox:       lib.StrToPtr("2885"),
		Entry:       lib.StrToPtr("AP3"),
		DormVisible: lib.BoolToPtr(true),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{
					Name: "Currier",
				},
				Name: "East",
			},
			Number: "103",
		},
		HomeVisible: lib.BoolToPtr(true),
		HomeTown:    lib.StrToPtr("Palo Alto"),
		HomeState:   lib.StrToPtr("California"),
		HomeCountry: lib.StrToPtr("United States"),
	}
	User2 := &User{
		Type:        UserTypeStudent,
		Name:        "Bar",
		UnixID:      "bar1",
		Title:       lib.StrToPtr("Research Assistant"),
		ClassYear:   lib.IntToPtr(2021),
		Major:       lib.StrToPtr("Psychology"),
		SUBox:       lib.StrToPtr("1024"),
		Entry:       lib.StrToPtr("MD3"),
		DormVisible: lib.BoolToPtr(true),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{
					Name: "Greylock",
				},
				Name: "Greylock",
			},
			Number: "104",
		},
		HomeVisible: lib.BoolToPtr(true),
		HomeTown:    lib.StrToPtr("New York"),
		HomeState:   lib.StrToPtr("New York"),
		HomeCountry: lib.StrToPtr("United States"),
	}
	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User:   User1,
			UserID: &userID1,
			Course: &Course{
				Number: "256",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Computer Science",
					Abbreviation: "CSCI",
					Department: &Department{
						Name: "Computer Science",
					},
				},
			},
			CourseID: &courseID1,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User:   User2,
			UserID: &userID2,
			Course: &Course{
				Number: "101",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Psychology",
					Abbreviation: "101",
					Department: &Department{
						Name: "Psychology",
					},
				},
			},
			CourseID: &courseID2,
			Hidden:   false,
			Semester: SemesterSpring,
			Year:     &year,
		},
		{
			User:   User2,
			UserID: &userID2,
			Course: &Course{
				Number: "151",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Calculus",
					Abbreviation: "141",
					Department: &Department{
						Name: "Mathematics",
					},
				},
			},
			CourseID: &courseID3,
			Hidden:   true,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	for i := range courseSchedulerSelections {
		assert.NoError(db.Create(&courseSchedulerSelections[i]).Error)
	}

	var res []*CourseSchedulerSelection
	assert.NoError(m.GetSelectionsByUserID(userID2, &res))

	assert.Equal(2, len(res))
	assert.Equal(res[0].UserID, res[1].UserID)
	assert.Equal(res[0].UserID, &userID2)

}

func TestCourseSchedulerSelectionModel_SetSelectionsHiddenByUserIDAndCourseID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCourseSchedulerSelectionModel(db, zaptest.NewLogger(t).Sugar())

	var userID1 uint = 128
	var userID2 uint = 400
	var courseID1 uint = 256
	var courseID2 uint = 101
	var courseID3 uint = 151
	var year uint = 2024
	User1 := &User{
		Type:        UserTypeStudent,
		Name:        "Foo",
		UnixID:      "foo1",
		Title:       lib.StrToPtr("Research Assistant"),
		ClassYear:   lib.IntToPtr(2022),
		Major:       lib.StrToPtr("Computer Science"),
		SUBox:       lib.StrToPtr("2885"),
		Entry:       lib.StrToPtr("AP3"),
		DormVisible: lib.BoolToPtr(true),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{
					Name: "Currier",
				},
				Name: "East",
			},
			Number: "103",
		},
		HomeVisible: lib.BoolToPtr(true),
		HomeTown:    lib.StrToPtr("Palo Alto"),
		HomeState:   lib.StrToPtr("California"),
		HomeCountry: lib.StrToPtr("United States"),
	}
	User2 := &User{
		Type:        UserTypeStudent,
		Name:        "Bar",
		UnixID:      "bar1",
		Title:       lib.StrToPtr("Research Assistant"),
		ClassYear:   lib.IntToPtr(2021),
		Major:       lib.StrToPtr("Psychology"),
		SUBox:       lib.StrToPtr("1024"),
		Entry:       lib.StrToPtr("MD3"),
		DormVisible: lib.BoolToPtr(true),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{
					Name: "Greylock",
				},
				Name: "Greylock",
			},
			Number: "104",
		},
		HomeVisible: lib.BoolToPtr(true),
		HomeTown:    lib.StrToPtr("New York"),
		HomeState:   lib.StrToPtr("New York"),
		HomeCountry: lib.StrToPtr("United States"),
	}
	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User:   User1,
			UserID: &userID1,
			Course: &Course{
				Number: "256",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Computer Science",
					Abbreviation: "CSCI",
					Department: &Department{
						Name: "Computer Science",
					},
				},
			},
			CourseID: &courseID1,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User:   User2,
			UserID: &userID2,
			Course: &Course{
				Number: "101",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Psychology",
					Abbreviation: "101",
					Department: &Department{
						Name: "Psychology",
					},
				},
			},
			CourseID: &courseID2,
			Hidden:   false,
			Semester: SemesterSpring,
			Year:     &year,
		},
		{
			User:   User2,
			UserID: &userID2,
			Course: &Course{
				Number: "151",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Calculus",
					Abbreviation: "151",
					Department: &Department{
						Name: "Mathematics",
					},
				},
			},
			CourseID: &courseID3,
			Hidden:   true,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	for i := range courseSchedulerSelections {
		assert.NoError(db.Create(&courseSchedulerSelections[i]).Error)
	}

	assert.NoError(m.SetSelectionHiddenByUserIDAndCourseID(userID2, courseID3, false))

	// depends on GetAllCourseSchedulerSelections working, see test above
	var res []*CourseSchedulerSelection
	assert.NoError(m.GetAllCourseSchedulerSelections(&res, &GetAllCourseSchedulerSelectionsOptions{}))

	assert.Equal(3, len(res))
	assert.Equal(true, res[0].Hidden)
	assert.Equal(false, res[1].Hidden)
	assert.Equal(false, res[2].Hidden)

}
