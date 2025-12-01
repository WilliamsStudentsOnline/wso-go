package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrorMissingMenuHours = errors.New("failed to get meal hours")
)

type MealHours struct {
	Open  string
	Close string
}

type MealHoursMap map[string]MealHours

type ParsedMeal struct {
	Name    string
	Hours   *MealHours
	Courses map[string]*ParsedCourse
}

type ParsedCourse struct {
	Name  string
	Items []*ParsedFood
}

type ParsedFood struct {
	Name       string
	Vegetarian bool
	Vegan      bool
	GlutenFree bool
}

type weeklyMenuResponse struct {
	Days []dayData `json:"days"`
}

type dayData struct {
	Date      string     `json:"date"`
	MenuItems []menuItem `json:"menu_items"`
}

type menuItem struct {
	IsSectionTitle  bool     `json:"is_section_title"`
	IsStationHeader bool     `json:"is_station_header"`
	BlankLine       bool     `json:"blank_line"`
	Text            string   `json:"text"`
	Category        string   `json:"category"`
	Food            *food    `json:"food"`
	StationFoodTags []string `json:"station_food_tags"`
}

type food struct {
	Name string `json:"name"`
}

// json to parsed meal
func ParseWeeklyMenuToMeal(jsonData []byte, mealName string, targetDate time.Time, hours MealHoursMap) (*ParsedMeal, error) {
	mealHours, hoursOk := hours[mealName]
	var parsedHours *MealHours
	if hoursOk {
		parsedHours = &MealHours{
			Open:  mealHours.Open,
			Close: mealHours.Close,
		}
	}

	meal := &ParsedMeal{
		Name:    mealName,
		Hours:   parsedHours,
		Courses: make(map[string]*ParsedCourse),
	}

	var response weeklyMenuResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	targetDateStr := targetDate.Format("2006-01-02")
	var dayData *dayData
	for i := range response.Days {
		if response.Days[i].Date == targetDateStr {
			dayData = &response.Days[i]
			break
		}
	}

	if dayData == nil {
		return meal, nil
	}

	currentCourse := ""

	for _, item := range dayData.MenuItems {

		if item.BlankLine {
			continue
		}

		if item.IsSectionTitle {
			courseName := strings.TrimSpace(item.Text)
			if courseName != "" {
				currentCourse = courseName
				if _, exists := meal.Courses[currentCourse]; !exists {
					meal.Courses[currentCourse] = &ParsedCourse{
						Name:  currentCourse,
						Items: []*ParsedFood{},
					}
				}
			}
			continue
		}

		if item.IsStationHeader && item.Text != "" {
			stationName := strings.TrimSpace(item.Text)
			if stationName != "" {
				currentCourse = stationName
				if _, exists := meal.Courses[currentCourse]; !exists {
					meal.Courses[currentCourse] = &ParsedCourse{
						Name:  currentCourse,
						Items: []*ParsedFood{},
					}
				}
			}
			continue
		}

		if item.Food != nil && item.Food.Name != "" {
			courseName := currentCourse
			if courseName == "" && item.Category != "" {
				courseName = strings.TrimSpace(item.Category)
			}
			if courseName == "" {
				courseName = "Other"
			}

			course, courseExists := meal.Courses[courseName]
			if !courseExists {
				course = &ParsedCourse{
					Name:  courseName,
					Items: []*ParsedFood{},
				}
				meal.Courses[courseName] = course
			}

			foodName := strings.TrimSpace(item.Food.Name)
			if foodName == "" {
				foodName = strings.TrimSpace(item.Text)
			}

			if foodName != "" {
				foodItem := &ParsedFood{
					Name:       foodName,
					Vegetarian: false,
					Vegan:      false,
					GlutenFree: false,
				}

				for _, tag := range item.StationFoodTags {
					tagLower := strings.ToLower(tag)
					if strings.Contains(tagLower, "vegetarian") || strings.Contains(tagLower, "vgt") {
						foodItem.Vegetarian = true
					}
					if strings.Contains(tagLower, "vegan") || strings.Contains(tagLower, " v ") {
						foodItem.Vegan = true
					}
					if strings.Contains(tagLower, "gluten") || strings.Contains(tagLower, "gf") {
						foodItem.GlutenFree = true
					}
				}

				foodNameUpper := strings.ToUpper(foodName)
				if strings.HasSuffix(foodNameUpper, "VGT") {
					foodItem.Vegetarian = true
				}
				if strings.HasSuffix(foodNameUpper, " V") || strings.HasSuffix(foodNameUpper, "V") {
					foodItem.Vegan = true
				}
				if strings.HasSuffix(foodNameUpper, "GF") {
					foodItem.GlutenFree = true
				}

				course.Items = append(course.Items, foodItem)
			}
		} else if item.Text != "" && !item.IsSectionTitle && !item.IsStationHeader {
			
			courseName := currentCourse
			if courseName == "" && item.Category != "" {
				courseName = strings.TrimSpace(item.Category)
			}
			if courseName == "" {
				courseName = "Other"
			}

			course, courseExists := meal.Courses[courseName]
			if !courseExists {
				course = &ParsedCourse{
					Name:  courseName,
					Items: []*ParsedFood{},
				}
				meal.Courses[courseName] = course
			}

			foodName := strings.TrimSpace(item.Text)
			if foodName != "" {
				foodItem := &ParsedFood{
					Name:       foodName,
					Vegetarian: false,
					Vegan:      false,
					GlutenFree: false,
				}

				for _, tag := range item.StationFoodTags {
					tagLower := strings.ToLower(tag)
					if strings.Contains(tagLower, "vegetarian") || strings.Contains(tagLower, "vgt") {
						foodItem.Vegetarian = true
					}
					if strings.Contains(tagLower, "vegan") || strings.Contains(tagLower, " v ") {
						foodItem.Vegan = true
					}
					if strings.Contains(tagLower, "gluten") || strings.Contains(tagLower, "gf") {
						foodItem.GlutenFree = true
					}
				}

				foodNameUpper := strings.ToUpper(foodName)
				if strings.HasSuffix(foodNameUpper, "VGT") {
					foodItem.Vegetarian = true
				}
				if strings.HasSuffix(foodNameUpper, " V") || strings.HasSuffix(foodNameUpper, "V") {
					foodItem.Vegan = true
				}
				if strings.HasSuffix(foodNameUpper, "GF") {
					foodItem.GlutenFree = true
				}

				course.Items = append(course.Items, foodItem)
			}
		}
	}

	return meal, nil
}
