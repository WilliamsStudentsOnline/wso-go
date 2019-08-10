package ldap

import (
	"fmt"

	"gopkg.in/ldap.v3"
)

// TODO: Make this a connection pool (eg https://github.com/vetinari/go-ldappool)
type LDAP struct {
	Host  string
	Port  int
	Base  string
	Scope int
	conn  *ldap.Conn
}

func NewWilliamsLDAP() *LDAP {
	return &LDAP{
		Host:  "ldap.williams.edu",
		Base:  "ou=people,o=williams",
		Scope: ldap.ScopeWholeSubtree,
		Port:  389,
	}
}

func NewNDSLDAP() *LDAP {
	return &LDAP{
		Host:  "nds1.williams.edu",
		Base:  "o=williams",
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
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", l.Host, l.Port))
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
