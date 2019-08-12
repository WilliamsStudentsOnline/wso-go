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
)

func TestController_GetRankings(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	neighborhood := &models.Neighborhood{
		Name: "Currier",
	}
	assert.NoError(db.Create(neighborhood).Error)

	dorms := []*models.Dorm{
		// Max mean single size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeSingle,
					Area:     lib.IntToPtr(300),
				},
				{
					RoomType: models.DormRoomTypeSingle,
					Area:     lib.IntToPtr(250),
				},
			},
		},
		// Min mean single size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeSingle,
					Area:     lib.IntToPtr(30),
				},
				{
					RoomType: models.DormRoomTypeSingle,
					Area:     lib.IntToPtr(50),
				},
			},
		},
		// Max mean dbl size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(600),
				},
			},
		},
		// Min mean dbl size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(150),
				},
			},
		},
		// Most singles
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeSingle,
					Area:     lib.IntToPtr(150),
				},
				{
					RoomType: models.DormRoomTypeSingle,
					Area:     lib.IntToPtr(130),
				},
				{
					RoomType: models.DormRoomTypeSingle,
					Area:     lib.IntToPtr(110),
				},
			},
		},
		// Most doubles
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(3),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(190),
				},
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(210),
				},
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(200),
				},
			},
		},
		// Best bathroom ratio
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(5),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(195),
				},
			},
		},
		// Worst ratio
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(1),
			DormRooms: []*models.DormRoom{
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(195),
				},
				{
					RoomType: models.DormRoomTypeDouble,
					Area:     lib.IntToPtr(200),
				},
			},
		},
	}

	for i := range dorms {
		dorms[i].Name = "Dorm " + strconv.Itoa(i)
		for j := range dorms[i].DormRooms {
			dorms[i].DormRooms[j].Number = strconv.Itoa(i) + "_" + strconv.Itoa(j)
		}
		assert.NoError(db.Create(&dorms[i]).Error)
	}

	// Get test rankings
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/rankings"), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.DormtrakRanking
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check length
	assert.Len(resp.MaxMeanSingleSize, 3)
	assert.Len(resp.MinMeanSingleSize, 3)
	assert.Len(resp.BiggestSingles, 3)
	assert.Len(resp.SmallestSingles, 3)
	assert.Len(resp.MaxMeanDoubleSize, 3)
	assert.Len(resp.MinMeanDoubleSize, 3)
	assert.Len(resp.BiggestDoubles, 3)
	assert.Len(resp.SmallestDoubles, 3)
	assert.Len(resp.MostSingles, 3)
	assert.Len(resp.MostDoubles, 3)
	assert.Len(resp.MostBathrooms, 3)
	assert.Len(resp.FewestBathrooms, 3)

	// Test by ID
	assert.Equal(dorms[0].ID, resp.MaxMeanSingleSize[0].ID)
	assert.Equal(dorms[1].ID, resp.MinMeanSingleSize[0].ID)
	assert.Equal(dorms[0].DormRooms[0].ID, resp.BiggestSingles[0].ID)
	assert.Equal(dorms[1].DormRooms[0].ID, resp.SmallestSingles[0].ID)
	assert.Equal(dorms[2].ID, resp.MaxMeanDoubleSize[0].ID)
	assert.Equal(dorms[3].ID, resp.MinMeanDoubleSize[0].ID)
	assert.Equal(dorms[2].DormRooms[0].ID, resp.BiggestDoubles[0].ID)
	assert.Equal(dorms[3].DormRooms[0].ID, resp.SmallestDoubles[0].ID)
	assert.Equal(dorms[4].ID, resp.MostSingles[0].ID)
	assert.Equal(dorms[5].ID, resp.MostDoubles[0].ID)
	assert.Equal(dorms[6].ID, resp.MostBathrooms[0].ID)
	assert.Equal(dorms[7].ID, resp.FewestBathrooms[0].ID)
}
