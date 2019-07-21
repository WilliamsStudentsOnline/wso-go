package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Our configuration
type Config struct {
	Env          string   `yaml:"env"`
	DatabaseType string   `yaml:"database_type"`
	DatabaseArgs string   `yaml:"database_args"`
	JWTRealm     string   `yaml:"jwt_realm"`
	GinMode      string   `yaml:"gin_mode"`
	Port         string   `yaml:"port"`
	DisableLDAP  bool     `yaml:"disable_ldap"`
	Secrets      *Secrets `yaml:"-"`
}

// Check what environment our config is in
func (c *Config) IsEnv(env string) bool {
	return c.Env == env
}

func (c *Config) IsDevelopment() bool {
	return c.IsEnv("development")
}

func (c *Config) IsTest() bool {
	return c.IsEnv("test")
}

func (c *Config) IsProduction() bool {
	return c.IsEnv("production")
}

// Get the config and parse any info
func GetConfig(env string, configPath string) (*Config, error) {
	// Default config path
	if configPath == "" {
		configPath = filepath.Join("config", "environment", env+".yml")
	}

	// Get the config file and decode it
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	dec := yaml.NewDecoder(configFile)
	err = dec.Decode(cfg)
	if err != nil {
		return nil, err
	}

	// Config Defaults
	if cfg.GinMode == "" {
		switch cfg.Env {
		case "development":
			cfg.GinMode = "development"
		case "test":
			cfg.GinMode = "test"
		case "production":
			cfg.GinMode = "production"
		default:
			cfg.GinMode = "development"
		}
	}

	return cfg, nil
}
