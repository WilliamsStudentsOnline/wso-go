package ldap

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

// TODO: Make this a connection pool (eg https://github.com/vetinari/go-ldappool)
type LDAP struct {
	Host  string
	Port  int
	Base  string
	Scope int
	conn  *ldap.Conn
}

// NewWilliamsLDAP connects to the Williams LDAP server
// Ye Shu Note Feb 2024: this is not updated since July 2023 and deprecated in favor of ODIR
func NewWilliamsLDAP() *LDAP {
	return &LDAP{
		Host:  "ldap://ldap.williams.edu",
		Base:  "ou=people,o=williams",
		Scope: ldap.ScopeWholeSubtree,
		Port:  389,
	}
}

func NewNDSLDAP() *LDAP {
	return &LDAP{
		Host:  "ldap://nds4.williams.edu",
		Base:  "o=williams",
		Scope: ldap.ScopeWholeSubtree,
		Port:  389,
	}
}

// NewADLDAP connects to the Williams AD LDAP server
// Ye Shu Note Feb 2024: this is used for authenticating users
// It contains all users, including students, faculty, staff, and some who have left the college
func NewADLDAP() *LDAP {
	return &LDAP{
		Host:  "ldaps://adldap.williams.edu",
		Base:  "ou=williams,dc=ad,dc=williams,dc=edu",
		Scope: ldap.ScopeWholeSubtree,
		Port:  636,
	}
}

// NewODIRLDAP connects to the OIT secret ODIR LDAP server
// Ye Shu Note Feb 2024: this is only accessible from wso-vm and wso-vm-dev
// Also, I have configured firewall to allow outgoing traffic to this server
// This is the backend LDAP server that powers the Online directory so it only contains active students, faculty, and staff.
func NewODIRLDAP() *LDAP {
	return &LDAP{
		Host:  "ldap://odirldap-vip.williams.edu",
		Base:  "ou=users,dc=odirldap,dc=williams,dc=edu",
		Scope: ldap.ScopeWholeSubtree,
		Port:  389,
	}
}

func (l *LDAP) ConnectWithBind(bindDN, password string) error {
	err := l.Connect()
	if err != nil {
		return err
	}

	err = l.conn.Bind(bindDN, password)
	if err != nil {
		l.Close()
		return err
	}

	return nil
}

func (l *LDAP) Connect() error {
	conn, err := ldap.DialURL(fmt.Sprintf("%s:%d", l.Host, l.Port))
	if err != nil {
		return err
	}

	l.conn = conn

	return nil
}

func (l *LDAP) Close() {
	if l.conn != nil {
		l.conn.Close()
		l.conn = nil
	}
}

// Gets an LDAP entry based on key/value. Pass optional attributes to get specific fields.
// If there is no entry, return nil. If there are more than one entries, return the first one.
func (l *LDAP) Get(key, value string, attributes ...string) (*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		l.Base,
		l.Scope, ldap.NeverDerefAliases, 0, 0, false,
		// Filter:
		fmt.Sprintf("(%s=%s)", key, value),
		// A list attributes to retrieve
		attributes,
		nil,
	)

	res, err := l.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(res.Entries) == 0 {
		return nil, nil
	}

	return res.Entries[0], nil
}

func (l *LDAP) Each(key, value string, attributes ...string) ([]*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		l.Base,
		l.Scope, ldap.NeverDerefAliases, 0, 0, false,
		// Filter:
		fmt.Sprintf("(%s=%s)", key, value),
		// A list attributes to retrieve
		attributes,
		nil,
	)

	res, err := l.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	return res.Entries, nil
}
