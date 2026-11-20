package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/williams_record_update"
)

func main() {
	var out string
	var feedURL string

	flag.StringVar(&out, "out", "williams_record.json", "path to output")
	flag.StringVar(&feedURL, "feed", williams_record_update.DefaultFeedURL, "RSS feed URL")

	flag.Parse()

	client := &http.Client{Timeout: 30 * time.Second}
	n, err := williams_record_update.Update(client, feedURL, out)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Wrote %d articles to %s", n, out)
}
