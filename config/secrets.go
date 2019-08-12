package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Secrets struct {
	JWTSecretKey    string `yaml:"jwt_secret_key" envconfig:"jwt_secret_key"`
	WsoLdapDN       string `yaml:"wso_ldap_dn" envconfig:"wso_ldap_dn"`
	WsoLdapPassword string `yaml:"wso_ldap_password" envconfig:"wso_ldap_password"`

	// MySQL
	MySQLPassword string `yaml:"mysql_password" envconfig:"mysql_password"`
}

func (s *Secrets) RequireLDAPAuth() error {
	if s.WsoLdapDN == "" || s.WsoLdapPassword == "" {
		return errors.New("LDAP auth secrets required")
	}
	return nil
}

// Get the secrets file and parse any info. Environment variables will like like WSO_SECRET_SECRET_TEXT_HERE
func LoadSecrets(secretsPath string) (*Secrets, error) {
	// We first get env variables for config. These override everything other form of config.
	secrets := &Secrets{}
	err := envconfig.Process("wso_secret", secrets)

	// If we have no file, just return.
	if secretsPath == "" {
		return secrets, nil
	}

	// Get the secrets file and decode it
	secretsPath, err = filepath.Abs(secretsPath)
	if err != nil {
		return nil, err
	}

	secretsFile, err := os.Open(secretsPath)
	if err != nil {
		return nil, err
	}

	yamlSecrets := Secrets{}

	dec := yaml.NewDecoder(secretsFile)
	err = dec.Decode(&yamlSecrets)
	if err != nil {
		return nil, err
	}

	// Set any empty values from secrets that are non-empty in yamlSecrets to the ones in yamlSecrets
	err = mergo.Merge(secrets, yamlSecrets)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}
