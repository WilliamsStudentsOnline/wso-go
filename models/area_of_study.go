package models

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
)

// AreaOfStudy Model
type AreaOfStudyModel struct {
	*BaseModel
}

func NewAreaOfStudyModel(db *gorm.DB) *AreaOfStudyModel {
	return &AreaOfStudyModel{
		BaseModel: NewBaseModel(db),
	}
}

type GetAllAreasOfStudyOptions struct {
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: department, courses
	Preload []string `json:"preload" form:"preload[]"`

	// Sorter: id, name (default to name)
	Sort string `json:"sort" form:"sort"`
}

// Preload specifically allowed parts if requested
func (o *GetAllAreasOfStudyOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "department") {
		db = db.Preload("Department")
	}
	if lib.StringsContains(o.Preload, "courses") {
		db = db.Preload("Courses")
	}

	return db
}

func (o *GetAllAreasOfStudyOptions) Order(db *gorm.DB) *gorm.DB {
	if o.Sort == "id" {
		return db.Order("areas_of_study.id ASC", true)
	}
	return db.Order("areas_of_study.name ASC", true)
}

func (o *GetAllAreasOfStudyOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Offset != nil {
		db = db.Offset(*o.Offset)
	}
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
	}

	return db
}

func (o *GetAllAreasOfStudyOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)
	db = o.Preloader(db)

	return db
}

// Gets all areas of study.
func (m *AreaOfStudyModel) GetAllAreasOfStudy(p *[]AreaOfStudy, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(p).Error
	return
}

// Gets area of study by its id with department preloaded.
func (m *AreaOfStudyModel) GetAreaOfStudyByID(id uint, p *AreaOfStudy) (err error) {
	err = m.DB.Preload("Department").First(p, id).Error
	return
}

// Gets area of study by its abbreviation. NOTE: all abbreviations are uppercase.
func (m *AreaOfStudyModel) GetAreaOfStudyByAbbreviation(abbreviation string, p *AreaOfStudy) (err error) {
	err = m.DB.Where(&AreaOfStudy{Abbreviation: strings.ToUpper(abbreviation)}).First(p).Error
	return
}

func (m *AreaOfStudyModel) DoesAreaOfStudyExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&AreaOfStudy{}).Where("areas_of_study.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}
