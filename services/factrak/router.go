package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)

	// Limited factrak scoping is the default r

	// Full factrak scoping:
	full := r.Group("")
	full.Use(auth.RequireScopes(auth.ScopeFactrakFull))

	// Professors Endpoint
	r.GET("/professors", c.ListProfessors) // List professors
	// For the following requests, you may specify a "?courseID=XXX" to scope your request to a specific course:
	r.GET("/professors/:professorID", c.GetProfessor)                    // Get specific prof
	full.GET("/professors/:professorID/surveys", c.ListProfessorSurveys) // Get surveys for a professor. just reuse inner methods of /surveys; include agreements
	full.GET("/professors/:professorID/ratings", c.GetProfessorRatings)  // Get professor ratings
	// Not this one, though:
	r.GET("/professors/:professorID/courses", c.ListProfessorCourses) // List prof's courses

	// Users Endpoint
	// SCOPE: self, admin, factrak_admin
	r.GET("/users/:userID/surveys", c.ListUserSurveys) // List user (students/alum) surveys

	// Courses Endpoint
	r.GET("/courses", c.ListCourses) // List courses
	// For the following requests, you may specify a "?courseID=XXX" to scope your request to a specific course:
	r.GET("/courses/:courseID", c.GetCourse)                    // Get specific course; give it the statistics
	full.GET("/courses/:courseID/surveys", c.ListCourseSurveys) // Get surveys for a course; include agreements
	full.GET("/courses/:courseID/ratings", c.GetCourseRatings)  // Get course ratings
	// Not this one, though:
	r.GET("/courses/:courseID/professors", c.ListCourseProfessors) // Get professors for a course

	r.GET("/departments", c.ListDepartments)                                   // List departments
	r.GET("/departments/:departmentID", c.GetDepartment)                       // Get specific department
	r.GET("/departments/:departmentID/professors", c.ListDepartmentProfessors) // Get department's professors
	r.GET("/departments/:departmentID/courses", c.ListDepartmentCourses)       // Get department's courses

	r.GET("/areas-of-study", c.ListAreasOfStudy)                                    // List areas
	r.GET("/areas-of-study/:areaOfStudyID", c.GetAreaOfStudy)                       // Get area
	r.GET("/areas-of-study/:areaOfStudyID/professors", c.ListAreaOfStudyProfessors) // Get area's professors
	r.GET("/areas-of-study/:areaOfStudyID/courses", c.ListAreaOfStudyCourses)       // Get area's courses

	full.GET("/surveys", c.ListSurveys)      // List surveys
	r.GET("/surveys/:surveyID", c.GetSurvey) // Get specific one (have agreements as a count)

	// Current workflow is to post data to survey (where it creates a course if necessary).
	// TODO: Could also do this (which one is better?):
	// GET/POST to course to get a valid course
	// POST survey model to surveys
	r.POST("/surveys", c.CreateSurvey) // Create

	r.PATCH("/surveys/:surveyID", c.UpdateSurvey)  // Edit
	r.DELETE("/surveys/:surveyID", c.DeleteSurvey) // Delete

	full.POST("/surveys/:surveyID/flag", c.FlagSurvey) // Flag survey

	// put these in factrak admin endpoint:
	admin := r.Group("")
	admin.Use(auth.RequireScopes(auth.ScopeFactrakAdmin, auth.ScopeAdminAll))
	admin.GET("/flagged_surveys")           // Get flagged surveys
	admin.DELETE("/surveys/:surveyID/flag") // Delete survey flags: require admin

	full.GET("/surveys/:surveyID/agreements") // Get survey agreements

	full.GET("/agreements/:agreementID")    // Get agreement
	full.POST("/agreements")                // Add agreement
	full.PUT("/agreements/:agreementID")    // Edit agreement
	full.DELETE("/agreements/:agreementID") // Delete agreement
}
