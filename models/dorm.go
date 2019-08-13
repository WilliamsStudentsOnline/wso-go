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
func (m *DormModel) GetAllDorms(p *[]*Dorm, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(p).Error
	return
}

type GetAllDormsOptions struct {
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: neighborhood, and dormRooms
	Preload *[]string `json:"preload" form:"preload"`
}

// Preload specifically allowed parts if requested
func (p *GetAllDormsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if p.Preload != nil {
		if lib.StringsContains(*p.Preload, "neighborhood") {
			db = db.Preload("Neighborhood")
		}
		if lib.StringsContains(*p.Preload, "dormRooms") {
			db = db.Preload("DormRooms")
		}
	}

	return db
}

func (p *GetAllDormsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("dorms.id ASC", true)
}

func (p *GetAllDormsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Offset != nil {
		db = db.Offset(*p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}

	return db
}

func (p *GetAllDormsOptions) Run(db *gorm.DB) *gorm.DB {
	return p.Paginate(p.Preloader(db))
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

// This updates the dorm review statistics.
func (m *DormModel) ReloadStatistics(id uint) (err error) {
	var dorm Dorm
	err = m.DB.First(&dorm, id).Error
	if err != nil {
		return
	}

	m.DB.Model(&DormtrakReview{}).
		Where("dormtrak_reviews.dorm_room_id in (?)",
			m.DB.Model(&DormRoom{}).
				Select("dorm_rooms.id").
				Where("dorm_rooms.dorm_id = ?", id).QueryExpr(),
		).
		Select("avg(dormtrak_reviews.wifi) AS wifi, " +
			"avg(dormtrak_reviews.comfort) AS comfort, " +
			"avg(dormtrak_reviews.convenience) AS convenience, " +
			"avg(dormtrak_reviews.location) AS location, " +
			"avg(dormtrak_reviews.loudness) AS loudness, " +
			"avg(dormtrak_reviews.satisfaction) AS satisfaction").
		Scan(&dorm)

	err = m.DB.Save(&dorm).Error
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

func (*DormModel) scopeTrakked(db *gorm.DB) *gorm.DB {
	return db.
		Joins("JOIN neighborhoods ON neighborhoods.id = dorms.neighborhood_id").
		Where("neighborhoods.trakked = ?", true)
}

// Preloads neighborhood
func (m *DormModel) preloadNeighborhood(db *gorm.DB) *gorm.DB {
	return db.Preload("Neighborhood")
}
