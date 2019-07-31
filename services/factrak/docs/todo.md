# TODO For PR

* Add scope tests
  * Admin scope required for `GET /admin/surveys`, `DELETE /admin/surveys/:surveyID/flag`, 
  can do all `DELETE /surveys/:surveyID`, gets all user IDs
  * Limited cannot access professor surveys, professor ratings, course surveys, course ratings, all surveys, 
  get survey that is not owned by me, flag survey, get/create/update/delete agreement
* Add model tests/documentation
  * Area of study
  * Course
  * Department
  * Factrak agreement
  * Factrak survey
  * Professor
  * User
  * Student
* Add auth tests
  * LDAP?
  * Authenticator
  * GenerateClaims
  * UpdateClaims
* Add services tests/documentation: RespondErrorCode