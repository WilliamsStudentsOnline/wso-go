package bulletin

import (
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListRides godoc
// @Summary List rides
// @Description lists all bulletin rides
// @ID bulletins-list-rides
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param type query string false "Ride Type (request, offer)"
// @Param all query string false "Get All Rides (no restriction on date)"
// @Success 200 {array} models.BulletinRide
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/rides [get]
func (t *Controller) ListRides(c *gin.Context) {
	params := models.GetAllBulletinRidesOptions{}

	err := c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var rides []*models.BulletinRide
	err = t.rideModel.GetAllRides(&rides, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.rideModel.CountAllRides(&params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	// Remove user info if not a user. Need this, as bulletin service is public
	removeUserInfoFromRides(c, rides)

	t.RespondOK(c, rides)
}

// GetRide godoc
// @Summary Get ride
// @Description Get bulletin ride by ID with user preloaded
// @ID bulletins-get-ride
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param rideID path uint true "Ride ID"
// @Success 200 {object} models.BulletinRide
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/rides/{rideID} [get]
func (t *Controller) GetRide(c *gin.Context) {
	// Decode rideID.
	rideID, err := services.GetUIntParam(c, "rideID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var ride models.BulletinRide
	err = t.rideModel.GetRideByID(rideID, &ride)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Remove user info if not a user. Need this, as bulletin service is public
	if hasUserAuth(c) {
		sanitize.User(ride.User, c)
	} else {
		ride.User = nil
	}

	t.RespondOK(c, ride)
}

// CreateRideParams is a struct to hold the parameters used to create a ride.
type CreateRideParams struct {
	Body        string    `json:"body" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	Offer       *bool     `json:"offer"`
	Source      string    `json:"source" binding:"required"`
	Destination string    `json:"destination" binding:"required"`
}

// CreateRide godoc
// @Summary Create ride
// @Description Creates a ride
// @ID bulletins-create-ride
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param createParams body bulletin.CreateRideParams true "Create Ride Params"
// @Success 201 {object} models.BulletinRide
// @Failure 1830 {object} lib.APIError "date cannot be in past"
// @Failure 1101 {object} lib.APIError "request data validation failed"
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/rides [post]
func (t *Controller) CreateRide(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := CreateRideParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if createData.Offer == nil {
		t.RespondError(c, lib.ErrorBulletinInvalidType)
		return
	}

	// Date has to be past current time
	if createData.Date.Before(time.Now()) {
		t.RespondAPIError(c, lib.ErrorBulletinRideDateInPast)
		return
	}

	// Construct new ride
	ride := models.BulletinRide{
		Body:        createData.Body,
		Date:        createData.Date,
		Offer:       createData.Offer,
		Source:      createData.Source,
		Destination: createData.Destination,
		UserID:      userID,
	}

	err = t.rideModel.CreateRide(&ride)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, ride)
}

// UpdateRideParams is a struct to hold the parameters used to update a ride.
type UpdateRideParams struct {
	Body  *string    `json:"body"`
	Date  *time.Time `json:"date"`
	Offer *bool      `json:"offer"`
}

// UpdateRide godoc
// @Summary Update ride
// @Description Updates a ride
// @ID bulletins-update-ride
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param updateParams body bulletin.UpdateRideParams true "Update Ride Params"
// @Param rideID path uint true "Ride ID"
// @Success 200 {object} models.BulletinRide
// @Failure 1830 {object} lib.APIError "date cannot be in past"
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/rides/{rideID} [patch]
func (t *Controller) UpdateRide(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode parameter
	rideID, err := services.GetUIntParam(c, "rideID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Bind update params
	updateData := UpdateRideParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Do database query to get ride
	var ride models.BulletinRide
	err = t.rideModel.GetRideByID(rideID, &ride)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must only be able to update own ride
	if ride.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Date has to be past current time
	if updateData.Date != nil && updateData.Date.Before(time.Now()) {
		t.RespondAPIError(c, lib.ErrorBulletinRideDateInPast)
		return
	}

	// Update fields: this is a bit long and verbose, but I don't want to mess with reflect

	if updateData.Body != nil {
		ride.Body = *updateData.Body
	}
	if updateData.Date != nil {
		ride.Date = *updateData.Date
	}
	if updateData.Offer != nil {
		ride.Offer = updateData.Offer
	}

	// Update the ride in the db
	err = t.rideModel.UpdateRide(&ride)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Return update ride
	t.RespondOK(c, ride)

}

// DeleteRide godoc
// @Summary Delete ride
// @Description Deletes a ride
// @ID bulletins-delete-ride
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param rideID path uint true "Ride ID"
// @Success 200 {object} models.BulletinRide
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/rides/{rideID} [delete]
func (t *Controller) DeleteRide(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode rideID.
	rideID, err := services.GetUIntParam(c, "rideID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query to get ride
	var ride models.BulletinRide
	err = t.rideModel.GetRideByID(rideID, &ride)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must only be able to delete own ride (or must be admin)
	if ride.UserID != userID && !auth.HasScope(c, auth.ScopeAdminAll) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Delete ride
	err = t.rideModel.DeleteRide(&ride)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, ride)
}
