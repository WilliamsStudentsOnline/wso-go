package mobile_fetcher

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Interval  time.Duration `yaml:"-"`
	IntervalS string        `yaml:"interval"`
	OutputDir string        `yaml:"output_dir"`
	Owner     string        `yaml:"owner"`
	Group     string        `yaml:"group"`
	Jobs      []Job         `yaml:"jobs"`
}

type Job struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Output  string            `yaml:"output"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type Overrides struct {
	OutputDir string
	Interval  time.Duration
}

// -1 = unchanged
type FileOwnership struct {
	UID int
	GID int
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := cfg.finalize(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) finalize() error {
	if c.IntervalS == "" {
		return fmt.Errorf("config: interval is required")
	}
	d, err := time.ParseDuration(c.IntervalS)
	if err != nil {
		return fmt.Errorf("config: invalid interval %q: %w", c.IntervalS, err)
	}
	if d <= 0 {
		return fmt.Errorf("config: interval must be positive")
	}
	c.Interval = d

	if c.OutputDir == "" {
		return fmt.Errorf("config: output_dir is required")
	}
	if len(c.Jobs) == 0 {
		return fmt.Errorf("config: at least one job is required")
	}

	for i := range c.Jobs {
		job := &c.Jobs[i]
		if job.Name == "" {
			return fmt.Errorf("config: jobs[%d]: name is required", i)
		}
		if job.URL == "" {
			return fmt.Errorf("config: job %q: url is required", job.Name)
		}
		if job.Output == "" {
			return fmt.Errorf("config: job %q: output is required", job.Name)
		}
		for k, v := range job.Headers {
			job.Headers[k] = os.ExpandEnv(v)
		}
	}
	return nil
}

func (c *Config) ApplyOverrides(o Overrides) {
	if o.OutputDir != "" {
		c.OutputDir = o.OutputDir
	}
	if o.Interval > 0 {
		c.Interval = o.Interval
		c.IntervalS = o.Interval.String()
	}
}

// owner/group -> ID; omitted = -1
func (c *Config) ResolveOwnership() (FileOwnership, error) {
	own := FileOwnership{UID: -1, GID: -1}
	if c.Owner != "" {
		uid, err := lookupUID(c.Owner)
		if err != nil {
			return own, fmt.Errorf("config: owner %q: %w", c.Owner, err)
		}
		own.UID = uid
	}
	if c.Group != "" {
		gid, err := lookupGID(c.Group)
		if err != nil {
			return own, fmt.Errorf("config: group %q: %w", c.Group, err)
		}
		own.GID = gid
	}
	return own, nil
}

func (o FileOwnership) set() bool {
	return o.UID >= 0 || o.GID >= 0
}

func lookupUID(name string) (int, error) {
	u, err := user.Lookup(name)
	if err != nil {
		return -1, err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, fmt.Errorf("parse uid %q: %w", u.Uid, err)
	}
	return uid, nil
}

func lookupGID(name string) (int, error) {
	g, err := user.LookupGroup(name)
	if err != nil {
		return -1, err
	}
	gid, err := strconv.Atoi(g.Gid)
	if err != nil {
		return -1, fmt.Errorf("parse gid %q: %w", g.Gid, err)
	}
	return gid, nil
}
