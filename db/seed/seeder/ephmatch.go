package seeder

import (
	"math/rand"

	"github.com/WilliamsStudentsOnline/wso-go/db/seed"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func Ephmatches(n int, db *gorm.DB) (err error) {
	profModel := models.NewEphmatchProfileModel(db, zap.S())

	var profiles []*models.EphmatchProfile
	err = profModel.GetAllProfiles(&profiles, nil)
	if err != nil {
		return
	}

	ephmModel := models.NewEphmatchModel(db, zap.S())
	matchModel := models.NewEphmatchMatchesModel(db, zap.S())

	for i := 0; i < n; i++ {
		u1 := profiles[rand.Intn(len(profiles))]
		u2 := profiles[rand.Intn(len(profiles))]
		if u1 == u2 {
			continue
		}
		isMatching, err := matchModel.IsMatching(u1.UserID, u2.UserID)
		if err != nil {
			return err
		}
		if isMatching {
			continue
		}

		_, err = ephmModel.CreateLikeAndMatch(u1.UserID, u2.UserID)
		if err != nil {
			return err
		}
	}

	return
}

func EphmatchProfiles(n int, db *gorm.DB) (err error) {
	userModel := models.NewUserModel(db, zap.S())
	profileModel := models.NewEphmatchProfileModel(db, zap.S())

	var students []models.User
	err = userModel.GetAllUsersByType(&students, models.UserTypeStudent)
	if err != nil {
		return
	}

	rand.Shuffle(len(students), func(i, j int) { students[i], students[j] = students[j], students[i] })

	if n > len(students) {
		n = len(students)
	}

	for i := 0; i < n; i++ {
		u := students[i]
		exists, err := profileModel.DoesProfileExist(u.ID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		p := seed.GenerateEphmatchProfile()
		p.UserID = u.ID

		err = profileModel.CreateProfile(p)
		if err != nil {
			return err
		}
	}

	return
}
