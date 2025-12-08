package dining_update

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var (
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

type MealHoursMap map[string]Hours

type NutriScliceWeeklyMenuResponse struct {
	Days []NutriSliceDayData `json:"days"`
}

type NutriSliceDayData struct {
	Date      string     `json:"date"`
	MenuItems []NutrisliceMenuItem `json:"menu_items"`
}

type NutrisliceMenuItem struct {
	IsSectionTitle  bool     `json:"is_section_title"`
	IsStationHeader bool     `json:"is_station_header"`
	BlankLine       bool     `json:"blank_line"`
	Text            string   `json:"text"`
	Category        string   `json:"category"`
	Food            *NutrisliceFood    `json:"food"`
	StationFoodTags []string `json:"station_food_tags"`
}

type NutrisliceFood struct {
	Name string `json:"name"`
	Icons NutrisliceIcons `json:"icons"`
	ID int						`json:"id"`
}

type NutrisliceIcons struct {
	FoodIcons []NutrisliceFoodIcon `json:"food_icons"`
}

type NutrisliceFoodIcon struct {
	SyncedName string `json:"synced_name"`
}


func parseFoodItem(nutrisliceFood *NutrisliceFood) *Food {
	foodName := strings.TrimSpace(nutrisliceFood.Name)
	if foodName == "" {
		return nil
	}

	parsedFood := &Food{
		Name:       foodName,
		ID: 								nutrisliceFood.ID,
		Vegetarian: false,
		Vegan:      false,
		GlutenFree: false,
	}

	for _, icon := range nutrisliceFood.Icons.FoodIcons {
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
func parseWeeklyMenuToMeals(jsonData []byte, mealName string, vendorHours map[string]MealHoursMap) (map[string]*Meal, error) {
	var response NutriScliceWeeklyMenuResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	meals := make(map[string]*Meal)

	for _, dayData := range response.Days {
		parsedDate, err := time.Parse("2006-01-02", dayData.Date)
		if err != nil {
			continue
		}

		dayOfWeek := strings.ToLower(parsedDate.Weekday().String())

		var parsedHours *Hours
		if dayHours, ok := vendorHours[dayOfWeek]; ok {
			if mealHours, ok := dayHours[mealName]; ok {
				parsedHours = &Hours{
					Open:  mealHours.Open,
					Close: mealHours.Close,
				}
			}
		}

		if parsedHours == nil {
			continue
		}

		meal := &Meal{
			Name:    mealName,
			Hours:   parsedHours,
			Courses: make(map[string]*Course),
		}

		var currentCourse *Course
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
					currentCourse = &Course{
						Name:  courseName,
						Items: []*Food{},
					}
					meal.Courses[courseName] = currentCourse
				}				
			} else if (item.Food != nil && item.Food.Name != "" && currentCourse != nil) {
				if parsedFood := parseFoodItem(item.Food); parsedFood != nil {
					currentCourse.Items = append(currentCourse.Items, parsedFood)
				}
			}
		}

		meals[dayData.Date] = meal
	}

	for _, dayHours := range vendorHours {
		for mType := range dayHours {
			if _, ok := meals[mType]; !ok {
				meals[mType] = &Meal{
					Name:    mType,
					Hours:   &Hours{Open: dayHours[mType].Open, Close: dayHours[mType].Close},
					Courses: map[string]*Course{},
				}
			}
		}
	}

	return meals, nil
}
