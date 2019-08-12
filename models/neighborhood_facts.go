package models

import "github.com/jinzhu/gorm"

type NeighborhoodFacts struct {
	SophomoreCount int `json:"sophomoreCount"`
	JuniorCount    int `json:"juniorCount"`
	SeniorCount    int `json:"seniorCount"`
}

// Get neighborhood fast-facts. This is not a lightweight operation, but this
// gives us more in-depth info on the neighborhoods. But, this does do a number of SQL queries, so be wary.
func (m *NeighborhoodModel) GetNeighborhoodFacts(id uint, p *NeighborhoodFacts) (err error) {
	// Get the dorm
	var neighborhood Neighborhood
	err = m.GetNeighborhoodByID(id, &neighborhood)
	if err != nil {
		return
	}

	studModel := NewStudentModel(m.DB)

	queries := []*gorm.DB{
		m.DB.Model(&User{}).Where(
			"dorm_room_id in (?)",
			m.DB.Model(&DormRoom{}).Select("dorm_rooms.id").Where(
				"dorm_rooms.dorm_id IN (?)",
				m.DB.Model(&Dorm{}).Select("dorms.id").Where(
					"dorms.neighborhood_id = ?", id,
				).QueryExpr(),
			).QueryExpr(),
		).
			Where("users.type = ?", UserTypeStudent).
			Where("users.class_year = ?", studModel.SeniorYear()).Count(&p.SeniorCount),

		m.DB.Model(&User{}).Where(
			"dorm_room_id in (?)",
			m.DB.Model(&DormRoom{}).Select("dorm_rooms.id").Where(
				"dorm_rooms.dorm_id IN (?)",
				m.DB.Model(&Dorm{}).Select("dorms.id").Where(
					"dorms.neighborhood_id = ?", id,
				).QueryExpr(),
			).QueryExpr(),
		).
			Where("users.type = ?", UserTypeStudent).
			Where("users.class_year = ?", studModel.SeniorYear()+1).Count(&p.JuniorCount),

		m.DB.Model(&User{}).Where(
			"dorm_room_id in (?)",
			m.DB.Model(&DormRoom{}).Select("dorm_rooms.id").Where(
				"dorm_rooms.dorm_id IN (?)",
				m.DB.Model(&Dorm{}).Select("dorms.id").Where(
					"dorms.neighborhood_id = ?", id,
				).QueryExpr(),
			).QueryExpr(),
		).
			Where("users.type = ?", UserTypeStudent).
			Where("users.class_year = ?", studModel.SeniorYear()+2).Count(&p.SophomoreCount),
	}

	for _, query := range queries {
		if query.Error != nil {
			return query.Error
		}
	}

	return
}
