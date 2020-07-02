package admin

import (
	"github.com/gin-gonic/gin"
)

type Stats struct {
	NumberOfFactrakReviews  int     `json:"numberOfFactrakReviews"`
	NumberOfEphmatchMatches int     `json:"numberOfEphmatchMatches"`
	NumberOfEphmatchUsers   int     `json:"numberOfEphmatchUsers"`
	NumberOfUsers           int     `json:"numberOfUsers"`
	NumberOfUserStudents    int     `json:"numberOfUserStudents"`
	PercentageUseEphmatch   float64 `json:"percentageUseEphmatch"`
}

// Returns user statistics
// GetStats godoc
// @Summary Returns user statistics
// @Description Calls methods that query database and returns the information as statistics
// @ID get-stats
// @Tags admin
// @Accept  json
// @Produce  json
// @Success 201 {object} admin.Stats
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /admin/get-stats [get]

func (t *Controller) GetStats(c *gin.Context) {
	var stat Stats
	var err error

	stat.NumberOfFactrakReviews, err = t.factrakSurveyModel.CountSurveys()
	if err != nil {
		t.RespondError(c, err) // maybe if error, respond by setting stat.Statistic=0 instead?
	}

	stat.NumberOfEphmatchMatches, err = t.ephmatchMatchesModel.CountMatches()
	if err != nil {
		t.RespondError(c, err)
	}

	stat.NumberOfEphmatchUsers, err = t.ephmatchProfileModel.CountProfiles()
	if err != nil {
		t.RespondError(c, err)
	}

	stat.NumberOfUsers, err = t.userModel.CountAllUsers()
	if err != nil {
		t.RespondError(c, err)
	}
	stat.NumberOfUserStudents, err = t.userModel.CountAllStudents()
	if err != nil {
		t.RespondError(c, err)
	}

	// convert to float64 so more exact percentages can be displayed
	convertedNumberEphmatchUsers := float64(stat.NumberOfEphmatchUsers)
	convertedNumberUserStudents := float64(stat.NumberOfUserStudents)
	if convertedNumberUserStudents > 0 {
		stat.PercentageUseEphmatch = (convertedNumberEphmatchUsers / convertedNumberUserStudents) * 100
	}

	t.RespondOK(c, stat)
}

// added methods to user.go, ephmatch_match.go, factrack_survey.go
