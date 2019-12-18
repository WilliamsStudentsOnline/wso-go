package models

import "github.com/jinzhu/gorm"

type DormtrakRanking struct {
	MaxMeanSingleSize []*Dorm     `json:"maxMeanSingleSize"`
	MinMeanSingleSize []*Dorm     `json:"minMeanSingleSize"`
	BiggestSingles    []*DormRoom `json:"biggestSingles"`
	SmallestSingles   []*DormRoom `json:"smallestSingles"`

	MaxMeanDoubleSize []*Dorm     `json:"maxMeanDoubleSize"`
	MinMeanDoubleSize []*Dorm     `json:"minMeanDoubleSize"`
	BiggestDoubles    []*DormRoom `json:"biggestDoubles"`
	SmallestDoubles   []*DormRoom `json:"smallestDoubles"`

	MostSingles []*Dorm `json:"mostSingles"`
	MostDoubles []*Dorm `json:"mostDoubles"`

	MostBathrooms   []*Dorm `json:"mostBathrooms"`
	FewestBathrooms []*Dorm `json:"fewestBathrooms"`
}

func NewDormtrakRanking() *DormtrakRanking {
	return &DormtrakRanking{
		MaxMeanSingleSize: []*Dorm{},
		MinMeanSingleSize: []*Dorm{},
		BiggestSingles:    []*DormRoom{},
		SmallestSingles:   []*DormRoom{},
		MaxMeanDoubleSize: []*Dorm{},
		MinMeanDoubleSize: []*Dorm{},
		BiggestDoubles:    []*DormRoom{},
		SmallestDoubles:   []*DormRoom{},
		MostSingles:       []*Dorm{},
		MostDoubles:       []*Dorm{},
		MostBathrooms:     []*Dorm{},
		FewestBathrooms:   []*Dorm{},
	}
}

func (m *DormModel) GetDormtrakRankings(max int, p *DormtrakRanking) (err error) {
	drM := NewDormRoomModel(m.DB, m.log)

	queries := []*gorm.DB{
		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.average_single_area IS NOT NULL").
			Order("dorms.average_single_area DESC").Find(&p.MaxMeanSingleSize),
		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.average_single_area IS NOT NULL").
			Order("dorms.average_single_area ASC").Find(&p.MinMeanSingleSize),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes(drM.scopeTrakked, drM.preloadDorm).
			Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
			Where("dorm_rooms.area IS NOT NULL").
			Order("dorm_rooms.area DESC").Find(&p.BiggestSingles),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes(drM.scopeTrakked, drM.preloadDorm).
			Where("dorm_rooms.room_type = ?", DormRoomTypeSingle).
			Where("dorm_rooms.area IS NOT NULL").
			Order("dorm_rooms.area ASC").Find(&p.SmallestSingles),

		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.average_double_area IS NOT NULL").
			Order("dorms.average_double_area DESC").Find(&p.MaxMeanDoubleSize),
		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.average_double_area IS NOT NULL").
			Order("dorms.average_double_area ASC").Find(&p.MinMeanDoubleSize),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes(drM.scopeTrakked, drM.preloadDorm).
			Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
			Where("dorm_rooms.area IS NOT NULL").
			Order("dorm_rooms.area DESC").Find(&p.BiggestDoubles),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes(drM.scopeTrakked, drM.preloadDorm).
			Where("dorm_rooms.room_type = ?", DormRoomTypeDouble).
			Where("dorm_rooms.area IS NOT NULL").
			Order("dorm_rooms.area ASC").Find(&p.SmallestDoubles),

		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.number_singles IS NOT NULL").
			Order("dorms.number_singles DESC").Find(&p.MostSingles),
		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.number_doubles IS NOT NULL").
			Order("dorms.number_doubles DESC").Find(&p.MostDoubles),

		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.bathroom_ratio IS NOT NULL").
			Order("dorms.bathroom_ratio ASC").Find(&p.MostBathrooms),
		m.DB.Model(&Dorm{}).Limit(max).Scopes(m.ScopeTrakked).Where("dorms.bathroom_ratio IS NOT NULL").
			Order("dorms.bathroom_ratio DESC").Find(&p.FewestBathrooms),
	}

	for _, query := range queries {
		if query.Error != nil {
			return query.Error
		}
	}

	return
}
