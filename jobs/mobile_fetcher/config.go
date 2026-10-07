package mobile_fetcher

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Interval  time.Duration `yaml:"-"`
	IntervalS string        `yaml:"interval"`
	OutputDir string        `yaml:"output_dir"`
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
