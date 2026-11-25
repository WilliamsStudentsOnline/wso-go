package admin

import (
	"encoding/json"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
)

type ModeratorTaskResponse struct {
	ID               uint                            `json:"id"`
	CreatedAt        time.Time                       `json:"createdAt"`
	UpdatedAt        time.Time                       `json:"updatedAt"`
	RequestingUserID *uint                           `json:"requestingUserID"`
	RequestingUser   *models.User                    `json:"requestingUser,omitempty"`
	TaskType         models.ModeratorTaskType        `json:"taskType"`
	TaskFields       map[string]interface{}          `json:"taskFields"`
	TaskReference    string                          `json:"taskReference"`
	ResolvedAt       *time.Time                      `json:"resolvedAt"`
	Resolution       *models.ModeratorTaskResolution `json:"resolution"`
	ResolvedByUserID *uint                           `json:"resolvedByUserID"`
	ResolvedByUser   *models.User                    `json:"resolvedByUser,omitempty"`
	Context          interface{}                     `json:"context,omitempty"`
}

type FactrakReviewTaskContext struct {
	Survey *models.FactrakSurvey `json:"survey,omitempty"`
}

type UserTagTaskContext struct {
	TargetUserID uint        `json:"targetUserID"`
	TagName      string      `json:"tagName,omitempty"`
	Tag          *models.Tag `json:"tag,omitempty"`
}

type BoardThreadReviewTaskContext struct {
	Thread *models.BoardThread `json:"thread,omitempty"`
}

type BoardPostReviewTaskContext struct {
	Post *models.BoardPost `json:"post,omitempty"`
}

type UserPhotoReviewTaskContext struct {
	TargetUserID uint   `json:"targetUserID"`
	Name         string `json:"name,omitempty"`
	UnixID       string `json:"unixID,omitempty"`
}

type FactrakDeficitExemptionTaskContext struct {
	Exemption *FactrakDeficitExemptionResponse `json:"exemption,omitempty"`
}

type FactrakDeficitExemptionResponse struct {
	ID             uint                                 `json:"id"`
	CreatedAt      time.Time                            `json:"createdAt"`
	UserID         uint                                 `json:"userID"`
	User           *models.User                         `json:"user,omitempty"`
	SemesterSeason string                               `json:"semesterSeason" enums:"fall,spring"`
	SemesterYear   int                                  `json:"semesterYear"`
	Reason         string                               `json:"reason"`
	Status         models.FactrakDeficitExemptionStatus `json:"status" enums:"pending,approved,rejected"`
}

type ResolveTasksBody struct {
	Resolution models.ModeratorTaskResolution `json:"resolution" binding:"required" enums:"APPROVED,REJECTED"`
}

func taskToResponse(t *models.ModeratorTask) *ModeratorTaskResponse {
	fields := map[string]interface{}{}
	if t.TaskFields != "" {
		_ = json.Unmarshal([]byte(t.TaskFields), &fields)
	}
	return &ModeratorTaskResponse{
		ID:               t.ID,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
		RequestingUserID: t.RequestingUserID,
		RequestingUser:   t.RequestingUser,
		TaskType:         t.TaskType,
		TaskFields:       fields,
		TaskReference:    t.TaskReference,
		ResolvedAt:       t.ResolvedAt,
		Resolution:       t.Resolution,
		ResolvedByUserID: t.ResolvedByUserID,
		ResolvedByUser:   t.ResolvedByUser,
	}
}

func exemptionToResponse(e *models.FactrakDeficitExemption) *FactrakDeficitExemptionResponse {
	return &FactrakDeficitExemptionResponse{
		ID:             e.ID,
		CreatedAt:      e.CreatedAt,
		UserID:         e.UserID,
		User:           e.User,
		SemesterSeason: e.SemesterSeason,
		SemesterYear:   e.SemesterYear,
		Reason:         e.Reason,
		Status:         e.Status,
	}
}

func anonymizeSurveyAuthor(survey *models.FactrakSurvey) {
	if survey == nil {
		return
	}
	survey.UserID = 0
	survey.User = nil
}
