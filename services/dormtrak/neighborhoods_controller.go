package dormtrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all neighborhoods
// ListNeighborhoods godoc
// @Summary List neighborhoods
// @Description lists all neighborhoods
// @ID dormtrak-list-neighborhoods
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Neighborhood
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/neighborhoods [get]
func (t *Controller) ListNeighborhoods(c *gin.Context) {
	var neighborhoods []*models.Neighborhood
	err := t.neighborhoodModel.GetAllNeighborhoods(&neighborhoods)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, neighborhoods)
}

// GetNeighborhood godoc
// @Summary Get neighborhood
// @Description gets one neighborhood with dorms preloaded
// @ID dormtrak-get-neighborhood
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param neighborhoodID path uint true "Neighborhood ID"
// @Success 200 {object} models.Neighborhood
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/neighborhoods/{neighborhoodID} [get]
func (t *Controller) GetNeighborhood(c *gin.Context) {
	// Decode neighborhoodID.
	neighborhoodID, err := services.GetUIntParam(c, "neighborhoodID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var neighborhood models.Neighborhood
	err = t.neighborhoodModel.GetNeighborhoodByID(neighborhoodID, &neighborhood)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, neighborhood)
}

// GetNeighborhoodFacts godoc
// @Summary Get neighborhood facts
// @Description gets neighborhood facts of one neighborhood.
// @ID dormtrak-get-neighborhood-facts
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param neighborhoodID path uint true "Neighborhood ID"
// @Success 200 {object} models.NeighborhoodFacts
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/neighborhoods/{neighborhoodID}/facts [get]
func (t *Controller) GetNeighborhoodFacts(c *gin.Context) {
	// Decode dormID.
	neighborhoodID, err := services.GetUIntParam(c, "neighborhoodID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	facts := models.NeighborhoodFacts{}
	err = t.neighborhoodModel.GetNeighborhoodFacts(neighborhoodID, &facts)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, facts)
}
