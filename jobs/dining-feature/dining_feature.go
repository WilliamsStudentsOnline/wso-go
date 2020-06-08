package main

import(
	"fmt"
	// "strings"
	// "log"
	// "io"
	"encoding/json"
	"io/ioutil"
	// "os"
	"net/http"

	// "golang.org/x/net/html"
)

// Holds the information relevant to a dining hall's menu
type Menu struct {
	dining_hall String
	courses []Course
}

// Holds info relevant to a course (e.g. Appetizers or Entrees will be comprised of many meals)
type Course struct {
	meals []Meal
}

// Holds the information relevant to a meal
type Meal struct {
	meal_name string
	contains [10]string
	serving_size string
	price float32
}

func main(){
	parsedHTTP, err := getXML("https://dining.williams.edu/eats4ephs/")
	fmt.Errorf("Error: &v", err)
	fmt.Println(parsedHTTP)
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

    return string(data), nil
}
