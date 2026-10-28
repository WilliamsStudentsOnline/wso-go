package models

import (
	"encoding/json"
	"errors"
	"fmt"
)

type FeedbackPriority string

const (
	FeedbackPriorityUndefined FeedbackPriority = ""
	FeedbackPriorityNone      FeedbackPriority = "none"
	FeedbackPriorityLow       FeedbackPriority = "priority_low"
	FeedbackPriorityMed       FeedbackPriority = "priority_med"
	FeedbackPriorityHigh      FeedbackPriority = "priority_high"
	FeedbackPriorityResolved  FeedbackPriority = "resolved"
	FeedbackPriorityDismissed FeedbackPriority = "dismissed"
)

type Feedback struct {
	BaseSchema

	UserID *uint `gorm:"index:index_feedback_on_user_id" json:"userID"`
	User   *User `json:"user,omitempty"`

	PreferredName *string          `json:"preferredName"`
	Priority      FeedbackPriority `gorm:"index:index_feedback_on_priority;not null;default:'none'" json:"priority" enums:",none,priority_low,priority_med,priority_high,resolved,dismissed"`
	URL           *string          `json:"url"`
	Notes         string           `gorm:"size:65535;not null" json:"notes"`
	AdminNotes    *string          `gorm:"size:65535" json:"adminNotes"`
	Issue         *string          `json:"issue"`
}

func (*Feedback) TableName() string {
	return "feedback"
}

func (p *FeedbackPriority) UnmarshalJSON(b []byte) error {
	type alias FeedbackPriority
	var v = (*alias)(p)
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch *p {
	case FeedbackPriorityUndefined, FeedbackPriorityNone, FeedbackPriorityLow, FeedbackPriorityMed,
		FeedbackPriorityHigh, FeedbackPriorityResolved, FeedbackPriorityDismissed:
		return nil
	}
	return errors.New("invalid feedback priority")
}

func (p FeedbackPriority) Validate() error {
	switch p {
	case FeedbackPriorityNone, FeedbackPriorityLow, FeedbackPriorityMed, FeedbackPriorityHigh,
		FeedbackPriorityResolved, FeedbackPriorityDismissed:
		return nil
	case FeedbackPriorityUndefined:
		return errors.New("priority is required")
	default:
		return fmt.Errorf("unknown feedback priority: %s", p)
	}
}
