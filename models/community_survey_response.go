package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type CommunitySurveyResponseModel struct {
	*BaseModel
}

func NewCommunitySurveyResponseModel(db *gorm.DB, log *zap.SugaredLogger) *CommunitySurveyResponseModel {
	return &CommunitySurveyResponseModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllCommunitySurveyResponseOptions struct {
	//offset is ignored unless limit is applied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	//filters
	SurveyID *uint `json:"surveyID" form:"surveyID"`
	UserID   *uint `json:"userID" form:"userID"`

	//Preloading: user,survey
	Preload []string `json:"preload" form:"preload[]"`
}

func (o *GetAllCommunitySurveyResponseOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("community_survey_response.created_at DESC")
}

func (o *GetAllCommunitySurveyResponseOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

func (o *GetAllCommunitySurveyResponseOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}
	if stringsContains(o.Preload, "user") {
		db = db.Preload("User")
	}
	if stringsContains(o.Preload, "survey") {
		db = db.Preload("Survey")
	}
	return db
}

func (m *CommunitySurveyResponseModel) GetAllCommunitySurveyResponse(c *[]*CommunitySurveyResponse, opts *GetAllCommunitySurveyResponseOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Run(db)
	}

	err = db.Find(c).Error
	return
}

func (o *GetAllCommunitySurveyResponseOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)
	m := NewCommunitySurveyResponseModel(db.New(), nil)
	if o.UserID != nil {
		db = m.withUser(*o.UserID)(db)
	}
	if o.SurveyID != nil {
		db = m.withSurvey(*o.SurveyID)(db)
	}
	return db
}

func (m *CommunitySurveyResponseModel) withUser(userID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

func (m *CommunitySurveyResponseModel) withSurvey(surveyID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("survey_id = ?", surveyID)
	}
}

func (m *CommunitySurveyResponseModel) CreateCommunitySurveyResponse(c *CommunitySurveyResponse) (err error) {
	err = m.DB.Create(c).Error
	if err != nil {
		return err
	}
	err = m.DB.Find(c).Error
	return
}

func (m *CommunitySurveyResponseModel) GetCommunitySurveyResponse(id uint, c *CommunitySurveyResponse) (err error) {
	err = m.DB.First(c, id).Error
	return
}

func (m *CommunitySurveyResponseModel) DeleteCommunitySurveyResponse(c *CommunitySurveyResponse) (err error) {
	err = m.DB.Delete(c).Error
	return
}
