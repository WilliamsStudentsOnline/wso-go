package main

import (
	"fmt"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/jobctl"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/registry"
	"github.com/jinzhu/gorm"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func main() {
	app := cli.NewApp()
	app.Name = "wso-jobctl"
	app.Usage = "manage WSO systemd job units and run history"
	app.HideVersion = true
	app.Commands = []cli.Command{
		{
			Name:  "gen-units",
			Usage: "write systemd service/timer units from the job registry",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "out",
					Value: "/etc/systemd/system",
					Usage: "directory to write unit files into",
				},
			},
			Action: genUnits,
		},
		{
			Name:  "record-start",
			Usage: "mark a job run as started",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Value: jobctl.DefaultConfigPath,
					Usage: "path to config.yaml",
				},
			},
			Action: recordStart,
		},
		{
			Name:  "record-finish",
			Usage: "mark a job run as finished and copy logs",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Value: jobctl.DefaultConfigPath,
					Usage: "path to config.yaml",
				},
			},
			Action: recordFinish,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err.Error())
		os.Exit(1)
	}
}

func genUnits(c *cli.Context) error {
	out := c.String("out")
	if err := jobctl.GenUnits(registry.Default(), out); err != nil {
		return err
	}
	fmt.Printf("wrote units to %s (%d jobs)\n", out, len(registry.Default().All()))
	return nil
}

func withDB(configPath string, fn func(*gorm.DB) error) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	log := zap.NewNop().Sugar()
	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)
	return fn(db)
}

func recordStart(c *cli.Context) error {
	return withDB(c.String("config"), jobctl.RecordStart)
}

func recordFinish(c *cli.Context) error {
	return withDB(c.String("config"), jobctl.RecordFinish)
}
