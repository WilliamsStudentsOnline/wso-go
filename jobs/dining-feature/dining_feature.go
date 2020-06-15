package dining_feature

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	// MENU_URL stores the endpoint for the daily menu
	MENU_URL       = "https://dining.williams.edu/wp-json/dining/menus"
	DRAFT_MENU_URL = ""
)

// Holds the information relevant to the menu
type Menu struct {
	DiningHalls []*DiningHall
}

// Holds the information relevant to a dining hall
type DiningHall struct {
	DiningHallName string
	Meals          []*Meal
}

// Holds the information relevant to a meal (e.g. Breakfast, Dinner)
type Meal struct {
	MealName string
	Courses  []*Course
}

// Holds info relevant to a course (e.g. Appetizers or Entrees will be comprised of many Meals)
type Course struct {
	CourseName string
	Foods      []*Food
}

// Holds the information relevant to food
type Food struct {
	FoodName    string
	Contains    []string // allergens -- (e.g. soy, wheat, nuts) // TO IMPLEMENT
	ServingSize float64
	Unit        string  // (e.g. oz, cups, etc.)
	Price       float64 // TO IMPLEMENT
}

// RawMeal represents the unparsed course information we get from the dining-menu endpoint.
type RawMeal struct {
	ServiceUnit  string `json:"service_unit"`
	UnitID       string `json:"unitid"`
	Course       string `json:"course"`
	FormalName   string `json:"formal_name"`
	Meal         string `json:"meal"`
	PortionSize  string `json:"portion_size"`
	NetNutrition string `json:"net_nutrition"`
}

// GetRawMeals fetches the json from the MenuURL endpoint and parses it into an array of RawMeals
func GetRawMeals(draft bool) ([]RawMeal, error) {
	menuClient := &http.Client{
		Timeout: time.Second * 30, // maximum of 30s
	}
	var url string
	if draft {
		url = DRAFT_MENU_URL
		fmt.Println("DRAFT_URL: " + url + "")
	} else {
		url = MENU_URL
		fmt.Println("URL: " + url + "")
	}
	// send a GET request and get back the response
	res, err := menuClient.Get(url)
	if err != nil {
		return nil, err
	}
	var rawMeals []RawMeal
	err = json.NewDecoder(res.Body).Decode(&rawMeals)

	if err != nil {
		return nil, err
	}
	return rawMeals, nil
}

// ParseMenu processes the RawMeal data taken from the JSON endpoint to create a Menu object
func ParseMenu(rawMeals []RawMeal) (Menu, error) {
	var menu Menu

	// declare dining halls
	var driscoll *DiningHall = new(DiningHall)
	(*driscoll).DiningHallName = "Driscoll"

	var mission *DiningHall = new(DiningHall)
	(*mission).DiningHallName = "Mission"

	var paresky *DiningHall = new(DiningHall)
	(*paresky).DiningHallName = "Paresky"

	// define the slice
	menu.DiningHalls = []*DiningHall{driscoll, mission, paresky}

	// begin traversal of rawMeals
	for _, rawMeal := range rawMeals {
		var food *Food = new(Food)
		// reformat the formal name of the food
		rawMeal.FormalName = reformat(rawMeal.FormalName)
		// PARSE NAME
		(*food).FoodName = rawMeal.FormalName

		// PARSE SERVING_SIZE AND UNIT
		arr := strings.Split(rawMeal.PortionSize, " ")
		if len(arr) == 1 {
			// trim "." such as oz or oz.
			(*food).Unit = strings.TrimSuffix(strings.ToLower(arr[0]), ".")
		} else {
			var err error
			// ignore potentially non formatted human input
			arr = strings.Split(findFirstDigit(rawMeal.PortionSize), " ")
			// the length of the array should be 2 at this point
			(*food).ServingSize, err = strconv.ParseFloat(arr[0], 64)

			if err != nil {
				fmt.Println(err)
			}
			(*food).Unit = strings.TrimSuffix(strings.ToLower(arr[arr.length-1]), ".")
		}

		// TODO : get allergens
		// TODO: get price to IMPLEMENT

		switch sv := rawMeal.ServiceUnit; sv {
		case "Driscoll Dining Hall":
			addMeal(driscoll, rawMeal, food)
		case "Mission Dining Hall":
			addMeal(mission, rawMeal, food)
		case "Paresky Whitmans Market":
			addMeal(paresky, rawMeal, food)
		default:
			fmt.Printf("\nDining Hall %s does not exist.\n", sv)
		}

	} // end for loop

	return menu, nil
}

func addMeal(diningHall *DiningHall, rawMeal RawMeal, food *Food) {
	var newMeal *Meal
	var newCourse *Course
	rawMeal.Meal = capitalizeFirst(rawMeal.Meal)
	rawMeal.Course = reformat(rawMeal.Course)

	// if the meal name already exists in driscoll's Meals, do not add it
	mealExists := mealExists(rawMeal.Meal, (*diningHall).Meals)

	// check if the meal exists
	if !mealExists {
		// if the meal doesn't exist, allocate memory for it, set name, append
		newMeal = new(Meal)
		(*newMeal).MealName = rawMeal.Meal
		(*diningHall).Meals = append((*diningHall).Meals, newMeal)
	} else {
		// if the meal exists, get the meal object's pointer
		newMeal = getMeal(rawMeal.Meal, (*diningHall).Meals)
	}

	courseExists := courseExists(rawMeal.Course, newMeal.Courses)
	// check if the course exists
	if !courseExists {
		// if the course doesn't exist, allocate memory for it, set name, append
		newCourse = new(Course)
		(*newCourse).CourseName = rawMeal.Course
		newMeal.Courses = append(newMeal.Courses, newCourse)
	} else {
		// if the course exists, get the course object's pointer
		newCourse = getCourse(rawMeal.Course, newMeal.Courses)
	}
	// append the food to the course
	newCourse.Foods = append(newCourse.Foods, food)
}

// reformats words that are all uppercase to have 0th letter of each word capital
func reformat(name string) string {
	for _, ch := range name {
		if ch == ' ' {
			continue
		}
		// found lowercase letter, no need to reformat
		if !(unicode.IsUpper(ch)) {
			return name
		}
	}
	var formatted string
	arr := strings.Split(name, " ")
	// capitalize each word in the array and create formatted
	for _, word := range arr {
		formatted += capitalizeFirst(word) + " "
	}
	return formatted
}

func findFirstDigit(str string) string {
	for i, ch := range str {
		if unicode.IsDigit(ch) {
			fmt.Println("returning" + str[i:])
			return str[i:]
		}
	}
	// if unable to find a digit, return the string as is
	return str
}

func capitalizeFirst(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
}

func getMeal(mealName string, diningHallMeals []*Meal) *Meal {
	var nilMeal *Meal
	for _, existingMeal := range diningHallMeals {
		if mealName == (*existingMeal).MealName {
			return existingMeal
		}
	}
	return nilMeal
}

func mealExists(mealName string, diningHallMeals []*Meal) bool {
	for _, existingMeal := range diningHallMeals {
		if mealName == (*existingMeal).MealName {
			return true
		}
	}
	return false
}

func getCourse(courseName string, mealCourses []*Course) *Course {
	var nilCourse *Course
	for _, existingCourse := range mealCourses {
		if courseName == (*existingCourse).CourseName {
			return existingCourse
		}
	}
	return nilCourse
}

func courseExists(courseName string, mealCourses []*Course) bool {
	for _, existingCourse := range mealCourses {
		if courseName == (*existingCourse).CourseName {
			return true
		}
	}
	return false
}

// takes the Menu struct, exports data the JSON and writes it to a local file
//	returns date as well
func WriteToJSON(menu Menu) (string, error) {
	menuJSON, _ := json.Marshal(&menu)
	err := ioutil.WriteFile("dining_data.json", menuJSON, 0644)
	dt := time.Now()

	return dt.Format("01-02-2006 15:04:05 Mon"), err
}
