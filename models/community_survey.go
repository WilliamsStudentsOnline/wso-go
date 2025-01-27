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

type GetAllCommunitySurveyOptions struct {
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

func (o *GetAllCommunitySurveyOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("community_survey.created_at DESC")
}

func (o *GetAllCommunitySurveyOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

func (o *GetAllCommunitySurveyOptions) Preloader(db *gorm.DB) *gorm.DB {
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

func (o *GetAllCommunitySurveyOptions) Run(db *gorm.DB) *gorm.DB {
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

func (m *CommunitySurveyModel) GetAllCommunitySurvey(c *[]*CommunitySurvey, opts *GetAllCommunitySurveyOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Run(db)
	}

	err = db.Find(c).Error
	return
}

func (m *CommunitySurveyModel) CreateCommunitySurvey(c *CommunitySurvey) (err error) {
	err = m.DB.Create(c).Error
	if err != nil {
		return err
	}
	err = m.DB.Find(c).Error
	return
}

func (m *CommunitySurveyModel) GetCommunitySurveyByID(id uint, c *CommunitySurvey) (err error) {
	err = m.DB.First(c, id).Error
	return
}

func (m *CommunitySurveyModel) UpdateCommunitySurvey(c *CommunitySurvey) (err error) {
	err = m.DB.Save(c).Error
	return
}

func (m *CommunitySurveyModel) DeleteCommunitySurvey(c *CommunitySurvey) (err error) {
	err = m.DB.Delete(c).Error
	return
}
