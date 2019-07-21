package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Secrets struct {
	JWTSecretKey    string `yaml:"jwt_secret_key"`
	WsoLdapDN       string `yaml:"wso_ldap_dn"`
	WsoLdapPassword string `yaml:"wso_ldap_password"`
}

func (s *Secrets) RequireLDAPAuth() error {
	if s.WsoLdapDN == "" || s.WsoLdapPassword == "" {
		return errors.New("LDAP auth secrets required")
	}
	return nil
}

// Get the secrets file and parse any info
func GetSecrets(secretsPath string) (*Secrets, error) {
	// Get the secrets file and decode it
	secretsPath, err := filepath.Abs(secretsPath)
	if err != nil {
		return nil, err
	}

	secretsFile, err := os.Open(secretsPath)
	if err != nil {
		return nil, err
	}

	secrets := new(Secrets)

	dec := yaml.NewDecoder(secretsFile)
	err = dec.Decode(secrets)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}
