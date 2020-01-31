package auth

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
)

type TokenLevel int

// Token signing levels for payload to give scopes
const (
	TokenLevelUnauthenticated TokenLevel = iota
	TokenLevelOffCampus
	TokenLevelOnCampus
	TokenLevelUser
)

type AuthResponse struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

type AuthenticatorPayload struct {
	TokenLevel TokenLevel
	User       *models.User
}
