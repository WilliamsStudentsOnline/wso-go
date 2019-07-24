package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// FactrakSurvey Model
type FactrakSurveyModel struct {
	*BaseModel
}

func NewFactrakSurveyModel(db *gorm.DB) *FactrakSurveyModel {
	return &FactrakSurveyModel{
		BaseModel: NewBaseModel(db),
	}
}

func (m *FactrakSurveyModel) GetSurveysByProfessor(profID uint, u *[]*FactrakSurvey) (err error) {
	err = m.DB.Scopes(m.scopeDefault, m.scopeCurrent).Where(&FactrakSurvey{ProfessorID: profID}).Find(u).Error
	return
}


func (*FactrakSurveyModel) registrationStart() time.Time {
	// if changed, also change scheduled update user stuff
	springReg := time.October
	fallReg := time.March

	// Between October/X and February/X+1, want October/X
	now := time.Now()
	month := now.Month()

	// TODO: Ensure time.Local is EST/EDT on server
	if month <= time.February {
		return time.Date(now.Year()-1, springReg, 1, 1, 0, 0, 0, time.Local)
	} else if month >= time.October {
		return time.Date(now.Year(), springReg, 1, 1, 0, 0, 0, time.Local)
	} else {
		return time.Date(now.Year(), fallReg, 1, 1, 0, 0, 0, time.Local)
	}
}

func (*FactrakSurveyModel) scopeDefault(db *gorm.DB) *gorm.DB {
	return db.Order("factrak_surveys.created_at desc")
}

func (*FactrakSurveyModel) scopeFlagged(db *gorm.DB) *gorm.DB {
	return db.Where("factrak_surveys.flagged = ?", true)
}

func (*FactrakSurveyModel) scopeProfAtWilliams(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN users AS professors ON professors.id = factrak_surveys.professor_id").Where("professors.at_williams = ?", true)
}

func (m *FactrakSurveyModel) scopeThisSemester(db *gorm.DB) *gorm.DB {
	return db.Where("factrak_surveys.created_at = ?", m.registrationStart())
}

func (*FactrakSurveyModel) scopeCurrent(db *gorm.DB) *gorm.DB {
	return db.Where("factrak_surveys.created_at >= ?", time.Now().AddDate(-5, 0, 0))
}