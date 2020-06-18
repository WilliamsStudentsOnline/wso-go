package onboarding

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/ahe2nht1"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/canonical"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	/* Set up the canonical group */
	//Group all URLs with /canonical
	canonicalGroup := r.Group("/canonical")
	// Send all URLs in this group to the canonical service
	canonical.SetupRouter(canonicalGroup, db, cfg, log.Named("canonical"))

	exerciseGroup := r.Group("/ahe2nht1")
	ahe2nht1.SetupRouter(exerciseGroup, db, cfg, log.Named("ahe2nht1"))

}
