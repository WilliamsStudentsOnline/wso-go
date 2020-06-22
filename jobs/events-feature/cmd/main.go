package main

import (
	"fmt"
	events "github.com/WilliamsStudentsOnline/wso-go/jobs/events-feature"
)

func main() {
	announcements, _ := events.GetRawDailyMessages()
	fmt.Printf("%+v\n", announcements)

}
