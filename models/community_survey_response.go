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

type getAllCommunitySurveyResponseOptions struct {
	//offset is ignored unless limit is applied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	//filters
	SurveyID *uint `json:"surveyID" form:"surveyID"`
	UserID   *uint `json:"userID" form:"userID"`

	//Preloading: user,survey
	Preload []string `json:"preload" form:"preload[]"`
}

func (o *getAllCommunitySurveyResponseOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("community_survey_response.created_at DESC")
}

func (o *getAllCommunitySurveyResponseOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

func (o *getAllCommunitySurveyResponseOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	communitySurveyResponseModel := NewCommunitySurveyResponseModel(nil, nil)
	if stringsContains(o.Preload, "user") {
		db = communitySurveyResponseModel.preloadUser(db)
	}
	if stringsContains(o.Preload, "survey") {
		db = communitySurveyResponseModel.preloadSurvey(db)
	}
	return db
}

func (m *CommunitySurveyResponseModel) preloadUser(db *gorm.DB) *gorm.DB {
	return db.Preload("User")
}

func (m *CommunitySurveyResponseModel) preloadSurvey(db *gorm.DB) *gorm.DB {
	return db.Preload("Survey")
}

func (m *CommunitySurveyResponseModel) GetAllCommunitySurveyResponse(c *[]*CommunitySurveyResponse, opts *getAllCommunitySurveyResponseOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Run(db)
	}

	err = db.Find(c).Error
	return
}

func (o *getAllCommunitySurveyResponseOptions) Run(db *gorm.DB) *gorm.DB {
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
