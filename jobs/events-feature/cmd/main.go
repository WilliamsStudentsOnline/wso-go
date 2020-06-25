package main

import (
	//"fmt"
	events "github.com/WilliamsStudentsOnline/wso-go/jobs/events-feature"
)

func main() {
	rawCategories, _ := events.GetRawCategories()

  categories, _ := events.ParseDailyMessages(rawCategories)

  events.DumpCategories(categories)

}
