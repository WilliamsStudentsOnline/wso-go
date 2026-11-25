package courses

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	if client == nil {
		if err := ConfigureClient(cfg.RedisAddr); err != nil {
			log.Debugf("redis unavailable, term catalogs won't be cached: %v", err)
		}
	}
	c := NewController(db, cfg, log)
	factrakAccess := auth.RequireScopes(auth.ScopeFactrakFull, auth.ScopeFactrakAdmin)

	r.GET("/terms", c.ListTerms)
	r.GET("/terms/:strm", c.GetTermCatalog)
	r.GET("/terms/:strm/factrak", factrakAccess, c.GetTermFactrak)
	r.GET("/terms/:strm/:crseID", c.GetCourseInSemester)
	r.GET("/search", c.SearchCourses)
	r.GET("/catalog/:crseID", c.GetCourse)
	r.GET("/catalog/:crseID/reviews", factrakAccess, c.ListCourseReviews)
	r.GET("/catalog/:crseID/ratings", factrakAccess, c.GetCourseRatings)
	r.GET("/catalog/:crseID/stats", factrakAccess, c.GetCourseStats)
	r.GET("/catalog/:crseID/related", c.GetCourseRelated)
	r.GET("/attributes", c.ListAttributes)

	r.GET("/professors", c.SearchProfessors)
	r.GET("/professors/:professorID", c.GetProfessor)
	r.GET("/professors/:professorID/ratings", factrakAccess, c.GetProfessorRatings)
	r.GET("/professors/:professorID/reviews", factrakAccess, c.ListProfessorReviews)
	r.GET("/professors/:professorID/stats", factrakAccess, c.GetProfessorStats)
	r.GET("/areas", c.ListAreas)
}
