package main

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"gopkg.in/yaml.v3"
)

type EphmatchEra struct {
	Start      string `yaml:"start"`
	End        string `yaml:"end"`
	SeniorOnly bool   `yaml:"senior_only,omitempty"` // omitempty prevents 'senior_only: false'
}

func main() {
	filePath := flag.String("file", "", "the configuration file for wso-backend (required)")
	flag.Parse()

	cfg, err := config.LoadConfig(*filePath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	log, err := logging.SetupLog(cfg, "ephmatch-set-dates")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
	}
	defer log.Sync()

	absFilePath, err := filepath.Abs(*filePath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for %s: %v", *filePath, err)
	}

	dataRawBytes, err := os.ReadFile(absFilePath)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", absFilePath, err)
	}
	dataRaw := string(dataRawBytes)
	log.Infof("Successfully parsed %s", absFilePath)

	// GENERATE ERAS
	// THESE SHOULD BE MANUALLY MODIFIED PER-YEAR TO LINE UP WITH SCHEDULES, BUT EXIST SO THAT NOBODY FORGETS TO TURN IT ON
	// (AGAIN)
	year := time.Now().Year()
	dates := []EphmatchEra{
		{
			// winter study
			Start: time.Date(year, time.January, 5, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:   time.Date(year, time.February, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		{
			// valentine's
			Start: time.Date(year, time.February, 8, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:   time.Date(year, time.February, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		{
			// pre spring break (seniors only)
			Start:      time.Date(year, time.March, 7, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:        time.Date(year, time.March, 21, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			SeniorOnly: true,
		},
		{
			// post spring break for everyone
			Start: time.Date(year, time.April, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:   time.Date(year, time.April, 14, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		{
			// graduation season (seniors only)
			Start:      time.Date(year, time.May, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:        time.Date(year, time.July, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			SeniorOnly: true,
		},
		{
			// mountain day season
			Start: time.Date(year, time.October, 10, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:   time.Date(year, time.October, 31, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
		{
			// thanksgiving season (seniors only)
			Start:      time.Date(year, time.November, 14, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:        time.Date(year, time.November, 28, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			SeniorOnly: true,
		},
	}

	yamlBytes, err := yaml.Marshal(dates)
	if err != nil {
		log.Fatalf("Error marshalling dates to YAML: %v", err)
	}
	yamlString := string(yamlBytes)

	lines := strings.Split(strings.TrimSpace(yamlString), "\n")
	indentedLines := make([]string, len(lines))
	for i, line := range lines {
		indentedLines[i] = "  " + line
	}
	dataRawValue := strings.Join(indentedLines, "\n")

	// find all lines until the first with a non-whitespace character after our key
	pattern := `(?m)(^ephmatch_eras:\s*\n)(?:^\s+.*$\n?)*`
	re := regexp.MustCompile(pattern)

	if !re.MatchString(dataRaw) {
		log.Fatalf("Error: Could not find the 'ephmatch_eras:' block to replace in %s", absFilePath)
	}

	replacement := "ephmatch_eras:\n" + dataRawValue + "\n"
	updatedBlock := re.ReplaceAllString(dataRaw, replacement)

	err = os.WriteFile(absFilePath, []byte(updatedBlock), 0644) // let's hope that 0644 perms are fine for config
	if err != nil {
		log.Fatalf("Error writing updated content to file %s: %v", absFilePath, err)
	}

	log.Infof("Update complete!")
}
