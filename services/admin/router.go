package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)
	r.POST("/catalog-update", c.CatalogUpdate)
	r.POST("/update-all-users-from-ldap", c.UpdateAllUsersFromLDAP)
	r.POST("/update-all-factrak-survey-deficits", c.UpdateAllFactrakSurveyDeficits)
	r.POST("/dorms-update", c.DormsUpdate)

	r.GET("/jobs/:jobID/status", c.GetJobStatus)
	r.GET("/get-stats", c.GetStats)
}
