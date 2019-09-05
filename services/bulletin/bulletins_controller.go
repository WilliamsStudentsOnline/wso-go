package bulletin

import (
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListBulletins godoc
// @Summary List bulletins
// @Description lists all bulletins
// @ID bulletins-list-bulletins
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param type query string false "Bulletin Type"
// @Param all query string false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {array} models.Bulletin
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/bulletins [get]
func (t *Controller) ListBulletins(c *gin.Context) {
	params := new(models.GetAllBulletinsOptions)

	err := c.ShouldBindQuery(params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	var bulletins []*models.Bulletin
	err = t.bulletinModel.GetAllBulletinsWithOptions(&bulletins, params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.bulletinModel.CountAllBulletinsWithOptions(params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	// Remove user info if not a user. Need this, as bulletin service is public
	removeUserInfoFromBulletin(c, bulletins)

	t.RespondOK(c, bulletins)
}

// GetBulletin godoc
// @Summary Get bulletin
// @Description Get bulletin by ID with user preloaded
// @ID bulletins-get-bulletin
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param bulletinID path uint true "Bulletin ID"
// @Success 200 {object} models.Bulletin
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/bulletins/{bulletinID} [get]
func (t *Controller) GetBulletin(c *gin.Context) {
	// Decode bulletinID.
	bulletinID, err := services.GetUIntParam(c, "bulletinID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var bulletin models.Bulletin
	err = t.bulletinModel.GetBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Remove user info if not a user. Need this, as bulletin service is public
	if !hasUserAuth(c) {
		bulletin.User = nil
	}

	t.RespondOK(c, bulletin)
}

// CreateBulletinParams is a struct to hold the parameters used to create a bulletin.
type CreateBulletinParams struct {
	Type      string     `json:"type" binding:"required"`
	Title     string     `json:"title" binding:"required"`
	Body      string     `json:"body" binding:"required"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	Offer     *bool      `json:"offer"`
}

// CreateBulletin godoc
// @Summary Create bulletin
// @Description Create a bulletin
// @ID bulletins-create-bulletin
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param createParams body bulletin.CreateBulletinParams true "Create Bulletin Params"
// @Success 201 {object} models.Bulletin
// @Failure 1830 {object} lib.APIError "start date cannot be after end date"
// @Failure 1831 {object} lib.APIError "invalid bulletin type"
// @Failure 1101 {object} lib.APIError "request data validation failed"
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/bulletins [post]
func (t *Controller) CreateBulletin(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := CreateBulletinParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Default start date to now time (we also do this as a hook, but it's useful here too).
	if createData.StartDate == nil {
		createData.StartDate = lib.TimeToPtr(time.Now())
	}

	// Start date (or default time now) has to be before end date, if we have end date as a field.
	if createData.EndDate != nil {
		if createData.StartDate.After(*createData.EndDate) {
			t.RespondAPIError(c, lib.ErrorBulletinInvalidDates)
			return
		}
	}

	// Construct new bulletin
	bulletin := models.Bulletin{
		Type:      createData.Type,
		Title:     createData.Title,
		Body:      createData.Body,
		StartDate: *createData.StartDate,
		EndDate:   createData.EndDate,
		Offer:     createData.Offer,

		UserID: userID,
	}

	if !checkValidType(&bulletin) {
		t.RespondAPIError(c, lib.ErrorBulletinInvalidType)
		return
	}

	err = t.bulletinModel.CreateBulletin(&bulletin)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, bulletin)
}

// UpdateBulletinParams is a struct to hold the parameters used to update a bulletin.
type UpdateBulletinParams struct {
	Title     *string    `json:"title"`
	Body      *string    `json:"body"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	Offer     *bool      `json:"offer"`
}

// UpdateBulletin godoc
// @Summary Update bulletin
// @Description Updates a bulletin
// @ID bulletins-update-bulletin
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param updateParams body bulletin.UpdateBulletinParams true "Update Bulletin Params"
// @Param bulletinID path uint true "Bulletin ID"
// @Success 200 {object} models.Bulletin
// @Failure 1830 {object} lib.APIError "start date cannot be after end date"
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/bulletins/{bulletinID} [patch]
func (t *Controller) UpdateBulletin(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode parameter
	bulletinID, err := services.GetUIntParam(c, "bulletinID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Bind update params
	updateData := UpdateBulletinParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query to get bulletin
	var bulletin models.Bulletin
	err = t.bulletinModel.GetBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must only be able to update own bulletin
	if bulletin.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Update fields: this is a bit long and verbose, but I don't want to mess with reflect
	bulletin.Title = *lib.StrPtrDefaults(updateData.Title, &bulletin.Title)
	bulletin.Body = *lib.StrPtrDefaults(updateData.Body, &bulletin.Body)
	bulletin.StartDate = *lib.TimePtrDefaults(updateData.StartDate, &bulletin.StartDate)
	bulletin.EndDate = lib.TimePtrDefaults(updateData.EndDate, bulletin.EndDate)
	bulletin.Offer = lib.BoolPtrDefaults(updateData.Offer, bulletin.Offer)

	// Start date (or default time now) has to be before end date, if we have end date as a field.
	if updateData.StartDate != nil || updateData.EndDate != nil {
		if bulletin.StartDate.After(*bulletin.EndDate) {
			t.RespondAPIError(c, lib.ErrorBulletinInvalidDates)
			return
		}
	}

	// Update the bulletin in the db
	err = t.bulletinModel.UpdateBulletin(&bulletin)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Return update bulletin
	t.RespondOK(c, bulletin)

}

// DeleteBulletin godoc
// @Summary Delete bulletin
// @Description Deletes a bulletin
// @ID bulletins-delete-bulletin
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param bulletinID path uint true "Bulletin ID"
// @Success 200 {object} models.Bulletin
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/bulletins/{bulletinID} [delete]
func (t *Controller) DeleteBulletin(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode bulletinID.
	bulletinID, err := services.GetUIntParam(c, "bulletinID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query to get bulletin
	var bulletin models.Bulletin
	err = t.bulletinModel.GetBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must only be able to delete own bulletin (or must be admin)
	if bulletin.UserID != userID && !auth.HasScope(c, auth.ScopeAdminAll) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Delete Bulletin
	err = t.bulletinModel.DeleteBulletin(&bulletin)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, bulletin)
}

func checkValidType(b *models.Bulletin) bool {
	if b.IsAnnouncement() || b.IsExchange() || b.IsJob() || b.IsLostAndFound() {
		return true
	}
	return false
}
