package auth

import (
	"crypto/tls"
	"fmt"
	"net"

	"gopkg.in/ldap.v3"
)

const LDAPServer = "adldap.williams.edu"
const LDAPServerPort = 636

// Given a unix ID and a password, will check if the credentials are valid by
// authenticating into the OIT LDAP server.
// Returns a boolean where false means unauthenticated, true means authenticated.
// Also returns an error, such that if the error is not nil, we have an internal
// server error about the connection.
func OITAuth(unix, password string) (bool, error) {
	return LDAPAuth(
		LDAPServer,
		fmt.Sprintf("AD_WILLIAMS\\%s", unix),
		password,
		LDAPServerPort)
}

// Authenticates into an LDAP server (TLS) by binding a DN and password.
// Returns a boolean where false means unauthenticated, true means authenticated.
// Also returns an error, such that if the error is not nil, we have an internal
// server error about the connection.
func LDAPAuth(server, bindDN, password string, port int) (bool, error) {
	// Password cannot be blank
	if password == "" {
		return false, nil
	}

	// Connect to LDAP
	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", server, port), &tls.Config{
		ServerName:         server,
		InsecureSkipVerify: true, //TODO(EMERGENCY FIX 3/26/20): MAKE THIS false
	})
	if err != nil {
		return false, err
	}
	defer l.Close()

	// Try to bind to the DN and see if we authenticate
	err = l.Bind(bindDN, password)

	if err != nil {
		// If we have an error, check if it's an invalid credentials error.
		// If it is, return false but don't include the error.
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return false, nil
		}
		return false, err
	}

	// Client is authenticated; return true.
	return true, nil
}

var schoolSubnet = &net.IPNet{
	IP:   net.ParseIP("137.165.0.0"),
	Mask: net.CIDRMask(16, 32),
}

var localSubnet = &net.IPNet{
	IP:   net.ParseIP("192.168.0.0"),
	Mask: net.CIDRMask(24, 32),
}

// Check if the IP is on campus (or local) or not
func OnCampusIP(ipString string) bool {
	ip := net.ParseIP(ipString)

	if ip == nil {
		return false
	}

	return schoolSubnet.Contains(ip) || localSubnet.Contains(ip)
}
