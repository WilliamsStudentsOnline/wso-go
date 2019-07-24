package models

import (
	"github.com/jinzhu/gorm"
)

// Professor Model
type ProfessorModel struct {
	*UserModel
}

func NewProfessorModel(db *gorm.DB) *ProfessorModel {
	return &ProfessorModel{
		UserModel: NewUserModel(db),
	}
}

func (m *ProfessorModel) GetAllProfessors(u *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault, m.scopeAtWilliams).Find(u).Error
	return
}

func (m *ProfessorModel) DoesProfessorExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault, m.scopeAtWilliams).Where("users.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

func (m *ProfessorModel) GetProfessorByID(id uint, u *User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Preload("ProfessorFactrakSurveys").Where(NewUserWithID(id)).First(u).Error
	return
}

func (m *ProfessorModel) scopeDefault(db *gorm.DB) *gorm.DB {
	return db.Where("users.type = ?", UserTypeProfessor)
}
