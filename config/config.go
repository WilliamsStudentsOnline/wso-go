package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap/zapcore"
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
	// Timeout in hours (DEPRECATED: FOR OLD AUTH)
	JWTTimeoutHours int `yaml:"jwt_timeout_hours" envconfig:"jwt_timeout_hours"`
	// Auth 2.0 Timeouts: (for identity and API tokens)
	JWTIdentityTimeoutHours int `yaml:"jwt_identity_timeout_hours" envconfig:"jwt_identity_timeout_hours"`
	JWTAPITimeoutHours      int `yaml:"jwt_api_timeout_hours" envconfig:"jwt_api_timeout_hours"`
	// Asymmetric algorithm setup
	JWTPublicKeyFile  string `yaml:"jwt_public_key_file" envconfig:"jwt_public_key_file"`
	JWTPrivateKeyFile string `yaml:"jwt_private_key_file" envconfig:"jwt_private_key_file"`
	JWTSigningAlgo    string `yaml:"jwt_signing_algo" envconfig:"jwt_signing_algo"`

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
	//LogPath       string `yaml:"log_path" envconfig:"log_path"`
	LogDirectory string   `yaml:"log_directory" envconfig:"log_directory"`
	LogFormats   []string `yaml:"log_formats" envconfig:"log_formats"`

	/* Search */
	SearchBackend string `yaml:"search_backend" envconfig:"search_backend"`

	/* Ephmatch */
	// Periods/eras of time when ephmatch is open
	EphmatchEras []EphmatchEra `yaml:"ephmatch_eras" envconfig:"ephmatch_eras"`
	// Enable ephmatch indefinitely
	EphmatchEnableNow bool `yaml:"ephmatch_enable_now" envconfig:"ephmatch_enable_now"`

	/* Kubernetes */
	KubernetesEnabled   bool   `yaml:"kubernetes_enabled" envconfig:"kubernetes_enabled"`
	KubeNamespace       string `yaml:"kube_namespace" envconfig:"kube_namespace"`
	KubeJobImageVersion string `yaml:"kube_job_image_version" envconfig:"kube_job_image_version"`

	Secrets *Secrets `yaml:"-" envconfig:"-"`

	/* Pictures */
	PictureBackend   string `yaml:"picture_backend" envconfig:"picture_backend"`
	PictureLocalPath string `yaml:"picture_local_path" envconfig:"picture_local_path"`

	/* Chat */
	// The name of ejabberd service, like wso.williams.edu
	ChatEjabberdName string `yaml:"chat_ejabberd_name" envconfig:"chat_ejabberd_name"`

	/* Notifications */
	/* APNS */
	APNSAuthKey    string `yaml:"apns_auth_key" envconfig:"apns_auth_key"`
	APNSKeyID      string `yaml:"apns_key_id" envconfig:"apns_key_id"`
	APNSTeamID     string `yaml:"apns_team_id" envconfig:"apns_team_id"`
	APNSTopic      string `yaml:"apns_topic" envconfig:"apns_topic"`
	APNSProduction bool   `yaml:"apns_production" envconfig:"apns_production"`

	/* Dining (for read, not write) */
	DiningFile string `yaml:"dining_file" envconfig:"dining_file"`

	/* Goodrich */
	GoodrichManagerUnixes []string `yaml:"goodrich_manager_unixes" envconfig:"goodrich_manager_unixes"`
	// Use format: 2006-01-02
	GoodrichOpenDays []string `yaml:"goodrich_open_days" envconfig:"goodrich_open_days"`
	// Format: 15:04
	GoodrichOpen         string `yaml:"goodrich_open" envconfig:"goodrich_open"`
	GoodrichClose        string `yaml:"goodrich_close" envconfig:"goodrich_close"`
	GoodrichSlotSpotSize int    `yaml:"goodrich_slot_spot_size" envconfig:"goodrich_slot_spot_size"`
	GoodrichEmail        string `yaml:"goodrich_email" envconfig:"goodrich_email"`

	/* Email */
	EmailSMTPHost string `yaml:"email_smtp_host" envconfig:"email_smtp_host"`
	EmailSMTPPort int    `yaml:"email_smtp_port" envconfig:"email_smtp_port"`
}

type EphmatchEra struct {
	Start time.Time `yaml:"start"`
	End   time.Time `yaml:"end"`
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

func (c *Config) GoodrichMustParseTime(t string) (hour, minute int) {
	hour, minute, err := c.goodrichParseTime(t)
	if err != nil {
		panic(err)
	}
	return
}

func (c *Config) goodrichParseTime(t string) (hour, minute int, err error) {
	spl := strings.Split(t, ":")
	if len(spl) != 2 {
		return 0, 0, errors.New("could not parse time")
	}

	hour, err = strconv.Atoi(spl[0])
	if err != nil {
		return 0, 0, err
	}
	minute, err = strconv.Atoi(spl[1])
	if err != nil {
		return 0, 0, err
	}

	return
}

func (c *Config) MissingGoodrich() bool {
	return c.GoodrichOpen == "" ||
		c.GoodrichClose == "" ||
		c.GoodrichSlotSpotSize == 0 ||
		c.GoodrichEmail == ""
}

func (c *Config) MissingEmail() bool {
	return c.EmailSMTPHost == "" ||
		c.EmailSMTPPort == 0
}

func (c *Config) GenerateURL() *url.URL {
	scheme := "http"
	if c.EnableTLS {
		scheme = "https"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", c.Hostname, c.Port),
	}
}

func (c *Config) ParsedLogLevel() zapcore.Level {
	switch c.LogLevel {
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn":
		return zapcore.WarnLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
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

	if c.JWTAPITimeoutHours == 0 || c.JWTIdentityTimeoutHours == 0 {
		return errors.New("unknown JWT timeout for Auth 2.0")
	}

	if (c.JWTPublicKeyFile == "" || c.JWTPrivateKeyFile == "") && c.Secrets.JWTSecretKey == "" {
		return errors.New("missing JWT key signature: either need pub key & priv key, or secret key")
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

	if c.PictureBackend == "" {
		c.PictureBackend = "none"
	}

	switch c.JWTSigningAlgo {
	case "HS256", "HS384", "HS512":
		break
	case "RS256", "RS384", "RS512", "ES256", "ES384", "ES512":
		if c.JWTPrivateKeyFile == "" || c.JWTPublicKeyFile == "" {
			return errors.New("missing JWT priv or pub key for asymmetric algorithm")
		}
	case "":
		c.JWTSigningAlgo = "HS256"
	default:
		return errors.New("unknown JWT signing algorithm")
	}

	if c.ChatEjabberdName == "" {
		c.ChatEjabberdName = "wso.williams.edu"
	}

	if !c.MissingGoodrich() {
		if _, _, err := c.goodrichParseTime(c.GoodrichOpen); err != nil {
			return errors.New("bad goodrich open time")
		}
		if _, _, err := c.goodrichParseTime(c.GoodrichClose); err != nil {
			return errors.New("bad goodrich close time")
		}
		if c.Secrets.GoodrichEmailPassword == "" {
			return errors.New("goodrich email password missing")
		}
		if c.MissingEmail() {
			return errors.New("email SMTP settings missing")
		}
	}

	return nil
}

func setDefaultStr(str string, defaultValue string) string {
	if str == "" {
		return defaultValue
	}
	return str
}
