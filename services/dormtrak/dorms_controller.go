package dormtrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListDorms godoc
// @Summary List dorms
// @Description lists all dorms
// @ID dormtrak-list-dorms
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param q query string false "Search Query"
// @Success 200 {array} models.Dorm
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/dorms [get]
func (t *Controller) ListDorms(c *gin.Context) {
	var dorms []*models.Dorm
	var err error

	opts := models.GetAllDormsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if query, ok := c.GetQuery("q"); ok {
		err = t.dormtrakSearch.SearchDorms(query, &dorms, &opts)
	} else {
		err = t.dormModel.GetAllDorms(&dorms, &opts)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, dorms)
}

// GetDorm godoc
// @Summary Get dorm
// @Description gets one dorm with neighborhood and dorm rooms preloaded
// @ID dormtrak-get-dorm
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param dormID path uint true "Dorm ID"
// @Success 200 {object} models.Dorm
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/dorms/{dormID} [get]
func (t *Controller) GetDorm(c *gin.Context) {
	// Decode dormID.
	dormID, err := services.GetUIntParam(c, "dormID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var dorm models.Dorm
	err = t.dormModel.GetDormByID(dormID, &dorm)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, dorm)
}

// GetDormRooms godoc
// @Summary Get dorm rooms
// @Description gets dorm rooms of one dorm building
// @ID dormtrak-get-dorm-rooms
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param dormID path uint true "Dorm ID"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {array} models.DormRoom
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/dorms/{dormID}/rooms [get]
func (t *Controller) GetDormRooms(c *gin.Context) {
	// Decode dormID.
	dormID, err := services.GetUIntParam(c, "dormID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if dorm exists
	exists, err := t.dormModel.DoesDormExist(dormID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	pOff, pLim, err := services.GetPaginationParams(c)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var rooms []*models.DormRoom
	err = t.dormRoomModel.GetDormRoomsByDorm(dormID, &rooms, t.dormRoomModel.NewDormRoomPaginate(pOff, pLim))
	if err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.dormRoomModel.CountDormRoomsByDorm(dormID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	t.RespondOK(c, rooms)
}

// GetDormFacts godoc
// @Summary Get dorm facts
// @Description gets dorm facts of one dorm building. Some of these facts are in GetDorm, while others are generated here.
// @ID dormtrak-get-dorm-facts
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param dormID path uint true "Dorm ID"
// @Success 200 {object} models.DormFacts
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/dorms/{dormID}/facts [get]
func (t *Controller) GetDormFacts(c *gin.Context) {
	// Decode dormID.
	dormID, err := services.GetUIntParam(c, "dormID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	facts := models.NewDormFacts()
	err = t.dormModel.GetDormFacts(dormID, facts)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, facts)
}
