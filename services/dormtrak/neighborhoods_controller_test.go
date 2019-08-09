package dormtrak_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	testify "github.com/stretchr/testify/assert"
)

func TestController_ListNeighborhoods(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	SetupRouter(router, db)

	n1 := models.Neighborhood{
		Name: "Currier",
	}
	n2 := models.Neighborhood{
		Name: "Dodd",
	}
	n3 := models.Neighborhood{
		Name:    "First-year",
		Trakked: lib.BoolToPtr(false),
	}
	assert.NoError(db.Create(&n1).Create(&n2).Create(&n3).Error)

	// Get test neighborhood
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/neighborhoods", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Neighborhood
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct neighborhoods
	assert.Len(resp, 3)
	assert.Equal(n1.Name, resp[0].Name)
	assert.True(*n1.Trakked)
	assert.Equal(n2.Name, resp[1].Name)
	assert.Equal(n3.Name, resp[2].Name)
	assert.False(*n3.Trakked)
}

func TestController_GetNeighborhood(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	SetupRouter(router, db)

	n1 := models.Neighborhood{
		Name: "Currier",
		Dorms: []*models.Dorm{
			{
				Name: "East",
			},
			{
				Name: "Currier",
			},
		},
	}
	n2 := models.Neighborhood{
		Name: "Dodd",
	}
	assert.NoError(db.Create(&n1).Create(&n2).Error)

	// Get test neighborhood
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/neighborhoods/%d", n1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Neighborhood
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct neighborhood
	assert.Equal(n1.Name, resp.Name)
	assert.Len(resp.Dorms, 2)
	assert.Equal(n1.Dorms[0].Name, resp.Dorms[0].Name)
	assert.Equal(n1.Dorms[1].Name, resp.Dorms[1].Name)

	/* Get test bad neighborhood id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/neighborhoods/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}
