package dormtrak_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	testify "github.com/stretchr/testify/assert"
)

func TestController_ListDorms(t *testing.T) {
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
	assert.NoError(db.Create(&n1).Create(&n2).Error)

	d1 := models.Dorm{
		Neighborhood: &n1,
		Name:         "East",
	}
	d2 := models.Dorm{
		Neighborhood: &n1,
		Name:         "Faye",
	}
	d3 := models.Dorm{
		Neighborhood: &n2,
		Name:         "Hubble",
	}
	assert.NoError(db.Create(&d1).Create(&d2).Create(&d3).Error)

	// Get test dorms
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/dorms", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Dorm
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct neighborhoods
	assert.Len(resp, 3)
	assert.Equal(d1.Name, resp[0].Name)
	assert.Equal(d2.Name, resp[1].Name)
	assert.Equal(d3.Name, resp[2].Name)
}

func TestController_GetDorm(t *testing.T) {
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
	assert.NoError(db.Create(&n1).Create(&n2).Error)

	d1 := models.Dorm{
		Neighborhood: &n1,
		Name:         "East",
		DormRooms: []*models.DormRoom{
			{
				Number: "102",
			},
			{
				Number: "314",
			},
		},
	}
	d2 := models.Dorm{
		Neighborhood: &n1,
		Name:         "Faye",
	}
	d3 := models.Dorm{
		Neighborhood: &n2,
		Name:         "Hubble",
	}
	assert.NoError(db.Create(&d1).Create(&d2).Create(&d3).Error)

	// Get test neighborhood
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/dorms/%d", d1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Dorm
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct dorm
	assert.Equal(d1.Name, resp.Name)
	assert.Equal(d1.Neighborhood.Name, resp.Neighborhood.Name)
	assert.Len(resp.DormRooms, 2)
	assert.Equal(d1.DormRooms[0].Number, resp.DormRooms[0].Number)
	assert.Equal(d1.DormRooms[1].Number, resp.DormRooms[1].Number)

	/* Get test bad dorm id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/dorms/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_GetDormRooms(t *testing.T) {
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
	assert.NoError(db.Create(&n1).Create(&n2).Error)

	d1 := models.Dorm{
		Neighborhood: &n1,
		Name:         "East",
		DormRooms: []*models.DormRoom{
			{
				Number: "102",
			},
			{
				Number: "314",
			},
			{
				Number: "109",
			},
		},
	}
	d2 := models.Dorm{
		Neighborhood: &n2,
		Name:         "Hubble",
		DormRooms: []*models.DormRoom{
			{
				Number: "102",
			},
		},
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	// Get test neighborhood
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/dorms/%d/rooms", d1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.DormRoom
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct dorm rooms
	assert.Len(resp, 3)
	assert.Equal(d1.DormRooms[0].Number, resp[0].Number)
	assert.Equal(d1.DormRooms[1].Number, resp[1].Number)
	assert.Equal(d1.DormRooms[2].Number, resp[2].Number)

	/* Get test bad dorm id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/dorms/%d/rooms", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}
