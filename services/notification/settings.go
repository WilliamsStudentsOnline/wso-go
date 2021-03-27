package notification

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// GetSettings godoc
// @Summary Get notification settings
// @Description gets self's notification settings
// @ID notification-get-settings
// @Tags notification
// @Accept  json
// @Produce  json
// @Success 200 {object} models.NotificationSettings
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /notification/settings [get]
func (t *Controller) GetSettings(c *gin.Context) {
	userID := services.GetUserID(c)

	var settings models.NotificationSettings

	// Do database query
	err := t.notifSettingsModel.GetOrCreateSettings(userID, &settings)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, settings)
}

type SettingsUpdateParams struct {
	EnableNotifications *bool `json:"enableNotifications"`
	SalmonNotify        *bool `json:"salmonNotify"`
}

// UpdateSettings godoc
// @Summary Update notification settings
// @Description updates (or creates if it does not exist) self's notification settings
// @ID notification-update-settings
// @Tags notification
// @Accept  json
// @Produce  json
// @Param updateParams body notification.SettingsUpdateParams true "Update Settings Params"
// @Success 200 {object} models.NotificationSettings
// @Failure 1101 {object} lib.APIError "request data validation failed"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /notification/settings [patch]
func (t *Controller) UpdateSettings(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	updateData := SettingsUpdateParams{}
	err := c.ShouldBind(&updateData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var settings models.NotificationSettings
	err = t.notifSettingsModel.GetOrCreateSettings(userID, &settings)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	settings.EnableNotifications = *lib.BoolPtrDefaults(updateData.EnableNotifications, &settings.EnableNotifications)
	settings.SalmonNotify = *lib.BoolPtrDefaults(updateData.SalmonNotify, &settings.SalmonNotify)

	// Update in DB
	err = t.notifSettingsModel.UpdateSettings(&settings)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, settings)
}
