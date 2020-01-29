package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// EphmatchProfile Model
type EphmatchProfileModel struct {
	*BaseModel
}

func NewEphmatchProfileModel(db *gorm.DB, log *zap.SugaredLogger) *EphmatchProfileModel {
	return &EphmatchProfileModel{
		BaseModel: NewBaseModel(db, log),
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

// Get all ephmatch profiles.
func (m *EphmatchProfileModel) GetAllProfilesNoSelf(p *[]*EphmatchProfile, selfID uint, opts Options) (err error) {
	// Get profiles
	db := m.DB.Model(&EphmatchProfile{}).Preload("User")

	db = m.scopeDefault(db)
	if opts != nil {
		db = opts.Run(db)
	}

	db = db.Not(EphmatchProfile{UserID: selfID})

	return db.Find(p).Error
}

type GetAllProfilesOptions struct {
	// Can be new, updated, or alphabetical
	Sort *string `json:"sort" form:"sort"`

	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: tags
	Preload []string `json:"preload" form:"preload[]"`
}

func (p *GetAllProfilesOptions) Order(db *gorm.DB) *gorm.DB {
	if p.Sort != nil {
		if *p.Sort == "new" {
			return db.Order("ephmatch_profiles.created_at DESC", true)
		} else if *p.Sort == "updated" {
			return db.Order("ephmatch_profiles.updated_at DESC", true)
		}
	}

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

// Preload specifically allowed parts if requested
func (o *GetAllProfilesOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "tags") {
		db = db.Preload("User.Tags")
	}

	return db
}

func (p *GetAllProfilesOptions) Run(db *gorm.DB) *gorm.DB {
	return p.Paginate(p.Preloader(db))
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

// Create or update a new profile unscoped about deleted. Updating a deleted profile will make it be undeleted
func (m *EphmatchProfileModel) CreateOrUpdateProfileUnscoped(userID uint, newProfile EphmatchProfile, p *EphmatchProfile) (err error) {

	var tempProf EphmatchProfile
	err = m.DB.Unscoped().Where(EphmatchProfile{UserID: userID}).First(&tempProf).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	// If our profile was deleted, undelete it
	if tempProf.DeletedAt != nil {
		query := m.DB.
			Unscoped().
			Model(&EphmatchProfile{}).
			Where(EphmatchProfile{UserID: userID})
		if newProfile.Gender != nil {
			query = query.Update("gender", newProfile.Gender)
		}
		if newProfile.Description != nil {
			query = query.Update("description", newProfile.Description)
		}
		if newProfile.MatchMessage != nil {
			query = query.Update("match_message", newProfile.MatchMessage)
		}
		err = query.UpdateColumn("deleted_at", nil).
			Preload("User").
			First(p).
			Error

		return err
	}

	err = m.DB.
		Model(&EphmatchProfile{}).
		Where(EphmatchProfile{UserID: userID}).
		Assign(EphmatchProfile{
			Gender:       newProfile.Gender,
			Description:  newProfile.Description,
			MatchMessage: newProfile.MatchMessage,
		}).
		Preload("User").
		FirstOrCreate(p).Error
	return err
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
		Preload("User.Tags").
		Where("ephmatch_profiles.user_id = ?", userID).
		First(p).Error
	return
}

func (m *EphmatchProfileModel) GetSelfProfileByIDScopedNoDefault(userID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Preload("User").
		Preload("User.Tags").
		Where("ephmatch_profiles.user_id = ?", userID).
		First(p).Error
	return
}

func (m *EphmatchProfileModel) GetProfileByID(profileUserID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Scopes(m.scopeDefault).
		Where("ephmatch_profiles.user_id = ?", profileUserID).
		Preload("User").
		Preload("User.Tags").
		First(p).Error
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
		Where("users.type = ?", UserTypeStudent)
	return db
}
