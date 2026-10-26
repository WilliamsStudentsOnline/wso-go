package admin

import (
	"encoding/json"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
)

// ModeratorTaskDTO is the API shape for a single moderator_tasks row.
type ModeratorTaskDTO struct {
	ID               uint                            `json:"id"`
	CreatedAt        time.Time                       `json:"createdAt"`
	UpdatedAt        time.Time                       `json:"updatedAt"`
	RequestingUserID *uint                           `json:"requestingUserID"`
	RequestingUser   *models.User                    `json:"requestingUser,omitempty"`
	TaskType         models.ModeratorTaskType        `json:"taskType"`
	TaskFields       map[string]interface{}          `json:"taskFields"`
	SubjectKey       string                          `json:"subjectKey"`
	ResolvedAt       *time.Time                      `json:"resolvedAt"`
	Resolution       *models.ModeratorTaskResolution `json:"resolution"`
	ResolvedByUserID *uint                           `json:"resolvedByUserID"`
	ResolvedByUser   *models.User                    `json:"resolvedByUser,omitempty"`
}

// ModeratorTaskGroupDTO groups open/closed flags that share a subject_key.
type ModeratorTaskGroupDTO struct {
	SubjectKey string                 `json:"subjectKey"`
	TaskType   models.ModeratorTaskType `json:"taskType"`
	Tasks      []*ModeratorTaskDTO    `json:"tasks"`
	Context    interface{}            `json:"context,omitempty"`
}

// FactrakReviewTaskContext is enrich-on-read context for FACTRAK_REVIEW groups.
// Survey author identity is cleared (anonymous).
type FactrakReviewTaskContext struct {
	Survey *models.FactrakSurvey `json:"survey,omitempty"`
}

// ResolveTasksBody is the body for POST .../resolve.
type ResolveTasksBody struct {
	Resolution models.ModeratorTaskResolution `json:"resolution" binding:"required" enums:"APPROVED,REJECTED"`
}

// SubjectKeyResponse is returned by GET /tasks/subject-key.
type SubjectKeyResponse struct {
	SubjectKey string `json:"subjectKey"`
}

func taskToDTO(t *models.ModeratorTask) *ModeratorTaskDTO {
	fields := map[string]interface{}{}
	if t.TaskFields != "" {
		_ = json.Unmarshal([]byte(t.TaskFields), &fields)
	}
	return &ModeratorTaskDTO{
		ID:               t.ID,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
		RequestingUserID: t.RequestingUserID,
		RequestingUser:   t.RequestingUser,
		TaskType:         t.TaskType,
		TaskFields:       fields,
		SubjectKey:       t.SubjectKey,
		ResolvedAt:       t.ResolvedAt,
		Resolution:       t.Resolution,
		ResolvedByUserID: t.ResolvedByUserID,
		ResolvedByUser:   t.ResolvedByUser,
	}
}

func groupTasks(tasks []*models.ModeratorTask) []*ModeratorTaskGroupDTO {
	order := make([]string, 0)
	byKey := map[string]*ModeratorTaskGroupDTO{}

	for _, t := range tasks {
		g, ok := byKey[t.SubjectKey]
		if !ok {
			g = &ModeratorTaskGroupDTO{
				SubjectKey: t.SubjectKey,
				TaskType:   t.TaskType,
				Tasks:      []*ModeratorTaskDTO{},
			}
			byKey[t.SubjectKey] = g
			order = append(order, t.SubjectKey)
		}
		g.Tasks = append(g.Tasks, taskToDTO(t))
	}

	out := make([]*ModeratorTaskGroupDTO, 0, len(order))
	for _, key := range order {
		out = append(out, byKey[key])
	}
	return out
}

func anonymizeSurveyAuthor(survey *models.FactrakSurvey) {
	if survey == nil {
		return
	}
	survey.UserID = 0
	survey.User = nil
}
