package models

import (
	"math"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
)

// Dorm Model
type DormModel struct {
	*BaseModel
}

func NewDormModel(db *gorm.DB) *DormModel {
	return &DormModel{
		BaseModel: NewBaseModel(db),
	}
}

// Gets all dorms.
func (m *DormModel) GetAllDorms(p *[]*Dorm) (err error) {
	err = m.DB.Find(p).Error
	return
}

// Gets dorm by its id with neighborhood and dorm rooms preloaded.
func (m *DormModel) GetDormByID(id uint, p *Dorm) (err error) {
	err = m.DB.Preload("Neighborhood").Preload("DormRooms").First(p, id).Error
	return
}

func (m *DormModel) DoesDormExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Dorm{}).Where("dorms.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

// This updates the dorm statistics on room sizes and numbers. This function is expensive and it is currently called
// on every dorm update change. Luckily, this function is only called when the server is updating the dorm list,
// so it cannot be called by clients.
// TODO: Figure out a less-expensive way of calling this function: maybe a job queue that runs every minute?
func (m *DormModel) UpdateDormFacts(id uint) (err error) {
	var dorm Dorm
	err = m.DB.First(&dorm, id).Error
	if err != nil {
		return
	}

	type dormFacts struct {
		AvgSinglesArea  *float64
		AvgDoublesArea  *float64
		NumSingles      int
		NumDoubles      int
		NumFlexes       int
		ModeSinglesArea *int
		ModeDoublesArea *int
	}

	facts := dormFacts{}

	// Update average areas
	// Singles
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Select("avg(dorm_rooms.area) AS avg_singles_area").Scan(&facts).Error
	if err != nil {
		return
	}
	// Doubles
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Select("avg(dorm_rooms.area) AS avg_doubles_area").Scan(&facts).Error
	if err != nil {
		return
	}

	// Update room type counts
	// Single
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Count(&facts.NumSingles).Error
	if err != nil {
		return
	}
	// Double
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Count(&facts.NumDoubles).Error
	if err != nil {
		return
	}
	// Flex
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeFlex).
		Count(&facts.NumFlexes).Error
	if err != nil {
		return
	}

	// Update mode areas. We only error if it is not a record not found error, as if there are no rooms of that type,
	// we will get a record not found error, which is expected.
	// Singles
	err = m.DB.Model(&DormRoom{}).
		Select("area AS mode_singles_area").
		Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Group("dorm_rooms.area").Limit(1).
		Order("COUNT(*) DESC").
		Scan(&facts).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return
	}
	// Doubles
	err = m.DB.Model(&DormRoom{}).
		Select("area AS mode_doubles_area").
		Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Group("dorm_rooms.area").Limit(1).
		Order("COUNT(*) DESC").
		Scan(&facts).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return
	}

	// Have to cast float ptr to int ptr here. If nil, we ensure that it is also nil in the dorm
	if facts.AvgSinglesArea != nil {
		dorm.AverageSingleArea = lib.IntToPtr(int(math.Round(*facts.AvgSinglesArea)))
	} else {
		dorm.AverageSingleArea = nil
	}
	if facts.AvgDoublesArea != nil {
		dorm.AverageDoubleArea = lib.IntToPtr(int(math.Round(*facts.AvgDoublesArea)))
	} else {
		dorm.AverageDoubleArea = nil
	}
	dorm.NumberSingles = &facts.NumSingles
	dorm.NumberDoubles = &facts.NumDoubles
	dorm.NumberFlex = &facts.NumFlexes
	dorm.ModeSingleArea = facts.ModeSinglesArea
	dorm.ModeDoubleArea = facts.ModeDoublesArea

	err = m.DB.Save(&dorm).Error
	return
}
