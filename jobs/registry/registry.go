// Package registry describes batch jobs for the systemd-based jobs system.
package registry

import (
	"fmt"
	"sort"
	"strings"
)

// InstallDir is where CD installs job binaries on the prod host.
const InstallDir = "/opt/wso-go"

// ParamType is the kind of CLI flag a job accepts.
type ParamType string

const (
	ParamString ParamType = "string"
	ParamBool   ParamType = "bool"
	ParamEnum   ParamType = "enum"
	ParamFile   ParamType = "file"
)

// Param is a typed CLI flag the admin console (or jobctl) may supply.
type Param struct {
	Name        string    // flag name without leading dashes, e.g. "file"
	Type        ParamType // string, bool, enum, or file
	Required    bool
	Default     string   // used when the value is omitted; empty means no default
	Description string
	EnumValues  []string // required when Type is ParamEnum
}

// Schedule is a systemd OnCalendar expression for recurring runs.
// Nil Schedule means the job is manual / one-off only.
type Schedule struct {
	OnCalendar string // e.g. "*-*-* 01:00:00"
}

// Job is one batch job the host can run.
type Job struct {
	Name          string
	Description   string
	Binary        string // absolute path to the binary on the host
	Params        []Param
	Schedule      *Schedule // nil = no timer
	RuntimeMaxSec int       // 0 = systemd default
	MemoryMax     string    // e.g. "512M"; empty = systemd default
	User          string    // empty defaults to "wso"
	Group         string    // empty defaults to "wso"
}

// UnitName is the systemd unit basename for this job (without .service/.timer).
func (j Job) UnitName() string {
	return "wso-job-" + j.Name
}

// ServiceFile is the systemd service unit filename.
func (j Job) ServiceFile() string {
	return j.UnitName() + ".service"
}

// TimerFile is the systemd timer unit filename.
func (j Job) TimerFile() string {
	return j.UnitName() + ".timer"
}

// EffectiveUser returns the systemd User=, defaulting to wso.
func (j Job) EffectiveUser() string {
	if j.User == "" {
		return "wso"
	}
	return j.User
}

// EffectiveGroup returns the systemd Group=, defaulting to wso.
func (j Job) EffectiveGroup() string {
	if j.Group == "" {
		return "wso"
	}
	return j.Group
}

// ParamByName returns the param with the given name, if any.
func (j Job) ParamByName(name string) (Param, bool) {
	for _, p := range j.Params {
		if p.Name == name {
			return p, true
		}
	}
	return Param{}, false
}

// BuildArgs turns a map of param name -> value into argv for the job binary.
// Missing optional params use Default when set. Unknown names and invalid
// values are errors. Bool values accept "true"/"false" (case-insensitive).
func (j Job) BuildArgs(values map[string]string) ([]string, error) {
	if values == nil {
		values = map[string]string{}
	}

	for name := range values {
		if _, ok := j.ParamByName(name); !ok {
			return nil, fmt.Errorf("unknown param %q", name)
		}
	}

	var args []string
	for _, p := range j.Params {
		raw, ok := values[p.Name]
		if !ok || raw == "" {
			if p.Required && p.Default == "" {
				return nil, fmt.Errorf("param %q is required", p.Name)
			}
			if p.Default == "" {
				continue
			}
			raw = p.Default
		}

		flag := "-" + p.Name
		switch p.Type {
		case ParamBool:
			v, err := parseBool(raw)
			if err != nil {
				return nil, fmt.Errorf("param %q: %w", p.Name, err)
			}
			if v {
				args = append(args, flag)
			}
		case ParamEnum:
			if !contains(p.EnumValues, raw) {
				return nil, fmt.Errorf("param %q: %q is not one of %s", p.Name, raw, strings.Join(p.EnumValues, ", "))
			}
			args = append(args, flag, raw)
		case ParamString, ParamFile:
			args = append(args, flag, raw)
		default:
			return nil, fmt.Errorf("param %q: unknown type %q", p.Name, p.Type)
		}
	}
	return args, nil
}

// Validate checks that a Job is well-formed before registration.
func (j Job) Validate() error {
	if j.Name == "" {
		return fmt.Errorf("job name is required")
	}
	if strings.ContainsAny(j.Name, " \t\n/") || strings.Contains(j.Name, "..") {
		return fmt.Errorf("job name %q is invalid", j.Name)
	}
	if j.Binary == "" {
		return fmt.Errorf("job %s: binary is required", j.Name)
	}
	if j.Schedule != nil && j.Schedule.OnCalendar == "" {
		return fmt.Errorf("job %s: schedule OnCalendar is empty", j.Name)
	}

	seen := map[string]struct{}{}
	for _, p := range j.Params {
		if p.Name == "" {
			return fmt.Errorf("job %s: param name is required", j.Name)
		}
		if _, ok := seen[p.Name]; ok {
			return fmt.Errorf("job %s: duplicate param %q", j.Name, p.Name)
		}
		seen[p.Name] = struct{}{}

		switch p.Type {
		case ParamString, ParamBool, ParamFile:
		case ParamEnum:
			if len(p.EnumValues) == 0 {
				return fmt.Errorf("job %s: param %q: enum requires EnumValues", j.Name, p.Name)
			}
		default:
			return fmt.Errorf("job %s: param %q: unknown type %q", j.Name, p.Name, p.Type)
		}
	}
	return nil
}

// Registry is the set of known jobs.
type Registry struct {
	jobs map[string]Job
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{jobs: map[string]Job{}}
}

// Register adds a job. Names must be unique.
func (r *Registry) Register(j Job) error {
	if err := j.Validate(); err != nil {
		return err
	}
	if _, exists := r.jobs[j.Name]; exists {
		return fmt.Errorf("job %q already registered", j.Name)
	}
	r.jobs[j.Name] = j
	return nil
}

// Get returns a job by name.
func (r *Registry) Get(name string) (Job, bool) {
	j, ok := r.jobs[name]
	return j, ok
}

// All returns all jobs sorted by name.
func (r *Registry) All() []Job {
	out := make([]Job, 0, len(r.jobs))
	for _, j := range r.jobs {
		out = append(out, j)
	}
	sort.Slice(out, func(i, k int) bool { return out[i].Name < out[k].Name })
	return out
}

// Scheduled returns jobs that have a recurring schedule, sorted by name.
func (r *Registry) Scheduled() []Job {
	var out []Job
	for _, j := range r.jobs {
		if j.Schedule != nil {
			out = append(out, j)
		}
	}
	sort.Slice(out, func(i, k int) bool { return out[i].Name < out[k].Name })
	return out
}

// Default is the process-wide registry. Later layers register jobs into it.
func Default() *Registry {
	return defaultRegistry
}

var defaultRegistry = New()

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", s)
	}
}

func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
