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
	t.SkipNow()
}

func TestDormModel_GetDormFacts(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewDormModel(db)

	// Create dorm
	dorm := Dorm{
		Neighborhood: &Neighborhood{
			Name: "Currier",
		},
		Name:      "East",
		KeyOrCard: lib.StrToPtr("key"),
	}
	assert.NoError(db.Create(&dorm).Error)

	// Test 0: Missing dorm facts
	res := NewDormFacts()
	assert.NoError(m.GetDormFacts(dorm.ID, res))
	assert.Equal(0, res.SinglesCount)
	assert.Equal(0, res.DoublesCount)
	assert.Equal(0, res.FlexCount)
	assert.Equal(0, res.Capacity)
	assert.Equal(*dorm.KeyOrCard, *res.KeyOrCard)
	assert.Nil(res.AverageSinglesArea)
	assert.Equal(0, res.SophomoreCount)
	assert.Equal(0, res.JuniorCount)
	assert.Equal(0, res.SeniorCount)
	assert.Nil(res.CommonRoomAccessRatio)
	assert.Nil(res.AverageWifi)
	assert.Nil(res.AverageLocation)
	assert.Nil(res.AverageLoudness)
	assert.Nil(res.AverageSatisfaction)

	srYear := NewStudentModel(db).SeniorYear()

	// Test 1
	// Add one dorm room
	t1 := []*DormRoom{
		{
			Dorm:             &dorm,
			Number:           "a",
			RoomType:         DormRoomTypeSingle,
			CommonRoomAccess: lib.BoolToPtr(true),
			Area:             lib.IntToPtr(130),
			Wifi:             lib.Float64ToPtr(5),
			Location:         lib.Float64ToPtr(4),
			Loudness:         lib.Float64ToPtr(3),
			Satisfaction:     lib.Float64ToPtr(2),
			Users: []*User{
				{
					UnixID:    "a_u1",
					Name:      "a User 1",
					Type:      UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 2), // Sophomore
				},
			},
		},
	}
	for i := range t1 {
		assert.NoError(db.Create(&t1[i]).Error)
	}

	res = NewDormFacts()
	assert.NoError(m.GetDormFacts(dorm.ID, res))
	assert.Equal(1, res.SinglesCount)
	assert.Equal(0, res.DoublesCount)
	assert.Equal(0, res.FlexCount)
	assert.Equal(1, res.Capacity)
	assert.Equal(*dorm.KeyOrCard, *res.KeyOrCard)
	assert.Equal(130, *res.AverageSinglesArea)
	assert.Equal(1, res.SophomoreCount)
	assert.Equal(0, res.JuniorCount)
	assert.Equal(0, res.SeniorCount)
	assert.Equal(1.0, *res.CommonRoomAccessRatio)
	assert.Equal(5.0, *res.AverageWifi)
	assert.Equal(4.0, *res.AverageLocation)
	assert.Equal(3.0, *res.AverageLoudness)
	assert.Equal(2.0, *res.AverageSatisfaction)

	// Soft delete the previous dorm room(s)
	for i := range t1 {
		assert.NoError(db.Delete(&t1[i]).Error)
	}

	// Test 2
	// Add a bunch more dorm rooms and test them out
	t2 := []*DormRoom{
		{
			Dorm:             &dorm,
			RoomType:         DormRoomTypeSingle,
			CommonRoomAccess: lib.BoolToPtr(true),
			Area:             lib.IntToPtr(100),
			Wifi:             lib.Float64ToPtr(1),
			Location:         lib.Float64ToPtr(2),
			Loudness:         lib.Float64ToPtr(7),
			Satisfaction:     lib.Float64ToPtr(1),
			Users: []*User{
				{
					Type:      UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 2), // Sophomore
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         DormRoomTypeSingle,
			CommonRoomAccess: lib.BoolToPtr(true),
			Area:             lib.IntToPtr(80),
			Wifi:             lib.Float64ToPtr(3),
			Location:         lib.Float64ToPtr(4),
			Loudness:         lib.Float64ToPtr(4),
			Satisfaction:     lib.Float64ToPtr(2),
			Users: []*User{
				{
					Type:      UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 0), // senior
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         DormRoomTypeSingle,
			CommonRoomAccess: lib.BoolToPtr(false),
			Area:             lib.IntToPtr(80),
			Wifi:             lib.Float64ToPtr(5),
			Location:         lib.Float64ToPtr(6),
			Loudness:         lib.Float64ToPtr(3),
			Satisfaction:     lib.Float64ToPtr(3),
			Users: []*User{
				{
					Type:      UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 1), // Junior
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         DormRoomTypeDouble,
			CommonRoomAccess: lib.BoolToPtr(true),
			Area:             lib.IntToPtr(250),
			Wifi:             lib.Float64ToPtr(7),
			Location:         lib.Float64ToPtr(7),
			Loudness:         lib.Float64ToPtr(6),
			Satisfaction:     lib.Float64ToPtr(1),
			Users: []*User{
				{
					Type:      UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 2), // Sophomore
				},
				{
					Type:      UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 1), // Junior
				},
			},
		},
		{
			Dorm:             &dorm,
			RoomType:         DormRoomTypeFlex,
			CommonRoomAccess: lib.BoolToPtr(false),
			Area:             lib.IntToPtr(190),
			Wifi:             lib.Float64ToPtr(2),
			Location:         lib.Float64ToPtr(5),
			Loudness:         lib.Float64ToPtr(1),
			Satisfaction:     lib.Float64ToPtr(7),
			Users: []*User{
				{
					Type:      UserTypeStudent,
					ClassYear: lib.IntToPtr(srYear + 2), // Sophomore
				},
			},
		},
	}
	for i := range t2 {
		t2[i].Number = strconv.Itoa(i)
		for j := range t2[i].Users {
			t2[i].Users[j].UnixID = strconv.Itoa(i) + "_" + strconv.Itoa(j)
			t2[i].Users[j].Name = strconv.Itoa(i) + " User " + strconv.Itoa(j)
		}
		assert.NoError(db.Create(&t2[i]).Error)
	}

	res = NewDormFacts()
	assert.NoError(m.GetDormFacts(dorm.ID, res))
	assert.Equal(3, res.SinglesCount)
	assert.Equal(1, res.DoublesCount)
	assert.Equal(1, res.FlexCount)
	assert.Equal(7, res.Capacity)
	assert.Equal(*dorm.KeyOrCard, *res.KeyOrCard)
	assert.Equal(87, *res.AverageSinglesArea)
	assert.Equal(80, *res.ModeSinglesArea)
	assert.Equal(3, res.SophomoreCount)
	assert.Equal(2, res.JuniorCount)
	assert.Equal(1, res.SeniorCount)
	assert.Equal(3.0/5.0, *res.CommonRoomAccessRatio)
	assert.Equal(float64(2+7+5+3+1)/5.0, *res.AverageWifi)
	assert.Equal(4.8, *res.AverageLocation)
	assert.Equal(4.2, *res.AverageLoudness)
	assert.Equal(2.8, *res.AverageSatisfaction)
}
