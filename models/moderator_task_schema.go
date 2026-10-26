package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ModeratorTaskType is an extensible string enum for moderator console work items.
type ModeratorTaskType string

const (
	ModeratorTaskTypeUndefined     ModeratorTaskType = ""
	ModeratorTaskTypeFactrakReview ModeratorTaskType = "FACTRAK_REVIEW"
)

// ModeratorTaskResolution is set when a task group is closed.
type ModeratorTaskResolution string

const (
	ModeratorTaskResolutionUndefined ModeratorTaskResolution = ""
	ModeratorTaskResolutionApproved  ModeratorTaskResolution = "APPROVED"
	ModeratorTaskResolutionRejected  ModeratorTaskResolution = "REJECTED"
)

// FactrakReviewTaskFields is the task_fields payload for FACTRAK_REVIEW.
type FactrakReviewTaskFields struct {
	FactrakSurveyID uint `json:"factrakSurveyID"`
}

// ModeratorTask is a single moderation request (one row per flag/report).
type ModeratorTask struct {
	BaseSchema

	RequestingUserID *uint `gorm:"index:index_moderator_tasks_on_requesting_user_id" json:"requestingUserID"`
	RequestingUser   *User `json:"requestingUser,omitempty"`

	TaskType   ModeratorTaskType `gorm:"index:index_moderator_tasks_on_task_type;not null" json:"taskType" enums:",FACTRAK_REVIEW"`
	TaskFields string            `gorm:"column:task_fields;type:text;not null" json:"-"`

	// SubjectKey groups related open tasks (e.g. factrak_survey:123). Set at create.
	SubjectKey string `gorm:"index:index_moderator_tasks_on_subject_key;not null" json:"subjectKey"`

	ResolvedAt       *time.Time               `gorm:"index:index_moderator_tasks_on_resolved_at" json:"resolvedAt"`
	Resolution       *ModeratorTaskResolution `json:"resolution" enums:",APPROVED,REJECTED"`
	ResolvedByUserID *uint                    `gorm:"index:index_moderator_tasks_on_resolved_by_user_id" json:"resolvedByUserID"`
	ResolvedByUser   *User                    `json:"resolvedByUser,omitempty"`
}

func (*ModeratorTask) TableName() string {
	return "moderator_tasks"
}

// SubjectKeyForFactrakSurvey returns the canonical subject_key for a Factrak survey.
func SubjectKeyForFactrakSurvey(surveyID uint) string {
	return fmt.Sprintf("factrak_survey:%d", surveyID)
}

// SubjectKeyForTaskType builds a subject_key from task type + fields JSON.
func SubjectKeyForTaskType(taskType ModeratorTaskType, fieldsJSON string) (string, error) {
	switch taskType {
	case ModeratorTaskTypeFactrakReview:
		var fields FactrakReviewTaskFields
		if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
			return "", err
		}
		if fields.FactrakSurveyID == 0 {
			return "", errors.New("factrakSurveyID is required")
		}
		return SubjectKeyForFactrakSurvey(fields.FactrakSurveyID), nil
	case ModeratorTaskTypeUndefined:
		return "", errors.New("task type is required")
	default:
		return "", fmt.Errorf("unknown task type: %s", taskType)
	}
}

func (t *ModeratorTaskType) UnmarshalJSON(b []byte) error {
	type alias ModeratorTaskType
	var r = (*alias)(t)
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}
	switch *t {
	case ModeratorTaskTypeUndefined, ModeratorTaskTypeFactrakReview:
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
