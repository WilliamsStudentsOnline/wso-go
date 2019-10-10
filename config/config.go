package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Our configuration
type Config struct {
	// We can set our config file via environment variable, but not physical file
	ConfigFile string `envconfig:"config_file" yaml:"-"`
	// We can set our secrets file via env var or physical file
	SecretsFile string `envconfig:"secrets_file" yaml:"secrets_file"`

	// Environment of server
	Env string `yaml:"env" envconfig:"env"`

	/* Database */
	DatabaseType string `yaml:"database_type" envconfig:"database_type"`
	// This overrides any specific arguments passed for MySQL or SQLite
	DatabaseArgs string `yaml:"database_args" envconfig:"database_args"`
	// MySQL (some are in the secrets yml)
	MySQLUser     string            `yaml:"mysql_user" envconfig:"mysql_user"`
	MySQLHost     string            `yaml:"mysql_host" envconfig:"mysql_host"`
	MySQLPort     int               `yaml:"mysql_port" envconfig:"mysql_port"`
	MySQLDatabase string            `yaml:"mysql_database" envconfig:"mysql_database"`
	MySQLArgs     map[string]string `yaml:"mysql_args" envconfig:"mysql_args"`
	MySQLUnix     bool              `yaml:"mysql_unix" envconfig:"mysql_unix"` // Use a unix connection rather than a TCP connection
	// SQLite
	SQLiteFile string `yaml:"sqlite_file" envconfig:"sqlite_file"`

	SlackWebhookURL string `yaml:"slack_webhook_url" envconfig:"slack_webhook_url"`

	/* JWT */
	JWTRealm string `yaml:"jwt_realm" envconfig:"jwt_realm"`
	// Timeout in hours
	JWTTimeoutHours int `yaml:"jwt_timeout_hours" envconfig:"jwt_timeout_hours"`
	// Asymmetric algorithm setup
	JWTPublicKeyFile  string `yaml:"jwt_public_key_file" envconfig:"jwt_public_key_file"`
	JWTPrivateKeyFile string `yaml:"jwt_private_key_file" envconfig:"jwt_private_key_file"`
	JWTUseAsymmetric  bool   `yaml:"jwt_use_asymmetric" envconfig:"jwt_use_asymmetric"`

	/* Gin */
	GinMode string `yaml:"gin_mode" envconfig:"gin_mode"`
	Port    int    `yaml:"port" envconfig:"port"`
	// Enable TLS (HTTPS):
	EnableTLS   bool   `yaml:"enable_tls" envconfig:"enable_tls"`
	TLSCertPath string `yaml:"tls_cert_path" envconfig:"tls_cert_path"`
	TLSKeyPath  string `yaml:"tls_key_path" envconfig:"tls_key_path"`

	/* Server */
	EnableAPIDocs bool   `yaml:"enable_api_docs" envconfig:"enable_api_docs"`
	DisableLDAP   bool   `yaml:"disable_ldap" envconfig:"disable_ldap"`
	LogLevel      string `yaml:"log_level" envconfig:"log_level"`
	Hostname      string `yaml:"hostname" envconfig:"hostname"`

	/* Search */
	SearchBackend string `yaml:"search_backend" envconfig:"search_backend"`

	/* Kubernetes */
	KubernetesEnabled   bool   `yaml:"kubernetes_enabled" envconfig:"kubernetes_enabled"`
	KubeNamespace       string `yaml:"kube_namespace" envconfig:"kube_namespace"`
	KubeJobImageVersion string `yaml:"kube_job_image_version" envconfig:"kube_job_image_version"`

	Secrets        *Secrets     `yaml:"-" envconfig:"-"`
	LogLevelParsed logrus.Level `yaml:"-" envconfig:"-"`
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

// Get the config and parse any info. Environment variables will look like WSO_PUT_CONFIG_NAME_HERE
func LoadConfig(configPath string) (*Config, error) {
	// We first get env variables for config. These override everything other form of config.
	cfg := &Config{}
	err := envconfig.Process("wso", cfg)
	if err != nil {
		return nil, err
	}

	// If config file is not an env variable but is a cmd flag, set it to that
	if cfg.ConfigFile == "" && configPath != "" {
		cfg.ConfigFile = configPath
	}

	// If we have it, load config from disk
	if cfg.ConfigFile != "" {
		cfg.ConfigFile, err = filepath.Abs(cfg.ConfigFile)
		if err != nil {
			return nil, err
		}

		configFile, err := os.Open(cfg.ConfigFile)
		if err != nil {
			return nil, err
		}

		yamlCfg := Config{}
		dec := yaml.NewDecoder(configFile)
		err = dec.Decode(&yamlCfg)
		if err != nil {
			return nil, err
		}

		// Set any empty values from cfg that are non-empty in yamlCfg to the ones in yamlCfg
		err = mergo.Merge(cfg, yamlCfg)
		if err != nil {
			return nil, err
		}
	}

	// Load secrets by env and file
	cfg.Secrets, err = LoadSecrets(cfg.SecretsFile)
	if err != nil {
		return nil, err
	}

	// Do config/secrets defaults/setup
	err = SetupConfig(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Setup config. This runs defaults, requirements, and other configuration stuff.
func SetupConfig(c *Config) error {
	// Get default gin mode if we just have env
	if c.GinMode == "" {
		switch c.Env {
		case "development":
			c.GinMode = gin.DebugMode
		case "test":
			c.GinMode = gin.TestMode
		case "production":
			c.GinMode = gin.ReleaseMode
		default:
			c.GinMode = gin.DebugMode
		}
	}

	if c.JWTTimeoutHours == 0 {
		c.JWTTimeoutHours = 1
	}

	if (c.JWTPublicKeyFile == "" || c.JWTPrivateKeyFile == "") && c.Secrets.JWTSecretKey == "" {
		return errors.New("missing JWT key signature: either need pub key & priv key, or secret key")
	}

	if c.JWTUseAsymmetric && (c.JWTPrivateKeyFile == "" || c.JWTPublicKeyFile == "") {
		return errors.New("missing JWT priv or pub key for asymmetric algorithm")
	}

	// Set log level if empty
	if c.LogLevel == "" {
		if c.IsProduction() {
			c.LogLevel = "warn"
		} else if c.IsDevelopment() {
			c.LogLevel = "debug"
		} else if c.IsTest() {
			c.LogLevel = "trace"
		} else {
			c.LogLevel = "info"
		}
	}

	// Set log level parsed
	switch c.LogLevel {
	case "panic":
		c.LogLevelParsed = logrus.PanicLevel
	case "fatal":
		c.LogLevelParsed = logrus.FatalLevel
	case "error":
		c.LogLevelParsed = logrus.ErrorLevel
	case "warn":
		c.LogLevelParsed = logrus.WarnLevel
	case "info":
		c.LogLevelParsed = logrus.InfoLevel
	case "debug":
		c.LogLevelParsed = logrus.DebugLevel
	case "trace":
		c.LogLevelParsed = logrus.TraceLevel
	}

	// Setup database info
	switch c.DatabaseType {
	case "mysql":
		SetupMySQLConfig(c)
	case "sqlite3":
		SetupSQLiteConfig(c)
	case "sqlite":
		c.DatabaseType = "sqlite3"
		SetupSQLiteConfig(c)
	}

	// Default to latest
	if c.KubeJobImageVersion == "" {
		c.KubeJobImageVersion = "latest"
	}

	// Default to SearchBackendSQL
	if c.SearchBackend == "" {
		c.SearchBackend = search.SearchBackendSQL
	}

	if c.SearchBackend != search.SearchBackendSQL {
		return errors.New("unknown search backend")
	}

	// Default to port 8080
	if c.Port == 0 {
		c.Port = 8080
	}

	if c.EnableTLS && (c.TLSKeyPath == "" || c.TLSCertPath == "") {
		return errors.New("missing TLS cert/key path with TLS enabled")
	}

	if c.Hostname == "" {
		c.Hostname = "localhost"
	}

	return nil
}

func setDefaultStr(str string, defaultValue string) string {
	if str == "" {
		return defaultValue
	}
	return str
}
