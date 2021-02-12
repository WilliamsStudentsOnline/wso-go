package main

import (
	"flag"
	"log"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update"
)

func main() {
	var out string
	var vendorInfoPath string

	flag.StringVar(&out, "out", "dining.json", "path to output")
	flag.StringVar(&vendorInfoPath, "vi", "vendor_info.yaml", "path to vendor info yaml file")

	flag.Parse()

	err := dining_update.UpdateDining(out, vendorInfoPath)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Done.")
}
