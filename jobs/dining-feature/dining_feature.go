package main

import(
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	// "strings"
	// "log"
	// "io"


	// "os"
	// "golang.org/x/net/html"
)

const (
	// MENU_URL stores the endpoint for the daily menu
	MENU_URL      = "https://dining.williams.edu/wp-json/dining/menus"
	DRAFT_MENU_URL = ""
)


// Holds the information relevant to a dining hall's menu
type Menu struct {
	dining_halls []DiningHall
}

type DiningHall struct {
	dining_hall_name String
	meals []Meal
}

// Holds the information relevant to a meal (e.g. Breakfast, Dinner)
type Meal struct {
	meal_name String
	courses []Course
}
// Holds info relevant to a course (e.g. Appetizers or Entrees will be comprised of many meals)
type Course struct {
	course_name String
	foods []Food
}

// Holds the information relevant to food
type Food struct {
	food_name string
	contains [10]string // allergens -- (e.g. soy, wheat, nuts)
	serving_size int
	unit string // (e.g. oz, cups, etc.)
	price float32
}

// RawMeal represents the unparsed course information we get from the dining-menu endpoint.
type RawMeal struct {
	ServiceUnit string `json:"service_unit"`
	UnitID string `json:"unitid"`
	Course string `json:"course"`
	FormalName string `json:"formal_name"`
	Meal string  `json:"meal"`
	PortionSize string `json:"portion_size"`
	NetNutrition string `json:"net_nutrition"`
}

// GetRawMeals fetches the json from the MenuURL endpoint and parses it into an array of RawMeals
func GetRawMeals(draft bool) ([]RawMeal, error) {
	menuClient := &http.Client{
		Timeout: time.Second * 30, // maximum of 30s
	}
	var url string
	if draft {
		url =  DRAFT_MENU_URL
		fmt.Println("DRAFT_URL: " + url +  "")
		} else {
			url = MENU_URL
			fmt.Println("URL: " + url +  "")
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
		// Initialize the slice this way in order to ensure it will never respond as a nil slice
		meals := []Meal{}

		// declare dining halls
		var driscoll DiningHall
		driscoll.name = "driscoll"
		var mission DiningHall
		mission.name = "mission"
		var paresky DiningHall
		paresky.name = "paresky"

		var menu Menu
		menu.dining_halls = []DiningHall{driscoll, mission, paresky}

		for _, rawMeal := range rawMeals {
			var food Food

			// PARSE NAME
			food.food_name = rawMeal.FormalName

			// PARSE SERVING_SIZE AND UNIT
			arr := split(rawMeal.PortionSize, " ")
			if len(arr) == 1{
				food.unit = arr[0]
				} else{
					food.serving_size = arr[0]
					food.unit = arr[1]
				}

				// PARSE CONTAINS ***
				// PARSE PRICE *****

				var newMeal Meal
				switch sv:= rawMeal.ServiceUnit; sv {
				case "Driscoll Dining Hall":
					// if the meal name already exists in driscoll's meals, do not add it
					mealExists := mealExists(rawMeal.meal, driscoll.meals)

					// check if the meal exists
					if !mealExists {
						newMeal.name = rawMeal.meal
						//driscoll.meals = append(driscoll.meals, newMeal)
						} else {
							// if the meal exists, get the meal object

							// does this return a copy of the meal object or the meal itself? A COPY
							newMeal = getMeal(rawMeal.meal, driscoll.meals)
						}

						var newCourse course
						courseExists := courseExists(rawMeal.course, newMeal.courses)
						// check if the course exists
						if !courseExists {
							newCourse.name = rawMeal.course
							//append(newMeal, newMeal.cous)
						} else {
							// does this return a copy of the course object or the course itself? A COPY
							newCourse = getCourse(rawMeal.course, newMeal.courses)
						}

						// append the food to that specific course
						append(food, newCourse.foods)
						// append the course to to the meals
						// append the meal to the dining hall

						case "Mission Dining Hall": // ?
						case "Paresky Dining Hall":  // ?
					}



					meals = append(meals, meal)


					} // end for loop
					fmt.Println(meals)

					var men Menu
					return men, nil



				}

				func getMeal(mealName String, diningHallMeals []Meal) Meal {
					for _,existingMeal := range diningHallMeals {
						if mealName == existingMeal.name{
							return existingMeal
						}
					}
					return nil
				}

				func mealExists(mealName String, diningHallMeals []Meal) bool{
					for _,existingMeal := range diningHallMeals {
						if mealName == existingMeal.name{
							return true
						}
					}
					return false
				}

				func getCourse(courseName String, mealCourses []Course) Course {
					for _,existingCourse := range mealCourses {
						if courseName == existingCourse.name{
							return existingCourse
						}
					}
					return nil
				}

				func courseExists(courseName String, mealCourses []Course) boolean{
					for _,existingCourse := range mealCourses {
						if courseName == existingCourse.name{
							return true
						}
					}
					return false
				}



				func main(){
					// parsedHTTP, err := getXML("https://dining.williams.edu/wp-json/dining/menus")
					rawMeals, err := GetRawMeals(false)
					fmt.Errorf("Error: &v", err)
					menu, _ := ParseMenu(rawMeals)
					fmt.Println(menu)
					//fmt.Printf("RawMeals : \"%+v\"", rawMeals)
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
