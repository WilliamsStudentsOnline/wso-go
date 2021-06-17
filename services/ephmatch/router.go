package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	r.GET("/availability", c.GetAvailability)

	// Requires ephmatch eligibility
	selfGroup := r.Group("", auth.RequireScopes(auth.ScopeEphmatch, auth.ScopeAdminAll))

	// Always do self profile
	// Get self profile
	selfGroup.GET("/profile", c.GetSelfProfile)
	// Create profile
	selfGroup.POST("/profile", c.CreateProfile)
	// Edit profile
	selfGroup.PATCH("/profile", c.UpdateProfile)
	// Delete profile
	selfGroup.DELETE("/profile", c.DeleteProfile)
	// Upload profile photo
	selfGroup.PUT("/profile/photo", c.UploadEphmatchProfilePhoto)
	// Delete profile photo
	selfGroup.DELETE("/profile/photo", c.DeleteEphmatchProfilePhoto)

	// Only get matches with scope
	matchesGroup := selfGroup.Group("", auth.RequireScopes(auth.ScopeEphmatchMatches, auth.ScopeAdminAll))
	// TODO: set seen:true?
	matchesGroup.GET("/matches", c.ListMatches)
	matchesGroup.GET("/matches-count", c.CountMatches)

	// Only get profiles with scope
	profilesGroup := matchesGroup.Group("", auth.RequireScopes(auth.ScopeEphmatchProfiles, auth.ScopeAdminAll))
	profilesGroup.GET("/profiles", c.ListProfiles)
	profilesGroup.GET("/profiles/:profileUserID", c.GetProfile)
	profilesGroup.PUT("/profiles/:profileUserID/relation", c.SetProfileRelation)
}
