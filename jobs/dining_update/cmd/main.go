package main

import (
	"flag"
	"log"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update"
)

func main() {
	var out string
	var vendorInfoPath string
	var eats4Ephs bool

	flag.StringVar(&out, "out", "dining.json", "path to output")
	flag.StringVar(&vendorInfoPath, "vi", "vendor_info.yaml", "path to vendor info yaml file")
	flag.BoolVar(&eats4Ephs, "eats4Ephs", false, "use Eats4Ephs API instead of NetNutrition API")

	flag.Parse()

	err := dining_update.UpdateDining(out, eats4Ephs, vendorInfoPath)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Done.")
}
