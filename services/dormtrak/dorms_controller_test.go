package dormtrak_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestController_ListDorms(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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

func TestController_ListDormRooms(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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

func TestController_GetDormFacts(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	dorm := models.Dorm{
		Neighborhood: &models.Neighborhood{
			Name: "Currier",
		},
		Name:      "East",
		KeyOrCard: lib.StrToPtr("key"),
	}
	// User
	u1 := models.User{
		Name:   "User 1",
		UnixID: "u1",
	}
	assert.NoError(db.Create(&dorm).Create(&u1).Error)

	srYear := models.NewStudentModel(db, zap.S()).SeniorYear()

	rooms := []*models.DormRoom{
		{
			Dorm:             &dorm,
			RoomType:         models.DormRoomTypeSingle,
			CommonRoomAccess: lib.BoolToPtr(true),
			Area:             lib.IntToPtr(100),
			DormtrakReviews: []*models.DormtrakReview{
				{
					User:         &u1,
					Wifi:         lib.IntToPtr(1),
					Location:     lib.IntToPtr(2),
					Loudness:     lib.IntToPtr(7),
					Satisfaction: lib.IntToPtr(1),
				},
			},
			Users: []*models.User{
				{
					Type:      models.UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 2), // Sophomore
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         models.DormRoomTypeSingle,
			CommonRoomAccess: lib.BoolToPtr(true),
			Area:             lib.IntToPtr(80),
			DormtrakReviews: []*models.DormtrakReview{
				{
					User:         &u1,
					Wifi:         lib.IntToPtr(3),
					Location:     lib.IntToPtr(4),
					Loudness:     lib.IntToPtr(4),
					Satisfaction: lib.IntToPtr(2),
				},
			},
			Users: []*models.User{
				{
					Type:      models.UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 0), // senior
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         models.DormRoomTypeSingle,
			CommonRoomAccess: lib.BoolToPtr(false),
			Area:             lib.IntToPtr(80),
			DormtrakReviews: []*models.DormtrakReview{
				{
					User:         &u1,
					Wifi:         lib.IntToPtr(5),
					Location:     lib.IntToPtr(6),
					Loudness:     lib.IntToPtr(3),
					Satisfaction: lib.IntToPtr(3),
				},
			},
			Users: []*models.User{
				{
					Type:      models.UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 1), // Junior
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         models.DormRoomTypeDouble,
			CommonRoomAccess: lib.BoolToPtr(true),
			Area:             lib.IntToPtr(250),
			DormtrakReviews: []*models.DormtrakReview{
				{
					User:         &u1,
					Wifi:         lib.IntToPtr(7),
					Location:     lib.IntToPtr(7),
					Loudness:     lib.IntToPtr(6),
					Satisfaction: lib.IntToPtr(1),
				},
			},
			Users: []*models.User{
				{
					Type:      models.UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 2), // Sophomore
				},
				{
					Type:      models.UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 1), // Junior
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         models.DormRoomTypeFlex,
			CommonRoomAccess: lib.BoolToPtr(false),
			Area:             lib.IntToPtr(190),
			DormtrakReviews: []*models.DormtrakReview{
				{
					User:         &u1,
					Wifi:         lib.IntToPtr(2),
					Location:     lib.IntToPtr(5),
					Loudness:     lib.IntToPtr(1),
					Satisfaction: lib.IntToPtr(7),
				},
			},
			Users: []*models.User{
				{
					Type:      models.UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 2), // Sophomore
				},
			},
		},
	}

	for i := range rooms {
		rooms[i].Number = strconv.Itoa(i)
		for j := range rooms[i].Users {
			rooms[i].Users[j].UnixID = strconv.Itoa(i) + "_" + strconv.Itoa(j)
			rooms[i].Users[j].Name = strconv.Itoa(i) + " User " + strconv.Itoa(j)
		}
		assert.NoError(db.Create(&rooms[i]).Error)
	}

	// Get test dorm
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/dorms/%d/facts", dorm.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.DormFacts
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct dorm facts
	assert.Equal(3, resp.SinglesCount)
	assert.Equal(1, resp.DoublesCount)
	assert.Equal(1, resp.FlexCount)
	assert.Equal(7, resp.Capacity)
	assert.Equal(*dorm.KeyOrCard, *resp.KeyOrCard)
	assert.Equal(87, *resp.AverageSinglesArea)
	assert.Equal(80, *resp.ModeSinglesArea)
	assert.Equal(3, resp.SophomoreCount)
	assert.Equal(2, resp.JuniorCount)
	assert.Equal(1, resp.SeniorCount)
	assert.Equal(3.0/5.0, *resp.CommonRoomAccessRatio)
	assert.Equal(float64(2+7+5+3+1)/5.0, *resp.AverageWifi)
	assert.Equal(4.8, *resp.AverageLocation)
	assert.Equal(4.2, *resp.AverageLoudness)
	assert.Equal(2.8, *resp.AverageSatisfaction)

	/* Get test bad dorm id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/dorms/%d/facts", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_GetDormRoom(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	n1 := models.Neighborhood{
		Name: "Currier",
	}
	assert.NoError(db.Create(&n1).Error)

	dr1 := models.DormRoom{
		Number: "102",
		Dorm: &models.Dorm{
			Neighborhood: &n1,
			Name:         "East",
		},
	}
	assert.NoError(db.Create(&dr1).Error)

	// Get test neighborhood
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/rooms/%d", dr1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.DormRoom
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct dorm rooms
	assert.Equal(dr1.Number, resp.Number)
	assert.Equal(dr1.Dorm.Name, resp.Dorm.Name)
	assert.Equal(dr1.Dorm.Neighborhood.Name, resp.Dorm.Neighborhood.Name)

	/* Get test bad dorm id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/rooms/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}
