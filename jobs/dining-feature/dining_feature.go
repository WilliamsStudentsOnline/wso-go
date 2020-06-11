package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	// "strings"
	// "log"
	// "io"
	// "os"
	// "golang.org/x/net/html"
)

const (
	// MENU_URL stores the endpoint for the daily menu
	MENU_URL       = "https://dining.williams.edu/wp-json/dining/menus"
	DRAFT_MENU_URL = ""
)

// Holds the information relevant to the menu
type Menu struct {
	dining_halls []*DiningHall
}

type DiningHall struct {
	dining_hall_name string
	meals            []*Meal
}

// Holds the information relevant to a meal (e.g. Breakfast, Dinner)
type Meal struct {
	meal_name string
	courses   []*Course
}

// Holds info relevant to a course (e.g. Appetizers or Entrees will be comprised of many meals)
type Course struct {
	course_name string
	foods       []*Food
}

// Holds the information relevant to food
type Food struct {
	food_name    string
	contains     []string // allergens -- (e.g. soy, wheat, nuts)
	serving_size float64
	unit         string // (e.g. oz, cups, etc.)
	price        float64
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
	(*driscoll).dining_hall_name = "driscoll"

	var mission *DiningHall = new(DiningHall)
	(*mission).dining_hall_name = "mission"

	var paresky *DiningHall = new(DiningHall)
	(*paresky).dining_hall_name = "paresky"

	// define the slice
	menu.dining_halls = []*DiningHall{driscoll, mission, paresky}

	// begin traversal of rawMeals
	for _, rawMeal := range rawMeals {
		var food *Food = new(Food)

		// PARSE NAME
		(*food).food_name = rawMeal.FormalName

		// PARSE SERVING_SIZE AND UNIT
		arr := strings.Split(rawMeal.PortionSize, " ")
		if len(arr) == 1 {
			// trim "." such as oz or oz.
			(*food).unit = strings.TrimSuffix(strings.ToLower(arr[0]), ".")
		} else {
			var err error
			(*food).serving_size, err = strconv.ParseFloat(arr[0], 32)

			if err != nil {
				fmt.Println(err)
			}
			(*food).unit = strings.TrimSuffix(strings.ToLower(arr[1]), ".")
		}

		// PARSE CONTAINS ***
		// PARSE PRICE *****

		switch sv := rawMeal.ServiceUnit; sv {
		case "Driscoll Dining Hall":
			var newMeal *Meal
			var newCourse *Course

			// if the meal name already exists in driscoll's meals, do not add it
			mealExists := mealExists(rawMeal.Meal, driscoll.meals)

			// check if the meal exists
			if !mealExists {
				// if the meal doesn't exist, allocate memory for it, set name, append
				newMeal = new(Meal)
				(*newMeal).meal_name = rawMeal.Meal
				driscoll.meals = append(driscoll.meals, newMeal)
			} else {
				// if the meal exists, get the meal object
				newMeal = getMeal(rawMeal.Meal, driscoll.meals)
			}

			courseExists := courseExists(rawMeal.Course, newMeal.courses)
			// check if the course exists
			if !courseExists {
				// if the course doesn't exist, allocate memory for it, set name, append
				newCourse = new(Course)
				(*newCourse).course_name = rawMeal.Course
				newMeal.courses = append(newMeal.courses, newCourse)
			} else {
				// does this return a copy of the course object or the course itself? A COPY
				newCourse = getCourse(rawMeal.Course, newMeal.courses)
			}

			// append the food to the course
			newCourse.foods = append(newCourse.foods, food)

		case "Mission Dining Hall": // ?
		case "Paresky Dining Hall": // ?
		}

	} // end for loop

	return menu, nil
}

func getMeal(mealName string, diningHallMeals []*Meal) *Meal {
	var nilMeal *Meal
	for _, existingMeal := range diningHallMeals {
		if mealName == (*existingMeal).meal_name {
			return existingMeal
		}
	}
	return nilMeal
}

func mealExists(mealName string, diningHallMeals []*Meal) bool {
	for _, existingMeal := range diningHallMeals {
		if mealName == (*existingMeal).meal_name {
			return true
		}
	}
	return false
}

func getCourse(courseName string, mealCourses []*Course) *Course {
	var nilCourse *Course
	for _, existingCourse := range mealCourses {
		if courseName == (*existingCourse).course_name {
			return existingCourse
		}
	}
	return nilCourse
}

func courseExists(courseName string, mealCourses []*Course) bool {
	for _, existingCourse := range mealCourses {
		if courseName == (*existingCourse).course_name {
			return true
		}
	}
	return false
}

func main() {
	// parsedHTTP, err := getXML("https://dining.williams.edu/wp-json/dining/menus")
	rawMeals, err := GetRawMeals(false)
	fmt.Errorf("Error: &v", err)
	menu, err := ParseMenu(rawMeals)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(menu)
	spew.Dump(menu.dining_halls)

}

func getXML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Read body: %v", err)
	}
	fmt.Println(string(data))
	return string(data), nil
}
