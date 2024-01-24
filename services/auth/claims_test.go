package auth

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestRemoveBannedScope(t *testing.T) {
	scope := []string{auth.ScopeChat, auth.ScopeUsers, auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeFactrakFull,
		auth.ScopeFactrakLimited, auth.ScopeBulletinWrite, auth.ScopeEphmatch, auth.ScopeDormtrakWrite,
		auth.ScopeDormtrak, auth.ScopeDormtrakFull, auth.ScopeDormtrakLimited}
	banInfo := models.BannedUser{
		Factrak:       true,
		Dormtrak:      false,
		Ephcatch:      false,
		BulletinRead:  true,
		BulletinWrite: true,
		Ephmatch:      true,
	}

	removeBannedScope(&scope, &banInfo)

	t.Log(scope)

	testify.ElementsMatch(t, scope, []string{auth.ScopeChat, auth.ScopeUsers, auth.ScopeWriteSelf,
		auth.ScopeDormtrakWrite, auth.ScopeDormtrak})
}

func TestBannedUser(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()
	log := zaptest.NewLogger(t).Sugar()

	user := models.User{
		Type:                     "student",
		HasAcceptedFactrakPolicy: lib.BoolToPtr(true),
	}

	assert.NoError(db.Create(&user).Error)

	assert.NoError(db.Create(&models.BannedUser{
		User:          &user,
		Reason:        "cringe",
		Factrak:       false,
		Dormtrak:      false,
		Ephcatch:      false,
		Ephmatch:      false,
		BulletinRead:  true,
		BulletinWrite: true,
	}).Error)

	genClaims := GenerateClaimsFactory(cfg, db, log)

	claims := genClaims(&AuthenticatorPayload{
		TokenLevel: TokenLevelUser,
		User:       &user,
	}, TokenTypeAPI)

	assert.NotContains(claims["scope"], auth.ScopeBulletin)
	assert.NotContains(claims["scope"], auth.ScopeBulletinWrite)
	assert.Contains(claims["scope"], auth.ScopeUsers)
	assert.Contains(claims["scope"], auth.ScopeFactrakLimited)
	assert.Contains(claims["scope"], auth.ScopeDormtrakLimited)
}
