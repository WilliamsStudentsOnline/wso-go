package models

import (
	"github.com/jinzhu/gorm"
)

// Ephcatcher schema to cast into from a user
type Ephcatcher struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	UnixID string `json:"unixID"`
	Liked  bool   `json:"liked"` // If me (user) has an ephcatch entry where ephcatch.other_id=users.id and ephcatch.user_id=myID
}

// Ephcatcher Model
type EphcatcherModel struct {
	*UserModel
}

func NewEphcatcherModel(db *gorm.DB) *EphcatcherModel {
	return &EphcatcherModel{
		UserModel: NewUserModel(db),
	}
}

// Get all ephcatchers and populate the liked field. The userID should be self to see the liked field.
func (m *EphcatcherModel) GetAllEphcatchers(userID uint, p *[]*Ephcatcher, opts Options) (err error) {
	// Get ephcatchers
	db := m.DB.Model(&User{}).Scopes(m.scopeDefault)
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Preloader(db)
	}
	rows, err := db.Select("users.id, users.name, users.unix_id").Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	var ephcatchers []*Ephcatcher
	// Construct this for populating the liked field. It is a map of ephcatcher id to a ptr to ephcatcher struct
	ecIDToEC := make(map[uint]*Ephcatcher)

	// Directly scan it from the SQL to save time
	for rows.Next() {
		var ephcatcher Ephcatcher
		// ScanRows scan a row into ephcatcher
		err = m.DB.ScanRows(rows, &ephcatcher)
		if err != nil {
			return
		}

		// Add the new row/ephcatcher to the slice and map
		ephcatchers = append(ephcatchers, &ephcatcher)
		ecIDToEC[ephcatcher.ID] = &ephcatcher
	}

	// Get ephcatches made by user self (userID) to populate "liked" field. This is significantly faster
	// than a SQL query by several magnitudes.
	var ephcatches []*Ephcatch
	err = m.DB.Where("user_id = ?", userID).Find(&ephcatches).Error
	if err != nil {
		return
	}

	// If the ec is in the ephcatches list, we know the user liked them.
	for _, ephcatch := range ephcatches {
		if ec, ok := ecIDToEC[ephcatch.OtherID]; ok && ec != nil {
			ec.Liked = true
		}
	}

	// This copies our version of the ephcatchers list onto the passed list
	*p = ephcatchers

	return
}

type GetAllEphcatchersOptions struct {
	Offset *string `json:"offset" form:"offset"`
	Limit  *uint   `json:"limit" form:"limit"`
}

// Preload specifically allowed parts if requested
func (p *GetAllEphcatchersOptions) Preloader(db *gorm.DB) *gorm.DB {
	return db
}

func (p *GetAllEphcatchersOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("users.name ASC")
}

func (p *GetAllEphcatchersOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Offset != nil {
		db = db.Where("users.name > ?", *p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}

	return db
}

func (m *EphcatcherModel) GetEphcatcherByID(ephcatcherID uint, userID uint, p *Ephcatcher) (err error) {
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id = ?", ephcatcherID).Scan(p).Error
	if err != nil {
		return
	}

	var numEphcatches int
	err = m.DB.Model(&Ephcatch{}).
		Where("user_id = ? AND other_id = ?", userID, ephcatcherID).
		Count(&numEphcatches).Error
	if err != nil {
		return
	}

	if numEphcatches > 0 {
		p.Liked = true
	}
	return
}

// Get ephcatch matches of user.
func (m *EphcatcherModel) GetMatches(userID uint, p *[]*User) (err error) {
	// Get matches
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id IN (?)",
		m.DB.Model(&Ephcatch{}).Select("ephcatches.other_id").
			Joins("INNER JOIN ephcatches b ON b.other_id = ephcatches.user_id").
			Where("ephcatches.user_id = ? AND ephcatches.other_id = b.user_id", userID).QueryExpr(),
	).Find(p).Error
	return
}

func (m *EphcatcherModel) DoesEphcatcherExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

// Default scope: is student, is visible, and is ephcatch eligible
func (m *EphcatcherModel) scopeDefault(db *gorm.DB) *gorm.DB {
	db = m.scopeStudent(db)
	db = m.scopeVisible(db)
	return m.scopeEphcatchEligible(db)
}

func (m *EphcatcherModel) scopeStudent(db *gorm.DB) *gorm.DB {
	return db.Where("users.type = ?", UserTypeStudent)
}

// Ephcatch eligible means user is a senior and has not opted out or user is specifically marked eligible
func (m *EphcatcherModel) scopeEphcatchEligible(db *gorm.DB) *gorm.DB {
	return db.Where("(users.class_year = ? AND users.opt_out_ephcatch = ?) OR users.ephcatch_eligibility = ?",
		(&StudentModel{}).SeniorYear(), false, true)
}
