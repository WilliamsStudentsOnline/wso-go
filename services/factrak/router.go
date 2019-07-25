package factrak

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)

	// Professors Endpoint
	r.GET("/professors", c.ListProfessors) // List professors
	// For the following requests, you may specify a "?courseID=XXX" to scope your request to a specific course:
	r.GET("/professors/:professorID", c.GetProfessor)                 // Get specific prof; give it the statistics
	r.GET("/professors/:professorID/surveys", c.ListProfessorSurveys) // Get surveys for a professor. just reuse inner methods of /surveys; include agreements
	r.GET("/professors/:professorID/ratings", c.GetProfessorRatings)  // Get professor ratings
	// Not this one, though:
	r.GET("/professors/:professorID/courses", c.ListProfessorCourses) // List prof's courses

	// Users Endpoint
	// SCOPE: self, admin, factrak_admin
	r.GET("/users/:userID/surveys", c.ListUserSurveys) // List user (students/alum) surveys

	// Courses Endpoint
	r.GET("/courses", c.ListCourses) // List courses
	// For the following requests, you may specify a "?courseID=XXX" to scope your request to a specific course:
	r.GET("/courses/:courseID", c.GetCourse)                 // Get specific course; give it the statistics
	r.GET("/courses/:courseID/surveys", c.ListCourseSurveys) // Get surveys for a course; include agreements
	r.GET("/courses/:courseID/ratings", c.GetCourseRatings)  // Get course ratings
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

	r.GET("/surveys", c.ListSurveys)                      // List surveys
	r.GET("/surveys/:surveyID", c.GetSurvey)            // Get specific one (have agreements as a count)

	// Current workflow is to post data to survey (where it creates a course if necessary).
	// TODO: Could also do this (which one is better?):
	// GET/POST to course to get a valid course
	// POST survey model to surveys
	r.POST("/surveys", c.CreateSurvey)                     // Create

	r.PUT("/surveys/:surveyID")            // Edit
	r.DELETE("/surveys/:surveyID")         // Delete
	r.GET("/surveys/:surveyID/agreements") // Get survey agreements
	r.POST("/surveys/:surveyID/flag")      // Flag survey
	r.DELETE("/surveys/:surveyID/flag")    // Delete survey flags

	r.GET("/agreements/:agreementID")    // Get agreement
	r.POST("/agreements")                // Add agreement
	r.PUT("/agreements/:agreementID")    // Edit agreemenr
	r.DELETE("/agreements/:agreementID") // Delete agreement
}
