package main

import (
	"fmt"
	DiningParser "github.com/WilliamsStudentsOnline/wso-go/jobs/dining-feature"
)

func main() {
	rawMeals, err := DiningParser.GetRawMeals(false)
	if err != nil {
		fmt.Println(err)
	}

	menu, err := DiningParser.ParseMenu(rawMeals)
	if err != nil {
		fmt.Println(err)
	}

	time, err := DiningParser.WriteToJSON(menu)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Success! New menu processed on: %s\n", time)
}
