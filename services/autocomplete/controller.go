package autocomplete

import (
	"net/http"
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/autocomplete"
	"github.com/WilliamsStudentsOnline/wso-go/lib/autocomplete/autocompletor"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Controller refers to the struct for the bulletinModel
type Controller struct {
	services.BaseController
	autocomplete autocomplete.Autocomplete
}

// NewController constructs a new user controller
func NewController(db *gorm.DB, cfg *config.Config) *Controller {
	return &Controller{
		autocomplete: autocompletor.NewAutocomplete(cfg, db),
	}
}

// AreaOfStudy godoc
// @Summary Autocomplete area of study
// @Description Given an input q, this autocompletes the area of study. Results are sorted by levenshtein distance.
// @ID autocomplete-area-of-study
// @Tags autocomplete
// @Accept  json
// @Produce  json
// @Param q query string true "String to Complete"
// @Param limit query int false "Limit"
// @Success 200 {array} autocomplete.ACEntry
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /autocomplete/area-of-study [get]
func (t *Controller) AreaOfStudy(c *gin.Context) {
	t.doAutocomplete(c, t.autocomplete.AreaOfStudy)
}

// Course godoc
// @Summary Autocomplete course
// @Description Given an input q, this autocompletes the course. Results are sorted by levenshtein distance.
// @ID autocomplete-course
// @Tags autocomplete
// @Accept  json
// @Produce  json
// @Param q query string true "String to Complete"
// @Param limit query int false "Limit"
// @Success 200 {array} autocomplete.ACEntry
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /autocomplete/course [get]
func (t *Controller) Course(c *gin.Context) {
	t.doAutocomplete(c, t.autocomplete.Course)
}

// Professor godoc
// @Summary Autocomplete professor
// @Description Given an input q, this autocompletes the professor. Results are sorted by levenshtein distance.
// @ID autocomplete-professor
// @Tags autocomplete
// @Accept  json
// @Produce  json
// @Param q query string true "String to Complete"
// @Param limit query int false "Limit"
// @Success 200 {array} autocomplete.ACEntry
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /autocomplete/professor [get]
func (t *Controller) Professor(c *gin.Context) {
	t.doAutocomplete(c, t.autocomplete.Professor)
}

// Tag godoc
// @Summary Autocomplete tag
// @Description Given an input q, this autocompletes the tag. Results are sorted by levenshtein distance.
// @ID autocomplete-tag
// @Tags autocomplete
// @Accept  json
// @Produce  json
// @Param q query string true "String to Complete"
// @Param limit query int false "Limit"
// @Success 200 {array} autocomplete.ACEntry
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /autocomplete/tag [get]
func (t *Controller) Tag(c *gin.Context) {
	t.doAutocomplete(c, t.autocomplete.Tag)
}

// Factrak godoc
// @Summary Autocomplete factrak professors and courses
// @Description Given an input q, this autocompletes the professor or course. Results are sorted by levenshtein distance.
// @ID autocomplete-factrak
// @Tags autocomplete
// @Accept  json
// @Produce  json
// @Param q query string true "String to Complete"
// @Param limit query int false "Limit"
// @Success 200 {array} autocomplete.ACEntry
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /autocomplete/factrak [get]
func (t *Controller) Factrak(c *gin.Context) {
	t.doAutocomplete(c, t.autocomplete.Factrak)
}

// Since all of the autocomplete control code is the same, we just put it here and then call it from the control
// functions, with a specified autocomplete method.
func (t *Controller) doAutocomplete(c *gin.Context, f func(string) ([]autocomplete.ACEntry, error)) {
	q := c.Query("q")
	if q == "" {
		t.RespondOK(c, []autocomplete.ACEntry{})
		return
	}

	entries, err := f(q)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			t.RespondErrorCode(c, http.StatusBadRequest, err)
			return
		}
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	t.RespondOK(c, entries)
}
