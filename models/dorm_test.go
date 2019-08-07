package models_test

import (
	"strconv"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
)

// This tests UpdateDormFacts and Dorm.BeforeSave()
func TestDormModel_UpdateDormFacts(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Create dorm
	dorm := Dorm{
		Neighborhood: &Neighborhood{
			Name: "Currier",
		},
		Name: "East",
	}
	assert.NoError(db.Create(&dorm).Error)

	// Test 0
	// Get dorm and make sure all fields are zeroed
	var res Dorm
	assert.NoError(db.First(&res, dorm.ID).Error)
	assert.Nil(res.AverageSingleArea)
	assert.Nil(res.AverageDoubleArea)
	assert.Equal(0, *res.NumberSingles)
	assert.Equal(0, *res.NumberDoubles)
	assert.Equal(0, *res.NumberFlex)
	assert.Nil(res.ModeSingleArea)
	assert.Nil(res.ModeDoubleArea)

	// Test 1
	// Add one dorm room
	t1 := []*DormRoom{
		{
			Dorm:     &dorm,
			Number:   "a",
			RoomType: DormRoomTypeSingle,
			Area:     lib.IntToPtr(130),
		},
	}
	for i := range t1 {
		assert.NoError(db.Create(&t1[i]).Error)
	}

	res = Dorm{}
	assert.NoError(db.First(&res, dorm.ID).Error)
	assert.Equal(130, *res.AverageSingleArea)
	assert.Nil(res.AverageDoubleArea)
	assert.Equal(1, *res.NumberSingles)
	assert.Equal(0, *res.NumberDoubles)
	assert.Equal(0, *res.NumberFlex)
	assert.Equal(130, *res.ModeSingleArea)
	assert.Nil(res.ModeDoubleArea)
	// Test auto-generated values too
	assert.Equal(1, *res.Capacity)

	// Soft delete the previous dorm room(s)
	for i := range t1 {
		assert.NoError(db.Delete(&t1[i]).Error)
	}

	// Test 2
	// Add a bunch more dorm rooms and test them out
	t2 := []*DormRoom{
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeSingle,
			Area:     lib.IntToPtr(100),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeSingle,
			Area:     lib.IntToPtr(80),
		},

		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeSingle,
			Area:     lib.IntToPtr(75),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeSingle,
			Area:     lib.IntToPtr(80),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeDouble,
			Area:     lib.IntToPtr(250),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeDouble,
			Area:     lib.IntToPtr(215),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeDouble,
			Area:     lib.IntToPtr(216),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeDouble,
			Area:     lib.IntToPtr(250),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeFlex,
			Area:     lib.IntToPtr(190),
		},
		{
			Dorm:     &dorm,
			RoomType: DormRoomTypeFlex,
			Area:     lib.IntToPtr(195),
		},
	}
	for i := range t2 {
		t2[i].Number = strconv.Itoa(i)
		assert.NoError(db.Create(&t2[i]).Error)
	}

	res = Dorm{}
	assert.NoError(db.First(&res, dorm.ID).Error)
	assert.Equal(84, *res.AverageSingleArea)
	assert.Equal(233, *res.AverageDoubleArea)
	assert.Equal(4, *res.NumberSingles)
	assert.Equal(4, *res.NumberDoubles)
	assert.Equal(2, *res.NumberFlex)
	assert.Equal(80, *res.ModeSingleArea)
	assert.Equal(250, *res.ModeDoubleArea)
	// Test auto-generated values too
	assert.Equal(16, *res.Capacity)
}

func TestDorm_BeforeSave(t *testing.T) {

}
