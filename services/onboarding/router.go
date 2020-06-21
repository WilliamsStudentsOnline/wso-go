package onboarding

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/canonical"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/ks19"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	/* Set up the canonical group */
	// Group all URLs with /canonical
	canonicalGroup := r.Group("/canonical")
	// Send all URLs in this group to the canonical service
	canonical.SetupRouter(canonicalGroup, db, cfg, log.Named("canonical"))

	unixGroup := r.Group("/ks19")
	ks19.SetupRouter(unixGroup, db, cfg, log.Named("canonical"))


}

