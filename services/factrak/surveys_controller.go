package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// List all surveys
func (t *Controller) ListSurveys(c *gin.Context) {
	var surveys []*models.FactrakSurvey
	err := t.surveyModel.GetAllSurveys(&surveys)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, surveys)

	t.RespondOK(c, surveys)
}

// Get one survey
func (t *Controller) GetSurvey(c *gin.Context) {
	// Decode surveyID.
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var survey models.FactrakSurvey
	err = t.surveyModel.GetSurveyByID(surveyID, &survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	err = t.surveyModel.PopulateAgreementCounts(&survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Remove user info unless self, admin, or admin factrak
	if !auth.CheckIDIsSelf(c, survey.UserID) && !auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeAdminFactrak) {
		survey.UserID = 0
		survey.User = nil
	}

	t.RespondOK(c, survey)
}

type SurveyCreateParams struct {
	// Must include either this:
	CourseID *uint `json:"courseID"`

	// Or this:
	AreaOfStudyAbbreviation *string `json:"areaOfStudyAbbreviation"` // Like "CSCI"
	CourseNumber *string `json:"courseNumber"` // Like "256"

	// Must include this:
	ProfessorID *uint `json:"professorID" binding:"required"`

	// Params:
	Comment              string  `json:"comment" binding:"required"`
	WouldRecommendCourse *bool   `json:"wouldRecommendCourse"`
	CourseWorkload       *int    `json:"courseWorkload"`
	CourseStimulating    *int    `json:"courseStimulating"`
	WouldTakeAnother     *bool   `json:"wouldTakeAnother"`
	Approachability      *int    `json:"approachability"`
	LeadLecture          *int    `json:"leadLecture"`
	PromoteDiscussion    *int    `json:"promoteDiscussion"`
	OutsideHelpfulness   *int    `json:"outsideHelpfulness"`
	GradeReceived        *string `json:"gradeReceived"`
}

func (t *Controller) CreateSurvey(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := SurveyCreateParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondError(c, lib.ErrorMalformedRequestData)
		return
	}

	// We have the have either a course id or a area of study and course number
	if createData.CourseID == nil && (createData.AreaOfStudyAbbreviation == nil || createData.CourseNumber == nil) {
		t.RespondError(c, lib.ErrorSurveyMissingCourseParams)
		return
	} else if createData.CourseNumber != nil && *createData.CourseNumber == "" {
		// Course number cannot be blank
		t.RespondError(c, lib.ErrorSurveyMissingCourseParams)
		return
	}

	if len(createData.Comment) < 100 {
		t.RespondError(c, lib.ErrorCommentTooSmall)
		return
	}

	user := new(models.User)
	if err = t.studentModel.GetStudentByID(userID, user); err != nil {
		// Don't return 404; instead, return student not found
		if gorm.IsRecordNotFoundError(err) {
			err = lib.ErrorSurveyStudentNotFound
		}

		t.RespondError(c, err)
		return
	}

	if user.Student().Prefrosh() {
		t.RespondError(c, lib.ErrorUserCannotBePrefrosh)
		return
	}

	// Check if professor exists
	prof := new(models.User)
	if err = t.professorModel.GetProfessorByID(*createData.ProfessorID, prof); err != nil {
		// Don't return 404; instead, return professor not found
		if gorm.IsRecordNotFoundError(err) {
			err = lib.ErrorSurveyProfessorNotFound
		}

		t.RespondError(c, err)
		return
	}

	course := new(models.Course)
	// Deal with passed course
	if createData.CourseID != nil {
		// Check if course actually exists
		if err = t.courseModel.GetCourseByID(*createData.CourseID, course); err != nil {
			// Don't return 404; instead, return course not found
			if gorm.IsRecordNotFoundError(err) {
				err = lib.ErrorSurveyCourseNotFound
			}

			t.RespondError(c, err)
			return
		}
	} else {
		// Create a course
		// Check if area of study exists
		area := new(models.AreaOfStudy)
		if err = t.areaOfStudyModel.GetAreaOfStudyByAbbreviation(*createData.AreaOfStudyAbbreviation,area); err != nil {
			// Don't return 404; instead, return area of study not found
			if gorm.IsRecordNotFoundError(err) {
				err = lib.ErrorSurveyAreaOfStudyNotFound
			}

			t.RespondError(c, err)
			return
		}

		// "Create" the course, but don't stick it in the DB yet.
		course = &models.Course{
			Number: *createData.CourseNumber,
			AreaOfStudy: area,
		}

		// Note to future user: we can find or create this safely, as we have passed all validations except for
		// duplicate. But, if we create the course, it means that there are no duplicates, so we are good.
		if err = t.courseModel.FindOrCreate(course); err != nil {
			t.RespondError(c, err)
			return
		}
	}

	dup, err := t.surveyModel.CheckDuplicateSurvey(user.ID, prof.ID, course.ID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// If duplicate survey, return
	if dup {
		t.RespondError(c, lib.ErrorSurveyAlreadyExists)
		return
	}

	// Construct new survey
	survey := &models.FactrakSurvey{
		UserID: user.ID,
		ProfessorID: prof.ID,
		CourseID: course.ID,

		Comment: createData.Comment,
		WouldRecommendCourse: createData.WouldRecommendCourse,
		CourseWorkload: createData.CourseWorkload,
		CourseStimulating: createData.CourseStimulating,
		WouldTakeAnother: createData.WouldTakeAnother,
		Approachability: createData.Approachability,
		LeadLecture: createData.LeadLecture,
		PromoteDiscussion: createData.PromoteDiscussion,
		OutsideHelpfulness: createData.OutsideHelpfulness,
		GradeReceived: createData.GradeReceived,

		// Defaults
		TotalAgree: 0,
		TotalDisagree: 0,
	}

	err = t.surveyModel.CreateSurvey(survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, survey)
}