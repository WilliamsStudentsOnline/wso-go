package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
			Flags: configFlag(),
			Action: recordStart,
		},
		{
			Name:  "record-finish",
			Usage: "mark a job run as finished and copy logs",
			Flags: configFlag(),
			Action: recordFinish,
		},
		{
			Name:      "run",
			Usage:     "enqueue a job run, starting now if idle else queueing",
			ArgsUsage: "<job-name>",
			Flags: append(configFlag(),
				cli.StringSliceFlag{
					Name:  "set",
					Usage: "param=value, repeatable",
				},
				cli.StringFlag{
					Name:  "run-at",
					Usage: "optional RFC3339 run time; queues until then",
				},
			),
			Action: runJob,
		},
		{
			Name:  "dispatch",
			Usage: "start due pending runs that are not blocked by an active run",
			Flags: configFlag(),
			Action: dispatch,
		},
		{
			Name:      "cancel",
			Usage:     "cancel a pending run or stop a running one",
			ArgsUsage: "<run-id>",
			Flags:     configFlag(),
			Action:    cancelRun,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err.Error())
		os.Exit(1)
	}
}

func configFlag() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: jobctl.DefaultConfigPath,
			Usage: "path to config.yaml",
		},
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

func runJob(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("job name required")
	}
	values, err := parseSetFlags(c.StringSlice("set"))
	if err != nil {
		return err
	}
	var runAt *time.Time
	if s := c.String("run-at"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fmt.Errorf("run-at: %w", err)
		}
		t = t.UTC()
		runAt = &t
	}

	return withDB(c.String("config"), func(db *gorm.DB) error {
		sd, err := jobctl.NewSystemd()
		if err != nil {
			return err
		}
		defer sd.Close()
		run, err := jobctl.Enqueue(db, sd, registry.Default(), jobctl.EnqueueOpts{
			JobName: name,
			Values:  values,
			RunAt:   runAt,
		})
		if err != nil {
			return err
		}
		fmt.Printf("run %d status=%s job=%s\n", run.ID, run.Status, run.JobName)
		return nil
	})
}

func dispatch(c *cli.Context) error {
	return withDB(c.String("config"), func(db *gorm.DB) error {
		sd, err := jobctl.NewSystemd()
		if err != nil {
			return err
		}
		defer sd.Close()
		n, err := jobctl.DispatchDue(db, sd, registry.Default())
		if err != nil {
			return err
		}
		fmt.Printf("started %d run(s)\n", n)
		return nil
	})
}

func cancelRun(c *cli.Context) error {
	idStr := c.Args().First()
	if idStr == "" {
		return fmt.Errorf("run id required")
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("run id: %w", err)
	}
	return withDB(c.String("config"), func(db *gorm.DB) error {
		sd, err := jobctl.NewSystemd()
		if err != nil {
			return err
		}
		defer sd.Close()
		run, err := jobctl.Cancel(db, sd, uint(id))
		if err != nil {
			return err
		}
		fmt.Printf("run %d status=%s\n", run.ID, run.Status)
		return nil
	})
}

func parseSetFlags(sets []string) (map[string]string, error) {
	out := map[string]string{}
	for _, s := range sets {
		k, v, ok := strings.Cut(s, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("invalid --set %q (want name=value)", s)
		}
		out[k] = v
	}
	return out, nil
}
