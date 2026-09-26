package dining_update

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
	"path/filepath"

	nutrisliceapi "github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/api"
)

var (
	ErrorMissingMenu      = errors.New("missing menu for date")
	ErrorMissingMenuHours = errors.New("failed to get meal hours")
)

type WeeklyDiningInfo struct {
	VendorsByDate map[string]map[string]Vendor `json:"vendors"`
	UpdateTime    string                       `json:"updateTime"`
}

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
	ID 		   int    `json:"id"`
	Vegetarian bool   `json:"vegetarian"`
	Vegan      bool   `json:"vegan"`
	GlutenFree bool   `json:"glutenFree"`
}

func UpdateDining(outputDir string, vendorInfoPath string, fetchDate time.Time) error {
	weeklyDiningInfo, err := loadDiningNutrislice(vendorInfoPath, fetchDate)
	if err != nil {
		return err
	}

	for dateStr, vendors := range weeklyDiningInfo.VendorsByDate {
		filename := filepath.Join(outputDir, dateStr + ".json")

		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		fileContent := ExportDining{
			Vendors: vendors,
			UpdateTime: weeklyDiningInfo.UpdateTime,
		}

		if err := json.NewEncoder(f).Encode(fileContent); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

func loadDiningNutrislice(vendorInfoPath string, startDate time.Time) (WeeklyDiningInfo, error) {
	vendorsInfo, err := ReadVendorInfo(vendorInfoPath)
	if err != nil {
		return WeeklyDiningInfo{}, err
	}

	api := nutrisliceapi.CreateNutriSliceAPI()

	weeklyDiningInfo := WeeklyDiningInfo{
		VendorsByDate: make(map[string]map[string]Vendor),
	}

	nutriVendorIDs := map[string]bool{}
	for vendorID, vendorInfo := range vendorsInfo {
		if vendorInfo.NutriSliceSlug == "" {
			continue
		}
		nutriVendorIDs[vendorID] = true

		// weekly parsed vendors: map[date]Vendor
		vendorsByDate, err := loadVendorNutrisliceWeekly(startDate, vendorInfo, api)
		if err != nil {
			return WeeklyDiningInfo{}, err
		}

		// insert each date
		for dateStr, vendor := range vendorsByDate {
			if weeklyDiningInfo.VendorsByDate[dateStr] == nil {
				weeklyDiningInfo.VendorsByDate[dateStr] = make(map[string]Vendor)
			}
			weeklyDiningInfo.VendorsByDate[dateStr][vendorID] = *vendor
		}
	}

	// non nutrislice vendors (hours-only, e.g. Goodrich)
	for viID, vi := range vendorsInfo {
		if nutriVendorIDs[viID] {
			continue
		}

		dateStr := startDate.Format("2006-01-02")

		if weeklyDiningInfo.VendorsByDate[dateStr] == nil {
			weeklyDiningInfo.VendorsByDate[dateStr] = make(map[string]Vendor)
		}

		nv := Vendor{
			Name:        vi.Name,
			Meals:       make(map[string]*Meal),
			OnlineOrder: vi.OnlineOrder,
			Operating:   vi.Operating,
		}

		dow := strings.ToLower(startDate.Weekday().String())
		if mealsToHours, ok := vi.Hours[dow]; ok {
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

		weeklyDiningInfo.VendorsByDate[dateStr][viID] = nv
	}

	weeklyDiningInfo.UpdateTime = time.Now().Format(time.RFC850)
	return weeklyDiningInfo, nil
}

func loadVendorNutrisliceWeekly(
	startDate time.Time,
	vendorInfo VendorInfo,
	api *nutrisliceapi.NutriSliceAPI,
) (map[string]*Vendor, error) {

	vendorHoursByDay := make(map[string]MealHoursMap)
	for dow, hours := range vendorInfo.Hours {
		m := make(MealHoursMap)
		for mealName, h := range hours {
			m[mealName] = Hours{Open: h.Open, Close: h.Close}
		}
		vendorHoursByDay[dow] = m
	}

	mealTypes := map[string]bool{}
	for _, hours := range vendorInfo.Hours {
		for mealName := range hours {
			mealTypes[mealName] = true
		}
	}

	mealsByDate := make(map[string]map[string]*Meal)

	for mealType := range mealTypes {
		apiMealType := nutriSliceMenuType(vendorInfo, mealType)
		raw, err := api.GetWeeklyMenu(vendorInfo.NutriSliceSlug, apiMealType, startDate)
		if err != nil {
			continue
		}

		parsed, err := parseWeeklyMenuToMeals(raw, mealType, vendorHoursByDay)
		if err != nil {
			continue
		}

		for dateStr, pm := range parsed {
			if mealsByDate[dateStr] == nil {
				mealsByDate[dateStr] = make(map[string]*Meal)
			}
			mealsByDate[dateStr][mealType] = pm
		}
	}

	vendorsByDate := make(map[string]*Vendor)

	for dateStr, dateMeals := range mealsByDate {

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil { continue }
		dow := strings.ToLower(date.Weekday().String())
		dayHours := vendorInfo.Hours[dow]

		meals := make(map[string]*Meal)

		for mealType, pm := range dateMeals {

			displayName := mealType
			if dowHours, ok := vendorInfo.Hours[dow]; ok {
				if mealHours, ok := dowHours[mealType]; ok && mealHours.DisplayName != "" {
					displayName = mealHours.DisplayName
				}
			}
			m := &Meal{
				Name:    displayName,
				Courses: make(map[string]*Course),
			}
			if pm.Hours != nil {
				m.Hours = &Hours{Open: pm.Hours.Open, Close: pm.Hours.Close}
			}

			for courseName, pc := range pm.Courses {
				c := &Course{Name: pc.Name}
				for _, pf := range pc.Items {
					c.Items = append(c.Items, &Food{
						Name:       pf.Name,
						Vegetarian: pf.Vegetarian,
						Vegan:      pf.Vegan,
						GlutenFree: pf.GlutenFree,
					})
				}
				m.Courses[courseName] = c
			}

			meals[mealType] = m
		}

		// add empty meal info
		for mealType, h := range dayHours {
			if _, ok := meals[mealType]; !ok {
				displayName := mealType
				if h.DisplayName != "" {
					displayName = h.DisplayName
				}
				meals[mealType] = &Meal{
					Name:    displayName,
					Hours:   &Hours{Open: h.Open, Close: h.Close},
					Courses: map[string]*Course{},
				}
			}
		}

		vendorsByDate[dateStr] = &Vendor{
			Name:        vendorInfo.Name,
			Meals:       meals,
			OnlineOrder: vendorInfo.OnlineOrder,
			Operating:   len(meals) > 0,
		}
	}

	return vendorsByDate, nil
}

// nutriSliceMenuType returns the Nutrislice menu-type slug for a vendor meal key.
// Prefers an explicit nutrislice_menu_type from vendor hours when present.
func nutriSliceMenuType(vendorInfo VendorInfo, mealType string) string {
	for _, dayHours := range vendorInfo.Hours {
		if h, ok := dayHours[mealType]; ok && h.NutriSliceMenuType != "" {
			return h.NutriSliceMenuType
		}
	}
	return mealType
}
