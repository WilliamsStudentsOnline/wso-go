package jobctl

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/registry"
	"github.com/coreos/go-systemd/v22/dbus"
	godbus "github.com/godbus/dbus/v5"
)

type Systemd struct {
	conn *dbus.Conn
}

func NewSystemd() (*Systemd, error) {
	conn, err := dbus.NewSystemConnection()
	if err != nil {
		return nil, err
	}
	return &Systemd{conn: conn}, nil
}

func (s *Systemd) Close() {
	if s != nil && s.conn != nil {
		s.conn.Close()
	}
}

func (s *Systemd) StartJobRun(j registry.Job, runID uint, args []string) (unitName string, err error) {
	if err := j.Validate(); err != nil {
		return "", err
	}
	unitName = TransientUnitName(j.Name, runID)
	if !ValidTransientUnitName(unitName, j.Name) {
		return "", fmt.Errorf("refusing unit name %q", unitName)
	}

	exec := append([]string{j.Binary}, args...)
	recordStart := []string{JobctlBinary, "record-start", "-config", DefaultConfigPath}
	recordFinish := []string{JobctlBinary, "record-finish", "-config", DefaultConfigPath}
	props := []dbus.Property{
		dbus.PropDescription(fmt.Sprintf("WSO job run %d (%s)", runID, j.Name)),
		dbus.PropType("oneshot"),
		dbus.PropExecStart(exec, false),
		propExecCmds("ExecStartPre", recordStart, true),
		propExecCmds("ExecStopPost", recordFinish, false),
		{Name: "User", Value: godbus.MakeVariant(j.EffectiveUser())},
		{Name: "Group", Value: godbus.MakeVariant(j.EffectiveGroup())},
		{Name: "Environment", Value: godbus.MakeVariant([]string{
			"WSO_JOB_NAME=" + j.Name,
			"WSO_JOB_RUN_ID=" + strconv.FormatUint(uint64(runID), 10),
			"WSO_JOB_UNIT=" + unitName,
		})},
	}
	if after := append([]string{"network-online.target"}, j.After...); len(after) > 0 {
		props = append(props, dbus.PropAfter(after...))
	}
	if len(j.Requires) > 0 {
		props = append(props, dbus.PropRequires(j.Requires...))
	}
	if len(j.Wants) > 0 {
		props = append(props, dbus.PropWants(j.Wants...))
	}
	if j.RuntimeMaxSec > 0 {
		props = append(props, dbus.Property{
			Name:  "RuntimeMaxUSec",
			Value: godbus.MakeVariant(uint64(j.RuntimeMaxSec) * 1_000_000),
		})
	}
	if j.MemoryMax != "" {
		props = append(props, dbus.Property{
			Name:  "MemoryMax",
			Value: godbus.MakeVariant(j.MemoryMax),
		})
	}

	reschan := make(chan string, 1)
	_, err = s.conn.StartTransientUnit(unitName, "replace", props, reschan)
	if err != nil {
		return "", err
	}
	select {
	case <-reschan:
		return unitName, nil
	case <-time.After(30 * time.Second):
		return unitName, fmt.Errorf("timeout waiting for StartTransientUnit %s", unitName)
	}
}

func (s *Systemd) StopUnit(unitName string) error {
	if !strings.HasPrefix(unitName, "wso-job-") || !strings.HasSuffix(unitName, ".service") {
		return fmt.Errorf("refusing to stop unit %q", unitName)
	}
	reschan := make(chan string, 1)
	_, err := s.conn.StopUnit(unitName, "replace", reschan)
	if err != nil {
		return err
	}
	select {
	case <-reschan:
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout stopping %s", unitName)
	}
}

func (s *Systemd) TimerNextElapse(jobName string) (*time.Time, error) {
	unit := "wso-job-" + jobName + ".timer"
	props, err := s.conn.GetUnitTypeProperties(unit, "Timer")
	if err != nil {
		return nil, err
	}
	raw, ok := props["NextElapseUSecRealtime"]
	if !ok {
		return nil, nil
	}
	usec, ok := raw.(uint64)
	if !ok || usec == 0 || usec == ^uint64(0) {
		return nil, nil
	}
	t := time.Unix(0, int64(usec)*1000).UTC()
	return &t, nil
}

func TransientUnitName(jobName string, runID uint) string {
	return fmt.Sprintf("wso-job-%s-r%d.service", jobName, runID)
}

func ValidTransientUnitName(unitName, jobName string) bool {
	prefix := "wso-job-" + jobName + "-r"
	return strings.HasPrefix(unitName, prefix) && strings.HasSuffix(unitName, ".service")
}

type execCmd struct {
	Path             string
	Args             []string
	UncleanIsFailure bool
}

func propExecCmds(name string, command []string, uncleanIsFailure bool) dbus.Property {
	return dbus.Property{
		Name: name,
		Value: godbus.MakeVariant([]execCmd{{
			Path:             command[0],
			Args:             command,
			UncleanIsFailure: uncleanIsFailure,
		}}),
	}
}
