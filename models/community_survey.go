package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type CommunitySurveyModel struct {
	*BaseModel
}

func NewCommunitySurveyModel(db *gorm.DB, log *zap.SugaredLogger) *CommunitySurveyModel {
	return &CommunitySurveyModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type getAllCommunitySurveyOptions struct {
	//offset is ignored unless limit is applied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	//filters
	SurveyID    *uint `json:"surveyID" form:"surveyID"`
	UserID      *uint `json:"userID" form:"userID"`
	AnswerCount *uint `json:"answerCount" form:"answerCount"`

	//Preloading: user,survey
	Preload []string `json:"preload" form:"preload[]"`
}

func (o *getAllCommunitySurveyOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("community_survey.created_at DESC")
}

func (o *getAllCommunitySurveyOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

func (o *getAllCommunitySurveyOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}
	CommunitySurveyModel := NewCommunitySurveyModel(nil, nil)
	if stringsContains(o.Preload, "affiliate") {
		db = CommunitySurveyModel.preloadUser(db)
	}
	if stringsContains(o.Preload, "survey") {
		db = CommunitySurveyModel.preloadSurvey(db)
	}
	return db
}

func (m *CommunitySurveyModel) preloadUser(db *gorm.DB) *gorm.DB {
	return db.Preload("Affiliate")
}

func (m *CommunitySurveyModel) preloadSurvey(db *gorm.DB) *gorm.DB {
	return db.Preload("Survey")
}

func (o *getAllCommunitySurveyOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)
	m := NewCommunitySurveyModel(db.New(), nil)
	if o.UserID != nil {
		db = m.withUser(*o.UserID)(db)
	}
	if o.SurveyID != nil {
		db = m.withSurvey(*o.SurveyID)(db)
	}
	if o.AnswerCount != nil {
		db = m.withAnswerCount(*o.AnswerCount)(db)
	}
	return db
}

func (m *CommunitySurveyModel) withUser(userID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

func (m *CommunitySurveyModel) withSurvey(surveyID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("survey_id = ?", surveyID)
	}
}

func (M *CommunitySurveyModel) withAnswerCount(answerCount uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("answer_count = ?", answerCount)
	}
}

func (m *CommunitySurveyModel) GetAllCommunitySurvey(c *[]*CommunitySurvey, opts *getAllCommunitySurveyOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Run(db)
	}

	err = db.Find(c).Error
	return
}
