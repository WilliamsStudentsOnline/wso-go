package factrak

import (
	"net/http"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// List all surveys
// @Summary List surveys
// @Description lists all surveys where the professor is at Williams and the survey is current. Order by creation date.
// @ID listSurveys
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param professorID query int false "Professor ID"
// @Param courseID query int false "Course ID"
// @Param userID query int false "User ID (must be self or blank)"
// @Param preload query []string false "Preload (course, professor)"
// @Param populateAgreements query bool false "Populate Agreement Counts"
// @Param populateClientAgreement query bool false "Populate Client's Agreement"
// @Success 200 {array} models.FactrakSurvey
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /factrak/surveys [get]
func (t *Controller) ListSurveys(c *gin.Context) {
	userID := services.GetUserID(c)
	params := models.GetAllFactrakSurveysOptions{}

	err := c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Add the ClientAgreementUserID to the options
	params.ClientAgreementUserID = userID

	// If we don't have the full scope and we aren't passing a user ID (ie self), we don't have the scope so fail.
	// So either we must have full scope or pass the user ID.
	if !auth.HasScope(c, auth.ScopeFactrakFull) && params.UserID == nil {
		t.RespondError(c, lib.ErrorNoScopeAuthorization)
		return
	}

	// Filter what is allowed in options
	// If we pass the userID param, we must either be that user or be admin
	if params.UserID != nil && (*params.UserID != userID && !auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeFactrakAdmin)) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	params.ProfAtWilliams = true

	var surveys []*models.FactrakSurvey
	err = t.surveyModel.GetAllSurveys(&surveys, &params)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, surveys)

	t.RespondOK(c, surveys)
}

// Get one survey
// @Summary Get survey
// @Description get one survey
// @ID getSurvey
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Success 200 {object} models.FactrakSurvey
// @Failure 1330 {object} services.BaseErrorResponse "no scope authorization"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /factrak/surveys/{surveyID} [get]
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

	// If the scope is limited and the scope is not by the user, error so the user doesn't get access to it
	if IsScopeLimited(c) && survey.UserID != services.GetUserID(c) {
		t.RespondError(c, lib.ErrorNoScopeAuthorization)
		return
	}

	err = t.surveyModel.PopulateAgreementCounts(&survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Populate response 2.0
	userID := services.GetUserID(c)
	err = t.surveyModel.PopulateClientAgreement(userID, &survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Remove user info unless self, admin, or admin factrak
	if !auth.CheckIDIsSelf(c, survey.UserID) && !auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeFactrakAdmin) {
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
	CourseNumber            *string `json:"courseNumber"`            // Like "256"

	// Must include this:
	ProfessorID *uint `json:"professorID" binding:"required"`

	// Params:
	Comment              string  `json:"comment" binding:"required"`
	WouldRecommendCourse *bool   `json:"wouldRecommendCourse"`
	CourseWorkload       *int    `json:"courseWorkload" binding:"omitempty,gte=0,lte=7"`
	CourseStimulating    *int    `json:"courseStimulating" binding:"omitempty,gte=0,lte=7"`
	WouldTakeAnother     *bool   `json:"wouldTakeAnother"`
	Approachability      *int    `json:"approachability" binding:"omitempty,gte=0,lte=7"`
	LeadLecture          *int    `json:"leadLecture" binding:"omitempty,gte=0,lte=7"`
	PromoteDiscussion    *int    `json:"promoteDiscussion" binding:"omitempty,gte=0,lte=7"`
	OutsideHelpfulness   *int    `json:"outsideHelpfulness" binding:"omitempty,gte=0,lte=7"`
	MentalHealthSupport  *int    `json:"mentalHealthSupport" binding:"omitempty,gte=0,lte=7"`
	GradeReceived        *string `json:"gradeReceived"`

	SemesterSeason *string `json:"semesterSeason"` // Fall, Winter Study, Spring
	SemesterYear   *int    `json:"semesterYear"`
	CourseFormat   *string `json:"courseFormat"` // Remote, Hybrid, In-Person
}

// @Summary Create survey
// @Description create a survey
// @ID createSurvey
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param createParams body factrak.SurveyCreateParams true "Create Survey Params"
// @Success 201 {object} models.FactrakSurvey
// @Failure 1531 {object} services.BaseErrorResponse "missing course parameters in create data: courseID or (areaOfStudyAbbreviation and courseNumber)"
// @Failure 1532 {object} services.BaseErrorResponse "comment must be 100 characters or more"
// @Failure 1533 {object} services.BaseErrorResponse "user must be a student and could not be found"
// @Failure 1534 {object} services.BaseErrorResponse "passed professor must be a professor and could not be found"
// @Failure 1535 {object} services.BaseErrorResponse "passed course could not be found"
// @Failure 1536 {object} services.BaseErrorResponse "passed area of study could not be found"
// @Failure 1537 {object} services.BaseErrorResponse "survey already exists with passed user ID, professor ID, and course ID"
// @Failure 1101 {object} services.BaseErrorResponse "request data validation failed"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /factrak/surveys [post]
func (t *Controller) CreateSurvey(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := SurveyCreateParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// We have the have either a course id or a area of study and course number
	if createData.CourseID == nil && (createData.AreaOfStudyAbbreviation == nil || createData.CourseNumber == nil) {
		t.RespondError(c, lib.ErrorSurveyMissingCourseParams)
		return
	}

	// Course number & area abbreviation cannot be blank
	if (createData.CourseNumber != nil && createData.AreaOfStudyAbbreviation != nil) &&
		(*createData.CourseNumber == "" || *createData.AreaOfStudyAbbreviation == "") {
		t.RespondError(c, lib.ErrorSurveyMissingCourseParams)
		return
	}

	if len(createData.Comment) < 100 {
		t.RespondError(c, lib.ErrorSurveyCommentTooSmall)
		return
	}

	// Either no semester data or both semester and year must be filled out
	if (createData.SemesterYear != nil && createData.SemesterSeason == nil) || (createData.SemesterYear == nil && createData.SemesterSeason != nil) {
		t.RespondError(c, lib.ErrorSurveyCourseSemesterBad)
		return
	}

	// If semester season isn't one of (fall, winter-study, spring)
	if createData.SemesterSeason != nil &&
		*createData.SemesterSeason != models.FactrakSurveySemesterSeasonFall &&
		*createData.SemesterSeason != models.FactrakSurveySemesterSeasonWinterStudy &&
		*createData.SemesterSeason != models.FactrakSurveySemesterSeasonSpring {

		t.RespondError(c, lib.ErrorSurveyCourseSemesterBad)
		return
	}

	// If semester year is before 2000, error out.
	if createData.SemesterYear != nil && *createData.SemesterYear < 2000 {
		t.RespondError(c, lib.ErrorSurveyCourseSemesterBad)
		return
	}

	// If semester year is in future, error
	if createData.SemesterYear != nil && *createData.SemesterYear > time.Now().Year() {
		t.RespondError(c, lib.ErrorSurveyCourseYearFuture)
		return
	}

	// If course format isn't one of (in-person, hybrid, remote)
	if createData.CourseFormat != nil &&
		*createData.CourseFormat != models.FactrakSurveyCourseFormatInPerson &&
		*createData.CourseFormat != models.FactrakSurveyCourseFormatHybrid &&
		*createData.CourseFormat != models.FactrakSurveyCourseFormatRemote {

		t.RespondError(c, lib.ErrorSurveyCourseFormatBad)
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
		if err = t.areaOfStudyModel.GetAreaOfStudyByAbbreviation(*createData.AreaOfStudyAbbreviation, area); err != nil {
			// Don't return 404; instead, return area of study not found
			if gorm.IsRecordNotFoundError(err) {
				err = lib.ErrorSurveyAreaOfStudyNotFound
			}

			t.RespondError(c, err)
			return
		}

		// "Create" the course, but don't stick it in the db yet.
		course = &models.Course{
			Number:        *createData.CourseNumber,
			AreaOfStudy:   area,
			AreaOfStudyID: &area.ID,
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
	survey := models.FactrakSurvey{
		UserID:      user.ID,
		ProfessorID: prof.ID,
		CourseID:    course.ID,

		// Trim comment of leading/trailing whitespaces
		Comment:              strings.TrimSpace(createData.Comment),
		WouldRecommendCourse: createData.WouldRecommendCourse,
		CourseWorkload:       createData.CourseWorkload,
		CourseStimulating:    createData.CourseStimulating,
		WouldTakeAnother:     createData.WouldTakeAnother,
		Approachability:      createData.Approachability,
		LeadLecture:          createData.LeadLecture,
		PromoteDiscussion:    createData.PromoteDiscussion,
		OutsideHelpfulness:   createData.OutsideHelpfulness,
		MentalHealthSupport:  createData.MentalHealthSupport,
		GradeReceived:        createData.GradeReceived,

		SemesterSeason: createData.SemesterSeason,
		SemesterYear:   createData.SemesterYear,
		CourseFormat:   createData.CourseFormat,

		// Defaults
		TotalAgree:    0,
		TotalDisagree: 0,
	}

	err = t.surveyModel.CreateSurvey(&survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Update the survey deficit. This might be more expensive, as it calculates the net surveys, rather than just
	// taking the current deficit less one, but the more we calculate the net surveys, the more accurate our
	// results should be.
	err = t.studentModel.UpdateFactrakSurveyDeficit(user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We should update the token, as we created a survey
	c.Set(services.UpdateTokenKey, true)

	t.RespondCreated(c, survey)
}

// Rails allows you to change the course, user, and professor of the survey. I don't like that, so you can only change
// survey details here. I am open to the idea of changing courses, though (if for example a user
// put a typo in their course number initially)
type SurveyUpdateParams struct {
	// Params:
	Comment              *string `json:"comment"`
	WouldRecommendCourse *bool   `json:"wouldRecommendCourse"`
	CourseWorkload       *int    `json:"courseWorkload" binding:"omitempty,gte=0,lte=7"`
	CourseStimulating    *int    `json:"courseStimulating" binding:"omitempty,gte=0,lte=7"`
	WouldTakeAnother     *bool   `json:"wouldTakeAnother"`
	Approachability      *int    `json:"approachability" binding:"omitempty,gte=0,lte=7"`
	LeadLecture          *int    `json:"leadLecture" binding:"omitempty,gte=0,lte=7"`
	PromoteDiscussion    *int    `json:"promoteDiscussion" binding:"omitempty,gte=0,lte=7"`
	OutsideHelpfulness   *int    `json:"outsideHelpfulness" binding:"omitempty,gte=0,lte=7"`
	MentalHealthSupport  *int    `json:"mentalHealthSupport" binding:"omitempty,gte=0,lte=7"`
	GradeReceived        *string `json:"gradeReceived"`

	SemesterSeason *string `json:"semesterSeason"` // Fall, Winter Study, Spring
	SemesterYear   *int    `json:"semesterYear"`
	CourseFormat   *string `json:"courseFormat"` // Remote, Hybrid, In-Person
}

// Update survey data
// @Summary Update survey
// @Description update a survey's data
// @ID updateSurvey
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param updateParams body factrak.SurveyUpdateParams true "Update Survey Params"
// @Param surveyID path uint true "Survey ID"
// @Success 200 {object} models.FactrakSurvey
// @Failure 1532 {object} services.BaseErrorResponse "comment must be 100 characters or more"
// @Failure 1101 {object} services.BaseErrorResponse "request data validation failed"
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /factrak/surveys/{surveyID} [patch]
func (t *Controller) UpdateSurvey(c *gin.Context) {
	userID := services.GetUserID(c)

	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Bind update params
	updateData := SurveyUpdateParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Do database query
	var survey models.FactrakSurvey
	err = t.surveyModel.GetSurveyByID(surveyID, &survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Survey must be owned by user id
	if survey.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Validate comment
	if updateData.Comment != nil && len(*updateData.Comment) < 100 {
		t.RespondError(c, lib.ErrorSurveyCommentTooSmall)
		return
	}

	// Either no semester data or both semester and year must be filled out
	if (updateData.SemesterYear != nil && updateData.SemesterSeason == nil) || (updateData.SemesterYear == nil && updateData.SemesterSeason != nil) {
		t.RespondError(c, lib.ErrorSurveyCourseSemesterBad)
		return
	}

	// If semester season isn't one of (fall, winter-study, spring)
	if updateData.SemesterSeason != nil &&
		*updateData.SemesterSeason != models.FactrakSurveySemesterSeasonFall &&
		*updateData.SemesterSeason != models.FactrakSurveySemesterSeasonWinterStudy &&
		*updateData.SemesterSeason != models.FactrakSurveySemesterSeasonSpring {

		t.RespondError(c, lib.ErrorSurveyCourseSemesterBad)
		return
	}

	// If semester year is before 2000, error out.
	if updateData.SemesterYear != nil && *updateData.SemesterYear < 2000 {
		t.RespondError(c, lib.ErrorSurveyCourseSemesterBad)
		return
	}

	// If semester year is in future, error
	if updateData.SemesterYear != nil && *updateData.SemesterYear > time.Now().Year() {
		t.RespondError(c, lib.ErrorSurveyCourseYearFuture)
		return
	}

	// If course format isn't one of (in-person, hybrid, remote)
	if updateData.CourseFormat != nil &&
		*updateData.CourseFormat != models.FactrakSurveyCourseFormatInPerson &&
		*updateData.CourseFormat != models.FactrakSurveyCourseFormatHybrid &&
		*updateData.CourseFormat != models.FactrakSurveyCourseFormatRemote {

		t.RespondError(c, lib.ErrorSurveyCourseFormatBad)
		return
	}

	// Update fields: this is a bit long and verbose, but I don't want to mess with reflect

	// Trim comment of leading/trailing whitespaces
	survey.Comment = strings.TrimSpace(*lib.StrPtrDefaults(updateData.Comment, &survey.Comment))
	survey.WouldRecommendCourse = lib.BoolPtrDefaults(updateData.WouldRecommendCourse, survey.WouldRecommendCourse)
	survey.CourseWorkload = lib.IntPtrDefaults(updateData.CourseWorkload, survey.CourseWorkload)
	survey.CourseStimulating = lib.IntPtrDefaults(updateData.CourseStimulating, survey.CourseStimulating)
	survey.WouldTakeAnother = lib.BoolPtrDefaults(updateData.WouldTakeAnother, survey.WouldTakeAnother)
	survey.Approachability = lib.IntPtrDefaults(updateData.Approachability, survey.Approachability)
	survey.LeadLecture = lib.IntPtrDefaults(updateData.LeadLecture, survey.LeadLecture)
	survey.PromoteDiscussion = lib.IntPtrDefaults(updateData.PromoteDiscussion, survey.PromoteDiscussion)
	survey.OutsideHelpfulness = lib.IntPtrDefaults(updateData.OutsideHelpfulness, survey.OutsideHelpfulness)
	survey.MentalHealthSupport = lib.IntPtrDefaults(updateData.MentalHealthSupport, survey.MentalHealthSupport)
	survey.GradeReceived = lib.StrPtrDefaults(updateData.GradeReceived, survey.GradeReceived)
	survey.SemesterSeason = lib.StrPtrDefaults(updateData.SemesterSeason, survey.SemesterSeason)
	survey.SemesterYear = lib.IntPtrDefaults(updateData.SemesterYear, survey.SemesterYear)
	survey.CourseFormat = lib.StrPtrDefaults(updateData.CourseFormat, survey.CourseFormat)

	// Do db update
	err = t.surveyModel.UpdateSurvey(&survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Populate response
	err = t.surveyModel.PopulateAgreementCounts(&survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We know user is owner, so don't need to delete user fields
	t.RespondOK(c, survey)
}

// Delete survey. Can either do this to self if a user, or to everything if admin
// @Summary Delete survey
// @Description delete a survey
// @ID deleteSurvey
// @Tags factrak,factrak-admin,admin
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Success 200 {object} models.FactrakSurvey
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /factrak/surveys/{surveyID} [delete]
func (t *Controller) DeleteSurvey(c *gin.Context) {
	userID := services.GetUserID(c)

	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var survey models.FactrakSurvey
	err = t.surveyModel.GetSurveyByID(surveyID, &survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Survey must be owned by user id or admin
	if survey.UserID != userID && !IsScopeAdmin(c) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Populate response (must do beforehand, as we then delete these agreements)
	// I chose to include this so the client would know how many agreements they deleted
	err = t.surveyModel.PopulateAgreementCounts(&survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do db delete
	err = t.surveyModel.DeleteSurvey(&survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Update the survey deficit. This might be more expensive, as it calculates the net surveys, rather than just
	// taking the current deficit plus one, but the more we calculate the net surveys, the more accurate our
	// results should be.
	user := new(models.User)
	// We assign the user id to be the owner of the survey, rather than the user who calls the function, as admins may
	// call this function, and we don't want them impacted.
	user.ID = survey.UserID
	err = t.studentModel.UpdateFactrakSurveyDeficit(user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We should update the token, as we deleted a course.
	c.Set(services.UpdateTokenKey, true)

	// We know user is owner, so don't need to delete user fields
	t.RespondOK(c, survey)
}

// Flag survey for mods
// @Summary Flag survey
// @Description flag a survey for mods
// @ID flagSurvey
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Success 200
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /factrak/surveys/{surveyID}/flag [post]
func (t *Controller) FlagSurvey(c *gin.Context) {
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Ensure survey exists
	exists, err := t.surveyModel.DoesSurveyExist(surveyID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Do db flag
	err = t.surveyModel.SetSurveyFlag(surveyID, true)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We know user is owner, so don't need to delete user fields
	t.RespondOK(c, nil)
}
