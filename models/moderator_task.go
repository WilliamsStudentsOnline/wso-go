package models

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type ModeratorTaskModel struct {
	*BaseModel
}

func NewModeratorTaskModel(db *gorm.DB, log *zap.SugaredLogger) *ModeratorTaskModel {
	return &ModeratorTaskModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// GetAllModeratorTasksOptions filters/paginates moderator_tasks
// Resolved nil defaults to open only (resolved_at IS NULL)
type GetAllModeratorTasksOptions struct {
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	TaskType   *ModeratorTaskType `json:"taskType" form:"taskType" enums:",FACTRAK_REVIEW"`
	SubjectKey *string            `json:"subjectKey" form:"subjectKey"`
	Resolved   *bool              `json:"resolved" form:"resolved"`

	Preload []string `json:"preload" form:"preload[]" enums:",requestingUser,resolvedByUser"`
}

func (o *GetAllModeratorTasksOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("moderator_tasks.created_at DESC", true)
}

func (o *GetAllModeratorTasksOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

func (o *GetAllModeratorTasksOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}
	m := NewModeratorTaskModel(nil, nil)
	if stringsContains(o.Preload, "requestingUser") {
		db = m.preloadRequestingUser(db)
	}
	if stringsContains(o.Preload, "resolvedByUser") {
		db = m.preloadResolvedByUser(db)
	}
	return db
}

func (o *GetAllModeratorTasksOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)
	m := NewModeratorTaskModel(db.New(), nil)

	if o.TaskType != nil {
		db = m.withTaskType(*o.TaskType)(db)
	}
	if o.SubjectKey != nil {
		db = m.withSubjectKey(*o.SubjectKey)(db)
	}

	wantResolved := false
	if o.Resolved != nil {
		wantResolved = *o.Resolved
	}
	return m.withResolved(wantResolved)(db)
}

func (m *ModeratorTaskModel) preloadRequestingUser(db *gorm.DB) *gorm.DB {
	return db.Preload("RequestingUser")
}

func (m *ModeratorTaskModel) preloadResolvedByUser(db *gorm.DB) *gorm.DB {
	return db.Preload("ResolvedByUser")
}

func (m *ModeratorTaskModel) withTaskType(taskType ModeratorTaskType) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("moderator_tasks.task_type = ?", taskType)
	}
}

func (m *ModeratorTaskModel) withSubjectKey(subjectKey string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("moderator_tasks.subject_key = ?", subjectKey)
	}
}

func (m *ModeratorTaskModel) withResolved(resolved bool) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if resolved {
			return db.Where("moderator_tasks.resolved_at IS NOT NULL")
		}
		return db.Where("moderator_tasks.resolved_at IS NULL")
	}
}

func (m *ModeratorTaskModel) GetAllTasks(tasks *[]*ModeratorTask, opts *GetAllModeratorTasksOptions) error {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Run(db)
	} else {
		db = db.Where("moderator_tasks.resolved_at IS NULL").Order("moderator_tasks.created_at DESC", true)
	}
	return db.Find(tasks).Error
}

func (m *ModeratorTaskModel) CountAllTasks(opts *GetAllModeratorTasksOptions) (count int, err error) {
	db := m.DB.Model(&ModeratorTask{})
	if opts != nil {
		db = opts.Run(db)
	} else {
		db = db.Where("moderator_tasks.resolved_at IS NULL")
	}
	err = db.Count(&count).Error
	return
}

func (m *ModeratorTaskModel) GetTaskByID(id uint, task *ModeratorTask, preload ...string) error {
	opts := &GetAllModeratorTasksOptions{Preload: preload}
	return opts.Preloader(m.DB).First(task, id).Error
}

func (m *ModeratorTaskModel) GetOpenTasksBySubjectKey(subjectKey string, tasks *[]*ModeratorTask, preload ...string) error {
	falseVal := false
	opts := &GetAllModeratorTasksOptions{
		SubjectKey: &subjectKey,
		Resolved:   &falseVal,
		Preload:    preload,
	}
	return opts.Run(m.DB).Order("moderator_tasks.created_at ASC", true).Find(tasks).Error
}

func (m *ModeratorTaskModel) CreateTask(task *ModeratorTask) error {
	return m.DB.Create(task).Error
}

func (m *ModeratorTaskModel) ResolveOpenTasksBySubjectKey(
	subjectKey string,
	resolution ModeratorTaskResolution,
	resolvedByUserID uint,
	resolvedAt time.Time,
) (int64, error) {
	res := m.DB.Model(&ModeratorTask{}).
		Where("subject_key = ? AND resolved_at IS NULL", subjectKey).
		Updates(map[string]interface{}{
			"resolved_at":         resolvedAt,
			"resolution":          resolution,
			"resolved_by_user_id": resolvedByUserID,
		})
	return res.RowsAffected, res.Error
}

func (m *ModeratorTaskModel) SoftDeleteOpenTasksBySubjectKey(subjectKey string) (int64, error) {
	res := m.DB.
		Where("subject_key = ? AND resolved_at IS NULL", subjectKey).
		Delete(&ModeratorTask{})
	return res.RowsAffected, res.Error
}

func EncodeTaskFields(fields interface{}) (string, error) {
	b, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func DecodeFactrakReviewFields(taskFieldsJSON string) (*FactrakReviewTaskFields, error) {
	var fields FactrakReviewTaskFields
	if err := json.Unmarshal([]byte(taskFieldsJSON), &fields); err != nil {
		return nil, err
	}
	return &fields, nil
}

func NewFactrakReviewTask(requestingUserID *uint, surveyID uint) (*ModeratorTask, error) {
	fields, err := EncodeTaskFields(FactrakReviewTaskFields{FactrakSurveyID: surveyID})
	if err != nil {
		return nil, err
	}
	return &ModeratorTask{
		RequestingUserID: requestingUserID,
		TaskType:         ModeratorTaskTypeFactrakReview,
		TaskFields:       fields,
		SubjectKey:       SubjectKey(ModeratorTaskTypeFactrakReview, surveyID),
	}, nil
}
