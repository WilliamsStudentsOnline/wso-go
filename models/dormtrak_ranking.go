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
	queries := []*gorm.DB{
		m.DB.Model(&Dorm{}).Limit(max).Where("average_single_area IS NOT NULL").
			Order("average_single_area DESC").Find(&p.MaxMeanSingleSize),
		m.DB.Model(&Dorm{}).Limit(max).Where("average_single_area IS NOT NULL").
			Order("average_single_area ASC").Find(&p.MinMeanSingleSize),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes().Where("room_type = ?", DormRoomTypeSingle).
			Where("area IS NOT NULL").
			Order("area DESC").Find(&p.BiggestSingles),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes().Where("room_type = ?", DormRoomTypeSingle).
			Where("area IS NOT NULL").
			Order("area ASC").Find(&p.SmallestSingles),

		m.DB.Model(&Dorm{}).Limit(max).Where("average_double_area IS NOT NULL").
			Order("average_double_area DESC").Find(&p.MaxMeanDoubleSize),
		m.DB.Model(&Dorm{}).Limit(max).Where("average_double_area IS NOT NULL").
			Order("average_double_area ASC").Find(&p.MinMeanDoubleSize),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes().Where("room_type = ?", DormRoomTypeDouble).
			Where("area IS NOT NULL").
			Order("area DESC").Find(&p.BiggestDoubles),
		m.DB.Model(&DormRoom{}).Limit(max).Scopes().Where("room_type = ?", DormRoomTypeDouble).
			Where("area IS NOT NULL").
			Order("area ASC").Find(&p.SmallestDoubles),

		m.DB.Model(&Dorm{}).Limit(max).Where("number_singles IS NOT NULL").
			Order("number_singles DESC").Find(&p.MostSingles),
		m.DB.Model(&Dorm{}).Limit(max).Where("number_doubles IS NOT NULL").
			Order("number_doubles DESC").Find(&p.MostDoubles),

		m.DB.Model(&Dorm{}).Limit(max).Where("bathroom_ratio IS NOT NULL").
			Order("bathroom_ratio ASC").Find(&p.MostBathrooms),
		m.DB.Model(&Dorm{}).Limit(max).Where("bathroom_ratio IS NOT NULL").
			Order("bathroom_ratio DESC").Find(&p.FewestBathrooms),
	}

	for _, query := range queries {
		if query.Error != nil {
			return query.Error
		}
	}

	return
}
