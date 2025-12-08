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

	SkipCourses = map[string]bool{
		"Drinks":             true,
		"Condiments":         true,
		"Salad Condiments":   true,
		"Cereal":             true,
		"Hot Cereal Station": true,
		"Salad Bar A":        true,
		"Salad Bar B":        true,
		"MP Salad Bar A":     true,
		"MP Salad Bar B":     true,
		"MP Lettuce Area":    true,
		"MP Salad Dressings": true,
		"Daily Salad":        true,
		"":        true, // idk why this would ever be the case but who knows
	}

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
	Icons icons `json:"icons"`
}

type icons struct {
	FoodIcons []foodIcon `json:"food_icons"`
}

type foodIcon struct {
	SyncedName string `json:"synced_name"`
}


func ParseFoodItem(menuItem *food) *ParsedFood {
	foodName := strings.TrimSpace(menuItem.Name)
	if foodName == "" {
		return nil
	}

	parsedFood := &ParsedFood{
		Name:       foodName,
		Vegetarian: false,
		Vegan:      false,
		GlutenFree: false,
	}

	for _, icon := range menuItem.Icons.FoodIcons {
		switch icon.SyncedName {
		case "VEGT":
			parsedFood.Vegetarian = true
		case "V", "VG", "Vegan":
			parsedFood.Vegan = true
		case "GF", "Gluten Free":
			parsedFood.GlutenFree = true
		}
	}

	return parsedFood
}


// ParseWeeklyMenuToMeals parses a weekly menu JSON response and returns meals for all available dates
// Returns a map of date string (YYYY-MM-DD) -> ParsedMeal
func ParseWeeklyMenuToMeals(jsonData []byte, mealName string, vendorHours map[string]MealHoursMap) (map[string]*ParsedMeal, error) {
	var response weeklyMenuResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	meals := make(map[string]*ParsedMeal)

	for _, dayData := range response.Days {
		parsedDate, err := time.Parse("2006-01-02", dayData.Date)
		if err != nil {
			continue
		}

		dayOfWeek := strings.ToLower(parsedDate.Weekday().String())

		var parsedHours *MealHours
		if dayHours, ok := vendorHours[dayOfWeek]; ok {
			if mealHours, ok := dayHours[mealName]; ok {
				parsedHours = &MealHours{
					Open:  mealHours.Open,
					Close: mealHours.Close,
				}
			}
		}

		if parsedHours == nil {
			continue
		}

		meal := &ParsedMeal{
			Name:    mealName,
			Hours:   parsedHours,
			Courses: make(map[string]*ParsedCourse),
		}

		var currentCourse *ParsedCourse
		for _, item := range dayData.MenuItems {
			if item.BlankLine {
				continue
			}

			// seems like is_section_titel adn is_sattion_header are the same
			// in the json but honeslty who knows what dining is on
			if item.IsSectionTitle || item.IsStationHeader {
				courseName := strings.TrimSpace(item.Text)
				if SkipCourses[courseName] {
					currentCourse = nil
					continue
				}
				currentCourse = meal.Courses[courseName]
				if currentCourse == nil {
					currentCourse = &ParsedCourse{
						Name:  courseName,
						Items: []*ParsedFood{},
					}
					meal.Courses[courseName] = currentCourse
				}				
			} else if (item.Food != nil && item.Food.Name != "" && currentCourse != nil) {
				if parsedFood := ParseFoodItem(item.Food); parsedFood != nil {
					currentCourse.Items = append(currentCourse.Items, parsedFood)
				}
			}
		}

		meals[dayData.Date] = meal
	}

	return meals, nil
}
