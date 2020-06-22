package onboarding

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/canonical"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	/* Set up the rmn1 group */
	// Group all URLs with /rmn1
	rmn1Group := r.Group("/rmn1")
	// Send all URLs in this group to the rmn1 service
	canonical.SetupRouter(rmn1Group, db, cfg, log.Named("rmn1"))
}
