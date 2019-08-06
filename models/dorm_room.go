package models

import "github.com/jinzhu/gorm"

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
