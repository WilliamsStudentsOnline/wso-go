package models

import "github.com/jinzhu/gorm"

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

func (m *DormModel) UpdateDormFacts(id uint) (err error) {
	var dorm Dorm
	err = m.DB.First(&dorm, id).Error
	if err != nil {
		return
	}

	// Update average areas
	// Singles
	var avgSingArea []interface{}
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Select("avg(area) AS avg_singles_area").Pluck("avg_singles_area", &avgSingArea).Error
	if err != nil {
		return
	}
	if len(avgSingArea) > 0 {
		if val, ok := avgSingArea[0].(int); ok {
			dorm.AverageSingleArea = &val
		}
	}
	// Doubles
	var avgDoubArea []interface{}
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Select("avg(area) AS avg_doubles_area").Pluck("avg_doubles_area", &avgDoubArea).Error
	if err != nil {
		return
	}
	if len(avgDoubArea) > 0 {
		if val, ok := avgDoubArea[0].(int); ok {
			dorm.AverageDoubleArea = &val
		}
	}

	// Update room type counts
	// Single
	var numSing int
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Count(&numSing).Error
	if err != nil {
		return
	}
	dorm.NumberSingles = &numSing
	// Double
	var numDoub int
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Count(&numDoub).Error
	if err != nil {
		return
	}
	dorm.NumberDoubles = &numDoub
	// Flex
	var numFlex int
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeFlex).
		Count(&numFlex).Error
	if err != nil {
		return
	}
	dorm.NumberFlex = &numFlex

	// update mode areas
	// Singles
	var modeSingArea []interface{}
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Group("dorm_rooms.area").Limit(1).Pluck("area", &modeSingArea).Error
	if err != nil {
		return
	}
	if len(modeSingArea) > 0 {
		if val, ok := modeSingArea[0].(int); ok {
			dorm.ModeSingleArea = &val
		}
	}

	// Doubles
	var modeDoubArea []interface{}
	err = m.DB.Model(&DormRoom{}).Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Group("dorm_rooms.area").Limit(1).Pluck("area", &modeDoubArea).Error
	if err != nil {
		return
	}
	if len(modeDoubArea) > 0 {
		if val, ok := modeDoubArea[0].(int); ok {
			dorm.ModeDoubleArea = &val
		}
	}

	err = m.DB.Save(dorm).Error
	return
}
