package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/daily_messages_update"
)

func main() {
	var outDir string
	var todayLink string
	var url string

	flag.StringVar(&outDir, "outdir", "./daily_messages", "directory to write YYYY-MM-DD.json files")
	flag.StringVar(&todayLink, "today", "", "path to symlink for today's messages (optional)")
	flag.StringVar(&url, "url", daily_messages_update.FeedURL, "daily messages feed URL")

	flag.Parse()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("Unable to create output directory: %v", err)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	if err := daily_messages_update.Update(client, url, outDir, todayLink, time.Now()); err != nil {
		log.Fatalf("Error updating daily messages: %v", err)
	}

	log.Println("Done.")
}
