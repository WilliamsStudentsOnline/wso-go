package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/mobile_fetcher"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config")
	outputDir := flag.String("output-dir", "", "override output directory")
	interval := flag.Duration("interval", 0, "override polling interval")
	once := flag.Bool("once", false, "fetch once and exit")
	flag.Parse()

	cfg, err := mobile_fetcher.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.ApplyOverrides(mobile_fetcher.Overrides{
		OutputDir: *outputDir,
		Interval:  *interval,
	})

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	fetcher := mobile_fetcher.NewFetcher(cfg.OutputDir, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("mobile-fetcher starting: %d jobs, interval=%s, output=%s",
		len(cfg.Jobs), cfg.Interval, cfg.OutputDir)

	fetcher.FetchAll(cfg.Jobs)
	if *once {
		log.Println("once mode complete")
		return
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			return
		case <-ticker.C:
			fetcher.FetchAll(cfg.Jobs)
		}
	}
}
