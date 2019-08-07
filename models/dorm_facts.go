package models

import (
	"github.com/jinzhu/gorm"
)

type DormFacts struct {
	SophomoreCount int `json:"sophomoreCount"`
	JuniorCount    int `json:"juniorCount"`
	SeniorCount    int `json:"seniorCount"`

	Capacity               int      `json:"capacity"`
	SinglesCount           int      `json:"singlesCount"`
	DoublesCount           int      `json:"doublesCount"`
	FlexCount              int      `json:"flexCount"`
	AverageSinglesArea     *int     `json:"averageSinglesArea"`
	AverageDoublesArea     *int     `json:"averageDoublesArea"`
	ModeSinglesArea        *int     `json:"modeSinglesArea"`
	ModeDoublesArea        *int     `json:"modeDoublesArea"`
	StudentToBathroomRatio *float64 `json:"studentToBathroomRatio"`
	KeyOrCard              *string  `json:"keyOrCard"`
	WashersCount           *int     `json:"washersCount"`

	BiggestSingle  *DormRoom `json:"biggestSingle"`
	SmallestSingle *DormRoom `json:"smallestSingle"`
	BiggestDouble  *DormRoom `json:"biggestDouble"`
	SmallestDouble *DormRoom `json:"smallestDouble"`

	CommonRoomAccessRatio *float64 `json:"commonRoomAccessRatio"`

	AverageWifi         *float64 `json:"averageWifi"`
	AverageLocation     *float64 `json:"averageLocation"`
	AverageLoudness     *float64 `json:"averageLoudness"`
	AverageSatisfaction *float64 `json:"averageSatisfaction"`
}

func NewDormFacts() *DormFacts {
	return &DormFacts{
		BiggestSingle:  &DormRoom{},
		SmallestSingle: &DormRoom{},
		BiggestDouble:  &DormRoom{},
		SmallestDouble: &DormRoom{},
	}
}

// Get dorm fast-facts. Some of these are found also in GetDorm, which is a lightweight operation, but this
// gives us more in-depth info on the dorm. But, this does do a number of SQL queries, so be wary.
func (m *DormModel) GetDormFacts(id uint, p *DormFacts) (err error) {
	// Get the dorm
	var dorm Dorm
	err = m.GetDormByID(id, &dorm)
	if err != nil {
		return
	}

	// Initialize fields
	p.BiggestSingle = &DormRoom{}
	p.SmallestSingle = &DormRoom{}
	p.BiggestDouble = &DormRoom{}
	p.SmallestDouble = &DormRoom{}

	// Load in the inherited facts (for the dorm)
	p.Capacity = *dorm.Capacity
	p.SinglesCount = *dorm.NumberSingles
	p.DoublesCount = *dorm.NumberDoubles
	p.FlexCount = *dorm.NumberFlex
	p.AverageSinglesArea = dorm.AverageSingleArea
	p.AverageDoublesArea = dorm.AverageDoubleArea
	p.ModeSinglesArea = dorm.ModeSingleArea
	p.ModeDoublesArea = dorm.ModeDoubleArea
	p.StudentToBathroomRatio = dorm.BathroomRatio
	p.KeyOrCard = dorm.KeyOrCard
	p.WashersCount = dorm.NumberWashers

	studModel := NewStudentModel(m.DB)

	var commonRoomAccessCount int

	queries := []*gorm.DB{
		m.DB.Model(&User{}).Where(
			"dorm_room_id in (?)",
			m.DB.Model(&DormRoom{}).Select("dorm_rooms.id").Where(
				"dorm_rooms.dorm_id = ?", id,
			).QueryExpr(),
		).
			Where("users.type = ?", UserTypeStudent).
			Where("users.class_year = ?", studModel.SeniorYear()).Count(&p.SeniorCount),

		m.DB.Model(&User{}).Where(
			"dorm_room_id in (?)",
			m.DB.Model(&DormRoom{}).Select("dorm_rooms.id").Where(
				"dorm_rooms.dorm_id = ?", id,
			).QueryExpr(),
		).
			Where("users.type = ?", UserTypeStudent).
			Where("users.class_year = ?", studModel.SeniorYear()+1).Count(&p.JuniorCount),

		m.DB.Model(&User{}).Where(
			"dorm_room_id in (?)",
			m.DB.Model(&DormRoom{}).Select("dorm_rooms.id").Where(
				"dorm_rooms.dorm_id = ?", id,
			).QueryExpr(),
		).
			Where("users.type = ?", UserTypeStudent).
			Where("users.class_year = ?", studModel.SeniorYear()+2).Count(&p.SophomoreCount),

		m.DB.Model(&DormRoom{}).
			Where("dorm_rooms.dorm_id = ?", id).
			Where("dorm_rooms.common_room_access = ?", true).
			Count(&commonRoomAccessCount),

		m.DB.Model(&DormRoom{}).
			Where("dorm_rooms.dorm_id = ?", id).
			Select("avg(dorm_rooms.wifi) AS average_wifi, " +
				"avg(dorm_rooms.location) AS average_location, " +
				"avg(dorm_rooms.loudness) AS average_loudness, " +
				"avg(dorm_rooms.satisfaction) AS average_satisfaction").
			Scan(&p),
	}

	for _, query := range queries {
		if query.Error != nil {
			return query.Error
		}
	}

	// Do top dorm room queries here, as we need to nil out stuff
	err = m.DB.Model(&DormRoom{}).
		Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Order("dorm_rooms.area DESC").
		First(p.BiggestSingle).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			p.BiggestSingle = nil
			err = nil
		} else {
			return
		}
	}
	err = m.DB.Model(&DormRoom{}).
		Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
		Order("dorm_rooms.area ASC").
		First(p.SmallestSingle).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			p.SmallestSingle = nil
			err = nil
		} else {
			return
		}
	}
	err = m.DB.Model(&DormRoom{}).
		Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Order("dorm_rooms.area DESC").
		First(p.BiggestDouble).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			p.BiggestDouble = nil
			err = nil
		} else {
			return
		}
	}
	err = m.DB.Model(&DormRoom{}).
		Where("dorm_rooms.dorm_id = ?", id).
		Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
		Order("dorm_rooms.area ASC").
		First(p.SmallestDouble).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			p.SmallestDouble = nil
			err = nil
		} else {
			return
		}
	}

	// Calculate common room access
	totalDorms := *dorm.NumberSingles + *dorm.NumberDoubles + *dorm.NumberFlex
	if totalDorms != 0 {
		croomAccessRatio := float64(commonRoomAccessCount) / float64(totalDorms)
		p.CommonRoomAccessRatio = &croomAccessRatio
	}

	return
}
