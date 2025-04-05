package onboarding

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/canonical"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/ks19"
	"github.com/WilliamsStudentsOnline/wso-go/services/onboarding/mgb4"
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

	//The ks19 group
	ks19Group := r.Group("/ks19")
	ks19.SetupRouter(ks19Group, db, cfg, log.Named("ks19"))

	//The mgb4 group
	mgb4Group := r.Group("/mgb4")
	mgb4.SetupRouter(mgb4Group, db, cfg, log.Named("mgb4"))

}
