package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, log *zap.SugaredLogger) {
	c := services.BaseController{Log: log}
	r.Any("/*path", func(ctx *gin.Context) {
		c.RespondAPIError(ctx, lib.ErrorEphmatchDeprecated)
	})
}
