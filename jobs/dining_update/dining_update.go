package dining_update

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/parse"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
)

type exportDining struct {
	Vendors    map[string]Vendor `json:"vendors"`
	UpdateTime string            `json:"updateTime"`
}

type Vendor struct {
	Name        string           `json:"name"`
	Meals       map[string]*Meal `json:"meals"`
	OnlineOrder bool             `json:"onlineOrder"`
	Operating   bool             `json:"operating"`
}

type Meal struct {
	Name    string             `json:"name"`
	Hours   *Hours             `json:"hours"`
	Courses map[string]*Course `json:"courses"`
}

type Hours struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

type Course struct {
	Name  string  `json:"name"`
	Items []*Food `json:"items"`
}

type Food struct {
	Name       string `json:"name"`
	Vegetarian bool   `json:"vegetarian"`
	Vegan      bool   `json:"vegan"`
	GlutenFree bool   `json:"glutenFree"`
}

func UpdateDining(outPath string, vendorInfoPath string) error {
	ed, err := loadDining(vendorInfoPath, time.Now())
	if err != nil {
		return err
	}

	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(ed)
}

func loadDining(vendorInfoPath string, date time.Time) (exportDining, error) {
	vendorsInfo, err := ReadVendorInfo(vendorInfoPath)
	if err != nil {
		return exportDining{}, err
	}

	d, err := api.CreateWilliamsDiningAPI()
	if err != nil {
		return exportDining{}, err
	}

	venues, err := parse.Populate(d)
	if err != nil {
		return exportDining{}, err
	}

	ed := exportDining{
		Vendors: make(map[string]Vendor),
	}

	drisc, err := loadDriscoll(date, venues["Driscoll"], vendorsInfo["driscoll"])
	if err != nil {
		return exportDining{}, err
	}
	ed.Vendors["driscoll"] = *drisc

	whitmans, err := loadWhitmans(date, venues["Paresky Student Center"], vendorsInfo["whitmans"])
	if err != nil {
		return exportDining{}, err
	}
	ed.Vendors["whitmans"] = *whitmans

	for viID, vi := range vendorsInfo {
		nv := Vendor{
			Name:        vi.Name,
			Meals:       make(map[string]*Meal),
			OnlineOrder: vi.OnlineOrder,
			Operating:   vi.Operating,
		}

		// Ignore doing this again if drisc or whitmans
		if viID == "driscoll" || viID == "whitmans" {
			continue
		}

		mealsToHours, ok := vi.Hours[strings.ToLower(date.Weekday().String())]
		if ok {
			for mealName, hours := range mealsToHours {
				nv.Meals[mealName] = &Meal{
					Name: mealName,
					Hours: &Hours{
						Open:  hours.Open,
						Close: hours.Close,
					},
				}
			}
		}

		ed.Vendors[viID] = nv
	}

	ed.UpdateTime = time.Now().Format(time.RFC850)

	return ed, nil
}

func loadDriscoll(date time.Time, venue *search.Venue, vendorInfo VendorInfo) (*Vendor, error) {
	meals, err := parseDailyMenu(date, venue.DiningHalls[0].Menus["Driscoll Daily Menu"], &vendorInfo)
	if err != nil {
		return nil, err
	}

	vendor := Vendor{
		Name:        vendorInfo.Name,
		Meals:       meals,
		OnlineOrder: vendorInfo.OnlineOrder,
		Operating:   vendorInfo.Operating,
	}

	return &vendor, nil
}

func loadWhitmans(date time.Time, venue *search.Venue, vendorInfo VendorInfo) (*Vendor, error) {
	meals, err := parseDailyMenu(date, venue.DiningHalls[0].Menus["Whitmans' Daily Menu"], &vendorInfo)
	if err != nil {
		return nil, err
	}

	vendor := Vendor{
		Name:        vendorInfo.Name,
		Meals:       meals,
		OnlineOrder: vendorInfo.OnlineOrder,
		Operating:   vendorInfo.Operating,
	}

	return &vendor, nil
}

func parseDailyMenu(date time.Time, dailyMenu search.MetaMenu, vendorInfo *VendorInfo) (map[string]*Meal, error) {
	menuDays, ok := dailyMenu.(*search.DailyMenu)
	if !ok {
		return nil, errors.New("failed to cast daily menu")
	}

	hours, ok := vendorInfo.Hours[strings.ToLower(date.Weekday().String())]
	if !ok {
		return nil, errors.New("failed to get meal hours")
	}

	dayMenu, ok := menuDays.Days[date.Format("Monday, January 2, 2006")]
	if !ok {
		return nil, errors.New("missing menu for date")
	}

	parsedMeals := make(map[string]*Meal)

	// Go thru every meal of the day
	for mealName, menu := range dayMenu {
		meal := Meal{
			Name:    mealName,
			Hours:   nil,
			Courses: make(map[string]*Course),
		}

		// TODO: log when not ok
		mealHours, ok := hours[mealName]
		if ok {
			meal.Hours = &Hours{
				Open:  mealHours.Open,
				Close: mealHours.Close,
			}
		}

		// Go thru every course in the meal
		for courseName, menuItems := range *menu {
			course := Course{
				Name:  courseName,
				Items: []*Food{},
			}

			// Go thru every food item in the course
			for _, itemName := range menuItems {
				itemName = strings.TrimSpace(itemName)
				foodItem := Food{
					Name:       itemName,
					Vegetarian: false,
					Vegan:      false,
					GlutenFree: false,
				}
				if strings.HasSuffix(itemName, "VGT") {
					foodItem.Vegetarian = true
					foodItem.Name = strings.TrimSpace(strings.TrimSuffix(itemName, "VGT"))
				}
				if strings.HasSuffix(itemName, "V") {
					foodItem.Vegan = true
					foodItem.Name = strings.TrimSpace(strings.TrimSuffix(itemName, "V"))
				}
				if strings.HasSuffix(itemName, "GF") {
					foodItem.GlutenFree = true
					foodItem.Name = strings.TrimSpace(strings.TrimSuffix(itemName, "GF"))
				}

				course.Items = append(course.Items, &foodItem)
			}

			meal.Courses[courseName] = &course
		}

		parsedMeals[mealName] = &meal
	}

	return parsedMeals, nil
}
