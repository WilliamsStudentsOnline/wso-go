package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ModeratorTaskType string

const (
	ModeratorTaskTypeUndefined      ModeratorTaskType = ""
	ModeratorTaskTypeFactrakReview  ModeratorTaskType = "FACTRAK_REVIEW"
	ModeratorTaskTypeNewUserTag     ModeratorTaskType = "NEW_USER_TAG"
	ModeratorTaskTypeDeleteUserTag  ModeratorTaskType = "DELETE_USER_TAG"
)

// MaxOpenUserTagTasksPerRequester caps concurrent open NEW/DELETE user-tag requests per user
const MaxOpenUserTagTasksPerRequester = 3

type ModeratorTaskResolution string

const (
	ModeratorTaskResolutionUndefined ModeratorTaskResolution = ""
	ModeratorTaskResolutionApproved  ModeratorTaskResolution = "APPROVED"
	ModeratorTaskResolutionRejected  ModeratorTaskResolution = "REJECTED"
)

type FactrakReviewTaskFields struct {
	FactrakSurveyID uint `json:"factrakSurveyID"`
}

type UserTagTaskFields struct {
	TargetUserID uint `json:"targetUserID"`
	TagID        uint `json:"tagID"`
}

type ModeratorTask struct {
	BaseSchema

	RequestingUserID *uint `gorm:"index:index_moderator_tasks_on_requesting_user_id" json:"requestingUserID"`
	RequestingUser   *User `json:"requestingUser,omitempty"`

	TaskType   ModeratorTaskType `gorm:"index:index_moderator_tasks_on_task_type;not null" json:"taskType" enums:",FACTRAK_REVIEW,NEW_USER_TAG,DELETE_USER_TAG"`
	TaskFields string            `gorm:"column:task_fields;type:text;not null" json:"-"`

	TaskReference string `gorm:"column:task_reference;index:index_moderator_tasks_on_task_reference;not null" json:"taskReference"`

	ResolvedAt       *time.Time               `gorm:"index:index_moderator_tasks_on_resolved_at" json:"resolvedAt"`
	Resolution       *ModeratorTaskResolution `json:"resolution" enums:",APPROVED,REJECTED"`
	ResolvedByUserID *uint                    `gorm:"index:index_moderator_tasks_on_resolved_by_user_id" json:"resolvedByUserID"`
	ResolvedByUser   *User                    `json:"resolvedByUser,omitempty"`
}

func (*ModeratorTask) TableName() string {
	return "moderator_tasks"
}

func TaskReference(taskType ModeratorTaskType, ids ...uint) string {
	parts := make([]string, 0, 1+len(ids))
	parts = append(parts, string(taskType))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(uint64(id), 10))
	}
	return strings.Join(parts, ":")
}

func (t *ModeratorTaskType) UnmarshalJSON(b []byte) error {
	type alias ModeratorTaskType
	var r = (*alias)(t)
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}
	switch *t {
	case ModeratorTaskTypeUndefined, ModeratorTaskTypeFactrakReview, ModeratorTaskTypeNewUserTag, ModeratorTaskTypeDeleteUserTag:
		return nil
	}
	return errors.New("invalid moderator task type")
}

func (r *ModeratorTaskResolution) UnmarshalJSON(b []byte) error {
	type alias ModeratorTaskResolution
	var v = (*alias)(r)
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch *r {
	case ModeratorTaskResolutionUndefined, ModeratorTaskResolutionApproved, ModeratorTaskResolutionRejected:
		return nil
	}
	return errors.New("invalid moderator task resolution")
}

func (t ModeratorTaskType) Validate() error {
	switch t {
	case ModeratorTaskTypeFactrakReview, ModeratorTaskTypeNewUserTag, ModeratorTaskTypeDeleteUserTag:
		return nil
	case ModeratorTaskTypeUndefined:
		return errors.New("task type is required")
	default:
		return fmt.Errorf("unknown task type: %s", t)
	}
}

func (t ModeratorTaskType) IsUserTagTask() bool {
	return t == ModeratorTaskTypeNewUserTag || t == ModeratorTaskTypeDeleteUserTag
}
