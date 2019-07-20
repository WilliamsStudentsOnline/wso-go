package models

//UserModel ...
type UserModel struct {
	BaseModel
}

func (m *UserModel) GetAllUsers(u *[]User) (err error) {
	err = m.DB.Find(u).Error
	return
}

func (m *UserModel) GetUserByID(id uint, u *User) (err error) {
	//.Set("gorm:auto_preload", true)
	err = m.DB.Where(NewUserWithID(id)).First(u).Error
	return
}

// Update the user. Only allow specific keys to be passed
func (m *UserModel) UpdateUser(id uint, update map[string]interface{}) (err error) {
	MapPermit(update, "visible", "dorm_visible", "home_visible", "pronoun", "off_cycle")
	err = m.DB.Model(NewUserWithID(id)).Updates(update).Error
	return
}
