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

type salmonMeal struct {
	Item   string
	Course string
	Meal   string
	Vendor string
}

func SalmonNotify(cfg *config.Config, db *gorm.DB, log *zap.SugaredLogger) error {
	sNtf, err := generateNotif(cfg.DiningFile)
	if err != nil {
		return err
	}

	// No salmon on menu today, so just return
	if sNtf == nil {
		return nil
	}

	// Just start up APNS for now, no Android
	notifTokenModel := models.NewNotificationTokenModel(db, log)

	var tokens []*models.NotificationToken
	err = notifTokenModel.GetTokensWhereSalmonNotif(&tokens)
	if err != nil {
		return err
	}

	notifClient, err := notification.NewClient(cfg)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		if token.Type != models.NotificationTokenTypeIOS {
			continue
		}

		delToken, notifErr := notifClient.NotifyIOS(sNtf, token.Token)
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

	return nil
}

func generateNotif(diningFile string) (*notification.Notification, error) {
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
	if !diningFileTime.Truncate(24 * time.Hour).Equal(time.Now().Truncate(24 * time.Hour)) {
		return nil, errors.New("dining file is out of date")
	}

	// List of found salmon mapped (vendor.meal) to list of salmonMeals (each item)
	var salmonMeals []salmonMeal
	// Search for salmon
	for _, vendor := range dining.Vendors {
		// We only want operating vendors
		if !vendor.Operating {
			continue
		}
		for _, meal := range vendor.Meals {
			for _, course := range meal.Courses {
				for _, item := range course.Items {
					if strings.Contains(strings.ToLower(item.Name), "salmon") {
						salmonMeals = append(salmonMeals, salmonMeal{
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

	if len(salmonMeals) == 0 {
		return nil, nil
	}

	return dataToNotif(salmonMeals)
}

func dataToNotif(m []salmonMeal) (*notification.Notification, error) {
	vendMealToItemMap := make(map[string][]salmonMeal)
	for _, meal := range m {
		v := vendMealToItemMap[meal.Vendor+"."+meal.Meal]
		vendMealToItemMap[meal.Vendor+"."+meal.Meal] = append(v, meal)
	}

	// Only one dhall/meal combo
	if len(vendMealToItemMap) == 1 {
		key := ""
		for k := range vendMealToItemMap {
			key = k
		}

		if len(vendMealToItemMap[key]) < 1 {
			return nil, errors.New("missing item in map, should never happen")
		}

		var itemsSlc []string
		for _, meal := range vendMealToItemMap[key] {
			itemsSlc = append(itemsSlc, meal.Item)
		}

		return &notification.Notification{
			Title: fmt.Sprintf("Salmon at %s %s!", vendMealToItemMap[key][0].Vendor, vendMealToItemMap[key][0].Meal),
			Body:  fmt.Sprintf("%s will be serving %s during %s.", vendMealToItemMap[key][0].Vendor, strings.Join(itemsSlc, ", "), vendMealToItemMap[key][0].Meal),
		}, nil
	}

	// Otherwise
	vendToMealToItemMap := make(map[string]map[string][]salmonMeal)
	for _, meal := range m {
		v, ok := vendToMealToItemMap[meal.Vendor]
		if !ok {
			v = make(map[string][]salmonMeal)
		}
		mealList := v[meal.Meal]
		v[meal.Meal] = append(mealList, meal)
		vendToMealToItemMap[meal.Vendor] = v
	}

	var vendorStrLs []string
	for _, vendMeals := range vendToMealToItemMap {
		vendName := ""
		var vendMealsStr []string
		for _, mealList := range vendMeals {
			vendName = mealList[0].Vendor
			vendMealsStr = append(vendMealsStr, mealList[0].Meal)
		}

		vendorStrLs = append(vendorStrLs, fmt.Sprintf("%s (%s)", vendName, strings.Join(vendMealsStr, "/")))
	}

	return &notification.Notification{
		Title: "Salmon on the menu today!",
		// Will look like: "Salmon at Discoll (lunch/dinner), Whitman's (express/lunch/dinner)!"
		Body: fmt.Sprintf("Dining Services will be serving salmon at %s.", strings.Join(vendorStrLs, ", ")),
	}, nil
}
