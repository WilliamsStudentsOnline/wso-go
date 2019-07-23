package models

import (
	"github.com/jinzhu/gorm"
)

// Professor Model
type ProfessorModel struct {
	*UserModel
}
func (m *ProfessorModel) GetAllProfessors(u *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault, m.scopeAtWilliams).Find(u).Error
	return
}

func (m *ProfessorModel) GetProfessorByID(id uint, u *User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Preload("FactrakSurveys").Preload("Courses").Where(NewUserWithID(id)).First(u).Error
	return
}

func (m *ProfessorModel) scopeDefault(db *gorm.DB) *gorm.DB {
	return db.Where("users.type = ?", UserTypeProfessor)
}
