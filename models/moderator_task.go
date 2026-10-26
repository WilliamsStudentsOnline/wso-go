package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
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

type GetAllModeratorTasksOptions struct {
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	TaskType      *ModeratorTaskType `json:"taskType" form:"taskType" enums:",FACTRAK_REVIEW,NEW_USER_TAG,DELETE_USER_TAG"`
	TaskReference *string            `json:"taskReference" form:"taskReference"`
	// leave empty to filter only open tasks
	ResolvedAt *time.Time `json:"resolvedAt" form:"resolvedAt"`
	// omit or taskReference to group siblings; createdAt for newest-first
	OrderBy string `json:"orderBy" form:"orderBy" enums:",taskReference,createdAt"`

	Preload []string `json:"preload" form:"preload[]" enums:",requestingUser,resolvedByUser"`
}

func (o *GetAllModeratorTasksOptions) Order(db *gorm.DB) *gorm.DB {
	switch o.OrderBy {
	case "createdAt":
		return db.Order("moderator_tasks.created_at DESC", true)
	default:
		// taskReference (default): keep same-subject rows adjacent
		return db.Order("moderator_tasks.task_reference ASC, moderator_tasks.created_at ASC", true)
	}
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
	if o.TaskReference != nil {
		db = m.withTaskReference(*o.TaskReference)(db)
	}
	return m.withResolvedAt(o.ResolvedAt)(db)
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

func (m *ModeratorTaskModel) withTaskReference(taskReference string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("moderator_tasks.task_reference = ?", taskReference)
	}
}

func (m *ModeratorTaskModel) withResolvedAt(resolvedAt *time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if resolvedAt == nil {
			return db.Where("moderator_tasks.resolved_at IS NULL")
		}
		return db.Where("moderator_tasks.resolved_at IS NOT NULL")
	}
}

func (m *ModeratorTaskModel) GetAllTasks(tasks *[]*ModeratorTask, opts *GetAllModeratorTasksOptions) error {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Run(db)
	} else {
		db = db.Where("moderator_tasks.resolved_at IS NULL").
			Order("moderator_tasks.task_reference ASC, moderator_tasks.created_at ASC", true)
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

func (m *ModeratorTaskModel) GetOpenTasksByTaskReference(taskReference string, tasks *[]*ModeratorTask, preload ...string) error {
	opts := &GetAllModeratorTasksOptions{
		TaskReference: &taskReference,
		Preload:       preload,
	}
	return opts.Run(m.DB).Order("moderator_tasks.created_at ASC", true).Find(tasks).Error
}

func (m *ModeratorTaskModel) CreateTask(task *ModeratorTask) error {
	return m.DB.Create(task).Error
}

func (m *ModeratorTaskModel) ResolveOpenTasksByTaskReference(
	taskReference string,
	resolution ModeratorTaskResolution,
	resolvedByUserID uint,
	resolvedAt time.Time,
) (int64, error) {
	res := m.DB.Model(&ModeratorTask{}).
		Where("task_reference = ? AND resolved_at IS NULL", taskReference).
		Updates(map[string]interface{}{
			"resolved_at":         resolvedAt,
			"resolution":          resolution,
			"resolved_by_user_id": resolvedByUserID,
		})
	return res.RowsAffected, res.Error
}

func (m *ModeratorTaskModel) SoftDeleteOpenTasksByTaskReference(taskReference string) (int64, error) {
	res := m.DB.
		Where("task_reference = ? AND resolved_at IS NULL", taskReference).
		Delete(&ModeratorTask{})
	return res.RowsAffected, res.Error
}

func (m *ModeratorTaskModel) SoftDeleteTaskByID(id uint) error {
	res := m.DB.Where("id = ?", id).Delete(&ModeratorTask{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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
		TaskReference:    TaskReference(ModeratorTaskTypeFactrakReview, surveyID),
	}, nil
}

func DecodeUserTagFields(taskFieldsJSON string) (*UserTagTaskFields, error) {
	var fields UserTagTaskFields
	if err := json.Unmarshal([]byte(taskFieldsJSON), &fields); err != nil {
		return nil, err
	}
	return &fields, nil
}

func NewUserTagNameTask(requestingUserID *uint, targetUserID uint, tagName string) (*ModeratorTask, error) {
	tagName = strings.TrimSpace(tagName)
	if targetUserID == 0 || tagName == "" {
		return nil, errors.New("targetUserID and tagName are required")
	}
	fields, err := EncodeTaskFields(UserTagTaskFields{TargetUserID: targetUserID, TagName: tagName})
	if err != nil {
		return nil, err
	}
	return &ModeratorTask{
		RequestingUserID: requestingUserID,
		TaskType:         ModeratorTaskTypeNewUserTag,
		TaskFields:       fields,
		TaskReference:    UserTagNameTaskReference(targetUserID, tagName),
	}, nil
}

func UserTagNameTaskReference(targetUserID uint, tagName string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(tagName))))
	return TaskReference(ModeratorTaskTypeNewUserTag, targetUserID) + ":" + hex.EncodeToString(sum[:8])
}

func (m *ModeratorTaskModel) CountOpenTasksForRequester(requestingUserID uint, taskTypes ...ModeratorTaskType) (int, error) {
	db := m.DB.Model(&ModeratorTask{}).
		Where("requesting_user_id = ? AND resolved_at IS NULL", requestingUserID)
	if len(taskTypes) > 0 {
		db = db.Where("task_type IN (?)", taskTypes)
	}
	var count int
	err := db.Count(&count).Error
	return count, err
}
