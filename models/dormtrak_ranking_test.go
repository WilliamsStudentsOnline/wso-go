package models_test

import (
	"strconv"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestDormModel_GetDormtrakRankings(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewDormModel(db, zaptest.NewLogger(t).Sugar())

	// Test 0: Missing dorm ranking
	res := NewDormtrakRanking()
	assert.NoError(m.GetDormtrakRankings(3, res))

	assert.Empty(res.MaxMeanSingleSize)
	assert.Empty(res.MinMeanSingleSize)
	assert.Empty(res.BiggestSingles)
	assert.Empty(res.SmallestSingles)
	assert.Empty(res.MaxMeanDoubleSize)
	assert.Empty(res.MinMeanDoubleSize)
	assert.Empty(res.BiggestDoubles)
	assert.Empty(res.SmallestDoubles)
	assert.Empty(res.MostSingles)
	assert.Empty(res.MostDoubles)
	assert.Empty(res.MostBathrooms)
	assert.Empty(res.FewestBathrooms)

	neighborhood := &Neighborhood{
		Name: "Currier",
	}
	assert.NoError(db.Create(neighborhood).Error)

	dorms := []*Dorm{
		// Max mean single size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeSingle,
					Area:     lib.IntToPtr(300),
				},
				{
					RoomType: DormRoomTypeSingle,
					Area:     lib.IntToPtr(250),
				},
			},
		},
		// Min mean single size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeSingle,
					Area:     lib.IntToPtr(30),
				},
				{
					RoomType: DormRoomTypeSingle,
					Area:     lib.IntToPtr(50),
				},
			},
		},
		// Max mean dbl size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeDouble,
					Area:     lib.IntToPtr(600),
				},
			},
		},
		// Min mean dbl size
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeDouble,
					Area:     lib.IntToPtr(150),
				},
			},
		},
		// Most singles
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(2),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeSingle,
					Area:     lib.IntToPtr(150),
				},
				{
					RoomType: DormRoomTypeSingle,
					Area:     lib.IntToPtr(130),
				},
				{
					RoomType: DormRoomTypeSingle,
					Area:     lib.IntToPtr(110),
				},
			},
		},
		// Most doubles
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(3),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeDouble,
					Area:     lib.IntToPtr(190),
				},
				{
					RoomType: DormRoomTypeDouble,
					Area:     lib.IntToPtr(210),
				},
				{
					RoomType: DormRoomTypeDouble,
					Area:     lib.IntToPtr(200),
				},
			},
		},
		// Best bathroom ratio
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(5),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeDouble,
					Area:     lib.IntToPtr(195),
				},
			},
		},
		// Worst ratio
		{
			Neighborhood:    neighborhood,
			NumberBathrooms: lib.IntToPtr(1),
			DormRooms: []*DormRoom{
				{
					RoomType: DormRoomTypeDouble,
					Area:     lib.IntToPtr(195),
				},
				{
					RoomType: DormRoomTypeDouble,
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

	// Test 1: Dorm ranking
	res = NewDormtrakRanking()
	assert.NoError(m.GetDormtrakRankings(3, res))

	// Check length
	assert.Len(res.MaxMeanSingleSize, 3)
	assert.Len(res.MinMeanSingleSize, 3)
	assert.Len(res.BiggestSingles, 3)
	assert.Len(res.SmallestSingles, 3)
	assert.Len(res.MaxMeanDoubleSize, 3)
	assert.Len(res.MinMeanDoubleSize, 3)
	assert.Len(res.BiggestDoubles, 3)
	assert.Len(res.SmallestDoubles, 3)
	assert.Len(res.MostSingles, 3)
	assert.Len(res.MostDoubles, 3)
	assert.Len(res.MostBathrooms, 3)
	assert.Len(res.FewestBathrooms, 3)

	// Test by ID
	assert.Equal(dorms[0].ID, res.MaxMeanSingleSize[0].ID)
	assert.Equal(dorms[1].ID, res.MinMeanSingleSize[0].ID)
	assert.Equal(dorms[0].DormRooms[0].ID, res.BiggestSingles[0].ID)
	assert.Equal(dorms[1].DormRooms[0].ID, res.SmallestSingles[0].ID)
	assert.Equal(dorms[2].ID, res.MaxMeanDoubleSize[0].ID)
	assert.Equal(dorms[3].ID, res.MinMeanDoubleSize[0].ID)
	assert.Equal(dorms[2].DormRooms[0].ID, res.BiggestDoubles[0].ID)
	assert.Equal(dorms[3].DormRooms[0].ID, res.SmallestDoubles[0].ID)
	assert.Equal(dorms[4].ID, res.MostSingles[0].ID)
	assert.Equal(dorms[5].ID, res.MostDoubles[0].ID)
	assert.Equal(dorms[6].ID, res.MostBathrooms[0].ID)
	assert.Equal(dorms[7].ID, res.FewestBathrooms[0].ID)
}
