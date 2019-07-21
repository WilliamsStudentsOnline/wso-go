package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLDAP_Get(t *testing.T) {
	ldaps := []*LDAP{NewWilliamsLDAP(), NewNDSLDAP()}

	for _, ldap := range ldaps {
		t.Run("LDAP="+ldap.Host, func(t *testing.T) {
			err := ldap.Connect()
			assert.NoError(t, err)
			defer ldap.Close()

			uid := "al15"

			// May have to replace this with another user when al15 graduates (in 2022)
			entry, err := ldap.Get("uid", uid)
			assert.NoError(t, err)

			entry.Print()

			assert.Equal(t, uid, entry.GetAttributeValue("uid"))
		})
	}
}
