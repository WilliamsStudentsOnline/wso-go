package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// Get all professors
func (t *Controller) FetchAllProfessors(c *gin.Context) {
	var profs []models.User
	err := t.professorModel.GetAllProfessors(&profs)

	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	t.RespondOK(c, profs)
}

// Get one professor
func (t *Controller) GetProfessor(c *gin.Context) {
	// Decode userID or self.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondError(c, http.StatusBadRequest, err)
	}

	// Do database query
	var prof models.User
	err = t.professorModel.GetProfessorByID(profID, &prof)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	if !prof.AtWilliams {
		t.RespondAPIError(c, lib.ErrorUserNotAtWilliams)
		return
	}

	t.RespondOK(c, prof)
}
