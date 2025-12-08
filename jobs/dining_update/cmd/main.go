package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update"
)

func main() {
	var outDir string
	var todayLink string
	var vendorInfoPath string
	var force bool
	var dateStr string

	flag.StringVar(&outDir, "outdir", "./menus", "directory to write YYYY-MM-DD.json menu files")
	flag.StringVar(&todayLink, "today", "", "path to symlink for today's menu (optional)")
	flag.StringVar(&vendorInfoPath, "vi", "vendor_info.yaml", "path to vendor info YAML file")
	flag.BoolVar(&force, "force", false, "force refresh even if today's file already exists")
	flag.StringVar(&dateStr, "date", "", "fetch week containing YYYY/MM/DD (default: today)")

	flag.Parse()

	// parse date
	var fetchDate time.Time
	if dateStr == "" {
		fetchDate = time.Now()
	} else {
		var err error
		fetchDate, err = time.Parse("2006/01/02", dateStr)
		if err != nil {
			log.Fatalf("Invalid -date. Expected YYYY/MM/DD. Got: %s", dateStr)
		}
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("Unable to create output directory: %v", err)
	}

	// find todays date for symlink update
	today := time.Now().Format("2006-01-02")
	todayFile := filepath.Join(outDir, today+".json")

	refresh := force
	if _, err := os.Stat(todayFile); err != nil {
		refresh = true
	}

	if refresh {
		if err := dining_update.UpdateDining(outDir, vendorInfoPath, fetchDate); err != nil {
			log.Fatalf("Error updating dining menus: %v", err)
		}
	}

	// update symlink so dining.json is accurate
	if todayLink != "" {
		_ = os.Remove(todayLink)
		if err := os.Symlink(todayFile, todayLink); err != nil {
			log.Fatalf("Failed to update symlink: %v", err)
		}
	}

	log.Println("Done.")
}
