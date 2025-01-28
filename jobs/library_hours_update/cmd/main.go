package main

import (
	"flag"
	"log"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/library_hours_update"
)

func main() {
	var out string

	flag.StringVar(&out, "out", "library_hours.json", "path to output")

	flag.Parse()

	err := library_hours_update.UpdateLibraryHours(out)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Done.")
}
