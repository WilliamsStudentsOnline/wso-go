package models

import (
	"github.com/jinzhu/gorm"
)

// Ephmatcher schema to cast into from a user
type Ephmatcher struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	UnixID string `json:"unixID"`
	Liked  bool   `json:"liked"` // If me (user) has an ephmatch entry where ephmatch.other_id=users.id and ephmatch.user_id=myID
}

// Ephmatcher Model
type EphmatcherModel struct {
	*UserModel
}

func NewEphmatcherModel(db *gorm.DB) *EphmatcherModel {
	return &EphmatcherModel{
		UserModel: NewUserModel(db),
	}
}

// Get all ephmatchers and populate the liked field. The userID should be self to see the liked field.
func (m *EphmatcherModel) GetAllEphmatchers(userID uint, p *[]*Ephmatcher, opts Options) (err error) {
	// Get ephmatchers
	db := m.DB.Model(&User{})
	db = m.scopeDefault(db)
	if opts != nil {
		db = opts.Run(db)
	}
	rows, err := db.Select("users.id, users.name, users.unix_id").Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	var ephmatchers []*Ephmatcher
	// Construct this for populating the liked field. It is a map of ephmatcher id to a ptr to ephmatcher struct
	ecIDToEC := make(map[uint]*Ephmatcher)

	// Directly scan it from the SQL to save time
	for rows.Next() {
		var ephmatcher Ephmatcher
		// ScanRows scan a row into ephmatcher
		err = m.DB.ScanRows(rows, &ephmatcher)
		if err != nil {
			return
		}

		// Add the new row/ephmatcher to the slice and map
		ephmatchers = append(ephmatchers, &ephmatcher)
		ecIDToEC[ephmatcher.ID] = &ephmatcher
	}

	// Get ephmatches made by user self (userID) to populate "liked" field. This is significantly faster
	// than a SQL query by several magnitudes.
	var ephmatches []*Ephmatch
	err = m.DB.Where("user_id = ?", userID).Find(&ephmatches).Error
	if err != nil {
		return
	}

	// If the ec is in the ephmatches list, we know the user liked them.
	for _, ephmatch := range ephmatches {
		if ec, ok := ecIDToEC[ephmatch.OtherID]; ok && ec != nil {
			ec.Liked = true
		}
	}

	// This copies our version of the ephmatchers list onto the passed list
	*p = ephmatchers

	return
}

type GetAllEphmatchersOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`
}

func (p *GetAllEphmatchersOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("users.name ASC", true)
}

func (p *GetAllEphmatchersOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
		if p.Offset != nil {
			db = db.Offset(*p.Offset)
		}
	}

	return db
}

func (p *GetAllEphmatchersOptions) Run(db *gorm.DB) *gorm.DB {
	return p.Paginate(db)
}

func (m *EphmatcherModel) CountEphmatchers() (count int, err error) {
	db := m.scopeDefault(m.DB.Model(&User{}))
	err = db.Count(&count).Error
	return
}

func (m *EphmatcherModel) GetEphmatcherByID(ephmatcherID uint, userID uint, p *Ephmatcher) (err error) {
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id = ?", ephmatcherID).Scan(p).Error
	if err != nil {
		return
	}

	var numEphmatches int
	err = m.DB.Model(&Ephmatch{}).
		Where("user_id = ? AND other_id = ?", userID, ephmatcherID).
		Count(&numEphmatches).Error
	if err != nil {
		return
	}

	if numEphmatches > 0 {
		p.Liked = true
	}
	return
}

// Get ephmatch matches of user.
func (m *EphmatcherModel) GetMatches(userID uint, p *[]*User) (err error) {
	// Get matches
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id IN (?)",
		m.DB.Model(&Ephmatch{}).Select("ephmatches.other_id").
			Joins("INNER JOIN ephmatches b ON b.other_id = ephmatches.user_id").
			Where("ephmatches.user_id = ? AND ephmatches.other_id = b.user_id", userID).QueryExpr(),
	).Find(p).Error
	return
}

func (m *EphmatcherModel) DoesEphmatcherExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

// Default scope: is student, is visible, and is ephmatch eligible
func (m *EphmatcherModel) scopeDefault(db *gorm.DB) *gorm.DB {
	db = m.scopeStudent(db)
	db = m.scopeVisible(db)
	return m.scopeEphmatchEligible(db)
}

func (m *EphmatcherModel) scopeStudent(db *gorm.DB) *gorm.DB {
	return db.Where("users.type = ?", UserTypeStudent)
}

// Ephmatch eligible means user has not opted out or user is specifically marked eligible
func (m *EphmatcherModel) scopeEphmatchEligible(db *gorm.DB) *gorm.DB {
	return db.Where("users.opt_out_ephmatch = ?", false)
}
