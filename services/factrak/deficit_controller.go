package factrak

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type DeficitExemptionCreateParams struct {
	// Why the student can't write this semester's surveys; at most 1000 characters
	Reason string `json:"reason" binding:"required"`
}

// No status, since the student doesn't see whether the request is approved
type DeficitExemptionCreatedResponse struct {
	ID             uint   `json:"id"`
	SemesterSeason string `json:"semesterSeason" enums:"fall,spring"`
	SemesterYear   int    `json:"semesterYear"`
}

// @Summary Request a survey deficit exemption
// @Description ask to be excused from this semester's Factrak survey requirement; a moderator approves or rejects it
// @ID createDeficitExemption
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param createParams body factrak.DeficitExemptionCreateParams true "Create Deficit Exemption Params"
// @Success 201 {object} factrak.DeficitExemptionCreatedResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data; 1101: request data validation failed; 1533: user must be a student and could not be found; 1561: reason is required and must be 1000 characters or fewer"
// @Failure 409 {object} services.BaseErrorResponse "1560: a deficit exemption was already requested for this semester"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /factrak/deficit/exemptions [post]
func (t *Controller) CreateDeficitExemption(c *gin.Context) {
	userID := services.GetUserID(c)

	createData := DeficitExemptionCreateParams{}
	if err := c.ShouldBind(&createData); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	reason := strings.TrimSpace(createData.Reason)
	if reason == "" || utf8.RuneCountInString(reason) > models.FactrakDeficitExemptionReasonMaxLength {
		t.RespondError(c, lib.ErrorDeficitExemptionReasonInvalid)
		return
	}

	user := new(models.User)
	if err := t.studentModel.GetStudentByID(userID, user); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = lib.ErrorSurveyStudentNotFound
		}
		t.RespondError(c, err)
		return
	}

	exemption, err := t.exemptionModel.CreateExemptionRequest(user.ID, reason, time.Now())
	if err != nil {
		if err == models.ErrFactrakDeficitExemptionExists {
			err = lib.ErrorDeficitExemptionExists
		}
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, DeficitExemptionCreatedResponse{
		ID:             exemption.ID,
		SemesterSeason: exemption.SemesterSeason,
		SemesterYear:   exemption.SemesterYear,
	})
}
