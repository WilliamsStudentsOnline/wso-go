package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
)

// Dorm Model
type DormRoomModel struct {
	*BaseModel
}

func NewDormRoomModel(db *gorm.DB) *DormRoomModel {
	return &DormRoomModel{
		BaseModel: NewBaseModel(db),
	}
}

// Gets all dorm rooms.
func (m *DormRoomModel) GetAllDormRooms(p *[]*DormRoom) (err error) {
	err = m.DB.Find(p).Error
	return
}

// Gets dorm rooms by its id with neighborhood and dorm preloaded.
func (m *DormRoomModel) GetDormRoomByID(id uint, p *DormRoom) (err error) {
	err = m.DB.Preload("Dorm").Preload("Dorm.Neighborhood").First(p, id).Error
	return
}

// Gets all dorm rooms by dorm building.
func (m *DormRoomModel) GetDormRoomsByDorm(dormID uint, p *[]*DormRoom, paginator Paginator) (err error) {
	db := m.DB.Where("dorm_rooms.dorm_id = ?", dormID)
	if paginator != nil {
		db = db.Scopes(paginator.Paginate)
	}
	err = db.Find(p).Error
	return
}

// This updates the latest room info based on the latest review. It also updates the dorm's rating statistics.
// In addition it has the consequence of running DormModel.UpdateDormFacts().
func (m *DormRoomModel) ReloadStatistics(id uint) (err error) {
	var room DormRoom
	err = m.DB.First(&room, id).Error
	if err != nil {
		return
	}

	var latestReview DormtrakReview
	err = m.DB.Where("dormtrak_reviews.dorm_room_id = ?", id).
		Order("dormtrak_reviews.created_at desc").
		First(&latestReview).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return
	}

	// Latest review either supersedes the previous value if it exists;
	// if there is no value in the latest review, default to previous room value.
	room.Closet = lib.StrPtrDefaults(latestReview.Closet, room.Closet)
	room.Flooring = lib.StrPtrDefaults(latestReview.Flooring, room.Flooring)
	room.CommonRoomAccess = lib.BoolPtrDefaults(latestReview.CommonRoomAccess, room.CommonRoomAccess)
	room.CommonRoomDesc = lib.StrPtrDefaults(latestReview.CommonRoomDesc, room.CommonRoomDesc)
	room.ThermostatAccess = lib.BoolPtrDefaults(latestReview.ThermostatAccess, room.ThermostatAccess)
	room.ThermostatDesc = lib.StrPtrDefaults(latestReview.ThermostatDesc, room.ThermostatDesc)
	room.OutletsDesc = lib.StrPtrDefaults(latestReview.OutletsDesc, room.OutletsDesc)
	room.KeyOrCard = lib.StrPtrDefaults(latestReview.KeyOrCard, room.KeyOrCard)
	room.Noise = lib.StrPtrDefaults(latestReview.Noise, room.Noise)
	room.BedAdjustable = lib.BoolPtrDefaults(latestReview.BedAdjustable, room.BedAdjustable)
	room.PrivateBathroom = lib.BoolPtrDefaults(latestReview.PrivateBathroom, room.PrivateBathroom)
	room.BathroomDesc = lib.StrPtrDefaults(latestReview.BathroomDesc, room.BathroomDesc)

	err = m.DB.Save(&room).Error
	if err != nil {
		return
	}

	// Update the dorm rating statistics now
	err = NewDormModel(m.DB).ReloadStatistics(room.DormID)
	return
}

type DormRoomPaginator struct {
	Offset uint
	Limit  uint
}

func (p *DormRoomPaginator) Order(db *gorm.DB) *gorm.DB {
	return db.Order("id ASC")
}

func (p *DormRoomPaginator) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db).Offset(p.Offset).Limit(p.Limit)
	return db
}

func (m *DormRoomModel) NewDormRoomPaginate(offset int, limit int) Paginator {
	o := uint(offset)
	l := uint(limit)

	if l == 0 {
		return &NoPaginator{}
	}

	return &DormRoomPaginator{
		Offset: o,
		Limit:  l,
	}
}
