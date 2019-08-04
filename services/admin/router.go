package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db, cfg)
	r.POST("/catalog-update", c.CatalogUpdate)
	r.POST("/update-all-users-from-ldap", c.UpdateAllUsersFromLDAP)
	r.POST("/update-all-factrak-survey-deficits", c.UpdateAllFactrakSurveyDeficits)
}
