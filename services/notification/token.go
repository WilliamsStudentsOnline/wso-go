package notification

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type TokenCreateParams struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// CreateToken godoc
// @Summary Create notification token
// @Description creates self's notification token
// @ID createNotificationToken
// @Tags notification
// @Accept  json
// @Produce  json
// @Param updateParams body notification.TokenCreateParams true "Create Token Params"
// @Success 200 {object} models.NotificationToken
// @Failure 1101 {object} services.BaseErrorResponse "request data validation failed"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /notification/app/token [post]
func (t *Controller) CreateToken(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind create params
	createData := TokenCreateParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if !models.ValidateNotificationTokenType(createData.Type) {
		t.RespondError(c, lib.ErrorNotificationInvalidTokenType)
		return
	}

	if len(createData.Token) == 0 {
		t.RespondError(c, lib.ErrorNotificationEmptyToken)
		return
	}

	newToken := models.NotificationToken{
		UserID: userID,
		Type:   createData.Type,
		Token:  createData.Token,
	}

	// Do database query
	err = t.notifTokenModel.CreateToken(&newToken)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, newToken)
}
