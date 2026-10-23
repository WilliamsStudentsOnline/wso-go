package main

import (
	"fmt"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/jobctl"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/registry"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "wso-jobctl"
	app.Usage = "manage WSO systemd job units"
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
