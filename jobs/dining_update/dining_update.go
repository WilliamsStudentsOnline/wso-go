package dining_update

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	nutrisliceapi "github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/nutrislice/api"
	nutrisliceparse "github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/nutrislice/parse"
)

var (
	ErrorMissingMenu      = errors.New("missing menu for date")
	ErrorMissingMenuHours = errors.New("failed to get meal hours")
)

type ExportDining struct {
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
	var ed ExportDining
	var err error

	ed, err = loadDiningNutrislice(vendorInfoPath, time.Now())
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


func loadDiningNutrislice(vendorInfoPath string, date time.Time) (ExportDining, error) {
	vendorsInfo, err := ReadVendorInfo(vendorInfoPath)
	if err != nil {
		return ExportDining{}, err
	}

	api := nutrisliceapi.CreateNutriSliceAPI()

	ed := ExportDining{
		Vendors: make(map[string]Vendor),
	}

	// process each vendor that uses nutrislice
	vendorIDs := []string{"driscoll", "whitmans", "mission", "82-grill", "fresh-n-go", "lees-snack-bar", "eco-cafe"}

	for _, vendorID := range vendorIDs {
		vendorInfo, ok := vendorsInfo[vendorID]
		if !ok {
			continue
		}

		vendor, err := loadVendorNutrislice(date, vendorInfo, api)
		if err != nil && (err != ErrorMissingMenu && err != ErrorMissingMenuHours) {
			return ExportDining{}, err
		}
		if err == nil {
			ed.Vendors[vendorID] = *vendor
		}
	}

	// load hours for non nutrislice dining
	for viID, vi := range vendorsInfo {

		if _, ok := ed.Vendors[viID]; ok {
			continue
		}

		nv := Vendor{
			Name:        vi.Name,
			Meals:       make(map[string]*Meal),
			OnlineOrder: vi.OnlineOrder,
			Operating:   vi.Operating,
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

func loadVendorNutrislice(date time.Time, vendorInfo VendorInfo, api *nutrisliceapi.NutriSliceAPI) (*Vendor, error) {
	// get hours
	dayOfWeek := strings.ToLower(date.Weekday().String())
	hours, ok := vendorInfo.Hours[dayOfWeek]
	if !ok {
		return nil, ErrorMissingMenuHours
	}

	meals := make(map[string]*Meal)

	// make map for parsing
	parseHours := make(nutrisliceparse.MealHoursMap)
	for mealName, mealHours := range hours {
		parseHours[mealName] = nutrisliceparse.MealHours{
			Open:  mealHours.Open,
			Close: mealHours.Close,
		}
	}

	// call api for each meal
	for mealName := range hours {
		jsonData, err := api.GetWeeklyMenu(vendorInfo.NutriSliceSlug, mealName, date)
		if err != nil {
			// on api fail just put hours in
			mealHours, hoursOk := hours[mealName]
			if hoursOk {
				meals[mealName] = &Meal{
					Name: mealName,
					Hours: &Hours{
						Open:  mealHours.Open,
						Close: mealHours.Close,
					},
					Courses: make(map[string]*Course),
				}
			}
			continue
		}

		// parse meal from json
		parsedMeal, err := nutrisliceparse.ParseWeeklyMenuToMeal(jsonData, mealName, date, parseHours)
		if err != nil {
			// if cant parse, just put hours 
			mealHours, hoursOk := hours[mealName]
			if hoursOk {
				meals[mealName] = &Meal{
					Name: mealName,
					Hours: &Hours{
						Open:  mealHours.Open,
						Close: mealHours.Close,
					},
					Courses: make(map[string]*Course),
				}
			}
			continue
		}

		// convert from parsed structs to structs to be returned
		meal := &Meal{
			Name:    parsedMeal.Name,
			Hours:   nil,
			Courses: make(map[string]*Course),
		}

		if parsedMeal.Hours != nil {
			meal.Hours = &Hours{
				Open:  parsedMeal.Hours.Open,
				Close: parsedMeal.Hours.Close,
			}
		}

		for courseName, parsedCourse := range parsedMeal.Courses {
			course := &Course{
				Name:  parsedCourse.Name,
				Items: make([]*Food, 0, len(parsedCourse.Items)),
			}

			for _, parsedFood := range parsedCourse.Items {
				course.Items = append(course.Items, &Food{
					Name:       parsedFood.Name,
					Vegetarian: parsedFood.Vegetarian,
					Vegan:      parsedFood.Vegan,
					GlutenFree: parsedFood.GlutenFree,
				})
			}

			meal.Courses[courseName] = course
		}

		meals[mealName] = meal
	}

	// build out vendor struct
	vendor := Vendor{
		Name:        vendorInfo.Name,
		Meals:       meals,
		OnlineOrder: vendorInfo.OnlineOrder,
		Operating:   vendorInfo.Operating,
	}

	return &vendor, nil
}
