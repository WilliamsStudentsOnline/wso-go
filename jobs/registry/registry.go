// Package registry is the catalog of batch jobs for wso-jobctl
package registry

import (
	"fmt"
	"sort"
	"strings"
)

const InstallDir = "/opt/wso-go"

type ParamType string

const (
	ParamString ParamType = "string"
	ParamBool   ParamType = "bool"
	ParamEnum   ParamType = "enum"
	ParamFile   ParamType = "file"
)

type Param struct {
	Name        string
	Type        ParamType
	Required    bool
	Default     string
	Description string
	EnumValues  []string
}

type Schedule struct {
	OnCalendar string
}

type Job struct {
	Name          string
	Description   string
	Binary        string
	Params        []Param
	Schedule      *Schedule
	RuntimeMaxSec int
	MemoryMax     string
	User          string
	Group         string
}

func (j Job) UnitName() string {
	return "wso-job-" + j.Name
}

func (j Job) ServiceFile() string {
	return j.UnitName() + ".service"
}

func (j Job) TimerFile() string {
	return j.UnitName() + ".timer"
}

func (j Job) EffectiveUser() string {
	if j.User == "" {
		return "wso"
	}
	return j.User
}

func (j Job) EffectiveGroup() string {
	if j.Group == "" {
		return "wso"
	}
	return j.Group
}

func (j Job) ParamByName(name string) (Param, bool) {
	for _, p := range j.Params {
		if p.Name == name {
			return p, true
		}
	}
	return Param{}, false
}

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

type Registry struct {
	jobs map[string]Job
}

func New() *Registry {
	return &Registry{jobs: map[string]Job{}}
}

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

func (r *Registry) Get(name string) (Job, bool) {
	j, ok := r.jobs[name]
	return j, ok
}

func (r *Registry) All() []Job {
	out := make([]Job, 0, len(r.jobs))
	for _, j := range r.jobs {
		out = append(out, j)
	}
	sort.Slice(out, func(i, k int) bool { return out[i].Name < out[k].Name })
	return out
}

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
