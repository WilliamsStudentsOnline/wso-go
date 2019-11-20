package models

import (
	"github.com/jinzhu/gorm"
)

// EphmatchProfile Model
type EphmatchProfileModel struct {
	*BaseModel
}

func NewEphmatchProfileModel(db *gorm.DB) *EphmatchProfileModel {
	return &EphmatchProfileModel{
		BaseModel: NewBaseModel(db),
	}
}

// Get all ephmatch profiles.
func (m *EphmatchProfileModel) GetAllProfiles(p *[]*EphmatchProfile, opts Options) (err error) {
	// Get profiles
	db := m.DB.Model(&EphmatchProfile{}).Preload("User")

	db = m.scopeDefault(db)
	if opts != nil {
		db = opts.Run(db)
	}

	return db.Find(p).Error
}

type GetAllProfilesOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`
}

func (p *GetAllProfilesOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("users.name ASC", true)
}

func (p *GetAllProfilesOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
		if p.Offset != nil {
			db = db.Offset(*p.Offset)
		}
	}

	return db
}

func (p *GetAllProfilesOptions) Run(db *gorm.DB) *gorm.DB {
	return p.Paginate(db)
}

func (m *EphmatchProfileModel) CountProfiles() (count int, err error) {
	db := m.scopeDefault(m.DB.Model(&EphmatchProfile{}))
	err = db.Count(&count).Error
	return
}

// Create a new survey.
func (m *EphmatchProfileModel) CreateProfile(p *EphmatchProfile) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}

	err = m.DB.
		Preload("User").
		First(p).Error
	return
}

// Create or update a new profile unscoped about deleted.
func (m *EphmatchProfileModel) CreateOrUpdateProfileUnscoped(userID uint, newProfile EphmatchProfile, p *EphmatchProfile) (err error) {
	tx := m.DB.Begin()

	defer tx.RollbackUnlessCommitted()

	var tempProf EphmatchProfile
	err = tx.Unscoped().Where(EphmatchProfile{UserID: userID}).First(&tempProf).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	// If our profile was deleted, undelete it
	if tempProf.DeletedAt != nil {
		err = tx.
			Unscoped().
			Model(&tempProf).
			Where(EphmatchProfile{UserID: userID}).
			Update("deleted_at = ?", nil).
			Error
		if err != nil {
			return err
		}
	}

	err = tx.
		Where(EphmatchProfile{UserID: userID}).
		Assign(EphmatchProfile{
			Gender:      newProfile.Gender,
			Description: newProfile.Description,
		}).
		Preload("User").
		FirstOrCreate(p).Error
	if err != nil {
		return err
	}

	return tx.Commit().Error
}

func (m *EphmatchProfileModel) UpdateProfile(p *EphmatchProfile) (err error) {
	err = m.DB.Save(&p).Error
	if err != nil {
		return
	}

	err = m.DB.Preload("User").First(p, p.ID).Error
	return
}

func (m *EphmatchProfileModel) DeleteProfile(p *EphmatchProfile) (err error) {
	err = m.DB.Delete(p).Error
	return
}

func (m *EphmatchProfileModel) GetSelfProfileByID(userID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Unscoped().
		Preload("User").
		Where("ephmatch_profiles.user_id = ?", userID).
		Scan(p).Error
	return
}

func (m *EphmatchProfileModel) GetSelfProfileByIDScopedNoDefault(userID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Preload("User").
		Where("ephmatch_profiles.user_id = ?", userID).
		Scan(p).Error
	return
}

func (m *EphmatchProfileModel) GetProfileByID(profileUserID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Scopes(m.scopeDefault).
		Where("ephmatch_profiles.user_id = ?", profileUserID).
		Preload("User").
		Scan(p).Error
	if err != nil {
		return
	}

	return
}

// Get ephmatch matches of user.
// TODO: Deprecate or remove this?
func (m *EphmatchProfileModel) GetMatches(userID uint, p *[]*User) (err error) {
	// Get matches
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id IN (?)",
		m.DB.Model(&Ephmatch{}).Select("ephmatches.other_id").
			Joins("INNER JOIN ephmatches b ON b.other_id = ephmatches.user_id").
			Where("ephmatches.user_id = ? AND ephmatches.other_id = b.user_id", userID).QueryExpr(),
	).Find(p).Error
	return
}

func (m *EphmatchProfileModel) DoesProfileExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&EphmatchProfile{}).Scopes(m.scopeDefault).Where("ephmatch_profiles.user_id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

// Default scope: is student and is visible. Make sure to enumerate users on join
func (m *EphmatchProfileModel) scopeDefault(db *gorm.DB) *gorm.DB {
	db = db.Joins("INNER JOIN users ON users.id = ephmatch_profiles.user_id").
		Where("users.type = ?", UserTypeStudent).
		Where("users.visible = ?", true)
	return db
}
