package schedule_notifs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib/notification"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type notificationInfo struct {
	Item   string
	Course string
	Meal   string
	Vendor string
}

func FoodNotify(cfg *config.Config, db *gorm.DB, log *zap.SugaredLogger) error {

	//get list of all keywords
	var keywords []*models.DiningKeyword
	m := models.NewDiningKeywordModel(db, log)
	m.GetAllKeywords(&keywords)

	// Just start up APNS for now, no Android

	notifClient, err := notification.NewClient(cfg)
	if err != nil {
		return err
	}

	for _, keyword := range keywords {
		keywordNtf, err := GenerateNotification(keyword.Keyword, cfg.DiningFile)

		if err != nil {
			return err
		}

		// checks to make sure whether the keyword was actually found in anything.
		if keywordNtf == nil {
			continue
		}
		var keywordUsers []*models.User
		m.GetUsersForKeyword(keyword, &keywordUsers)
		var keywordTokens []*models.NotificationToken
		notifTokenModel := models.NewNotificationTokenModel(db, log)

		// finds every notification token that is associated with the keywords' users
		for _, user := range keywordUsers {
			var tempTokens []*models.NotificationToken
			notifTokenModel.GetNotifTokensForUser(user, &tempTokens)
			keywordTokens = append(keywordTokens, tempTokens...)
		}

		for _, token := range keywordTokens {
			if token.Type != models.NotificationTokenTypeIOS {
				continue
			}

			delToken, notifErr := notifClient.NotifyIOS(keywordNtf, token.Token)
			if delToken {
				err = notifTokenModel.DeleteToken(token)
				if err != nil {
					log.Error("unhandled notification error", notifErr)
					return err
				}
			}
			if notifErr != nil {
				return notifErr
			}
		}
	}

	return nil
}

// will be called for each keyword in the database
func GenerateNotification(keyword string, diningFile string) (*notification.Notification, error) {
	diningData, err := ioutil.ReadFile(diningFile)
	if err != nil {
		return nil, err
	}

	dining := dining_update.ExportDining{}
	err = json.Unmarshal(diningData, &dining)
	if err != nil {
		return nil, err
	}

	diningFileTime, err := time.Parse(time.RFC850, dining.UpdateTime)
	if err != nil {
		return nil, err
	}

	// If updated file is not the same day as today,
	if time.Now().After(diningFileTime.Add(time.Hour * 24)) {
		return nil, errors.New("dining file is out of date")
	}

	// List of meals that have this keyword in its name
	var meals []notificationInfo

	for _, vendor := range dining.Vendors {
		// We only want operating vendors
		if !vendor.Operating {
			continue
		}
		for _, meal := range vendor.Meals {
			for _, course := range meal.Courses {
				for _, item := range course.Items {
					if strings.Contains(strings.ToLower(item.Name), keyword) {
						meals = append(meals, notificationInfo{
							Item:   item.Name,
							Course: course.Name,
							Meal:   meal.Name,
							Vendor: vendor.Name,
						})
					}
				}
			}
		}
	}
	if len(meals) == 0 {
		return nil, nil
	}

	return dataToNotification(meals, keyword)
}

func dataToNotification(m []notificationInfo, keyword string) (*notification.Notification, error) {
	var body_elems []string
	for _, meal := range m {
		var data_to_notif string = fmt.Sprintf("%s %s serving %s", meal.Vendor, meal.Meal, meal.Item)
		body_elems = append(body_elems, data_to_notif)
	}
	var body = strings.Join(body_elems, ", ")

	return &notification.Notification{
		Title: fmt.Sprintf("keyword \"%s\" found on the menu today!", keyword),
		// should look like "Found at driscoll lunch serving halal orange chicken. whitmans breakfast serving orange!!!"
		Body: fmt.Sprintf("Found at %s.", body),
	}, nil
}
