package main

import (
	//"fmt"
	events "github.com/WilliamsStudentsOnline/wso-go/jobs/events-feature"
)

func main() {
	rawCategories := events.GetRawCategories()

  categories := events.ParseDailyMessages(rawCategories)

  DumpCategories(categories)

}
