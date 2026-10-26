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

type ModeratorTaskGroup struct {
	TaskReference string                   `json:"taskReference"`
	TaskType      models.ModeratorTaskType `json:"taskType"`
	Tasks         []*ModeratorTaskResponse `json:"tasks"`
	Context       interface{}              `json:"context,omitempty"`
}

type FactrakReviewTaskContext struct {
	Survey *models.FactrakSurvey `json:"survey,omitempty"`
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

func groupTasks(tasks []*models.ModeratorTask) []*ModeratorTaskGroup {
	order := make([]string, 0)
	byKey := map[string]*ModeratorTaskGroup{}

	for _, t := range tasks {
		g, ok := byKey[t.TaskReference]
		if !ok {
			g = &ModeratorTaskGroup{
				TaskReference: t.TaskReference,
				TaskType:      t.TaskType,
				Tasks:         []*ModeratorTaskResponse{},
			}
			byKey[t.TaskReference] = g
			order = append(order, t.TaskReference)
		}
		g.Tasks = append(g.Tasks, taskToResponse(t))
	}

	out := make([]*ModeratorTaskGroup, 0, len(order))
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
