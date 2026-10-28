package seeder

import (
	"fmt"
	"math/rand"

	"github.com/WilliamsStudentsOnline/wso-go/db/seed"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const defaultBoardThreadCount = 12

// BoardThreads inserts synthetic board threads (and a few replies) using existing users.
// When n <= 0, seeds defaultBoardThreadCount threads cycling every board type.
func BoardThreads(n int, db *gorm.DB) error {
	if n <= 0 {
		n = defaultBoardThreadCount
	}

	userModel := models.NewUserModel(db, zap.S())
	var users []models.User
	if err := userModel.GetAllUsersByType(&users, models.UserTypeStudent); err != nil {
		return err
	}
	if len(users) == 0 {
		var all []*models.User
		if err := userModel.GetAllUsers(&all, nil); err != nil {
			return err
		}
		for _, u := range all {
			users = append(users, *u)
		}
	}
	if len(users) == 0 {
		return fmt.Errorf("board seed: no users in database")
	}

	tm := models.NewBoardThreadModel(db, zap.S())
	pm := models.NewBoardPostModel(db, zap.S())

	for i := 0; i < n; i++ {
		u := users[rand.Intn(len(users))]
		s := seed.GenerateBoardThreadCycled(i, u.ID, u.Name)
		if err := tm.CreateThread(s.Thread, s.Body, s.Ride); err != nil {
			return err
		}

		// A few threads get one synthetic reply so reply counts aren't all zero.
		if i%3 == 0 && len(users) > 1 {
			replier := users[rand.Intn(len(users))]
			reply := &models.BoardPost{
				UserID:     replier.ID,
				ExUserName: replier.Name,
				Content:    "Seed reply for local development.",
				IsOP:       lib.FalsePtr(),
			}
			if err := pm.CreateReply(s.Thread.ID, reply); err != nil {
				return err
			}
		}
	}

	return nil
}
