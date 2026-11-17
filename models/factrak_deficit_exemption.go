package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

var ErrFactrakDeficitExemptionExists = errors.New("deficit exemption already requested for this semester")

type FactrakDeficitExemptionModel struct {
	*BaseModel
}

func NewFactrakDeficitExemptionModel(db *gorm.DB, log *zap.SugaredLogger) *FactrakDeficitExemptionModel {
	return &FactrakDeficitExemptionModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Switches October 1 and March 1, when update-all-factrak-survey-deficits runs; change both together
func FactrakDeficitSemester(now time.Time) (season string, year int) {
	switch month := now.Month(); {
	case month >= time.October:
		return FactrakSurveySemesterSeasonFall, now.Year()
	case month <= time.February:
		return FactrakSurveySemesterSeasonFall, now.Year() - 1
	default:
		return FactrakSurveySemesterSeasonSpring, now.Year()
	}
}

func FactrakDeficitSemesterStart(now time.Time) time.Time {
	switch month := now.Month(); {
	case month >= time.October:
		return time.Date(now.Year(), time.October, 1, 0, 0, 0, 0, now.Location())
	case month <= time.February:
		return time.Date(now.Year()-1, time.October, 1, 0, 0, 0, 0, now.Location())
	default:
		return time.Date(now.Year(), time.March, 1, 0, 0, 0, 0, now.Location())
	}
}

func (m *FactrakDeficitExemptionModel) GetExemptionByID(id uint, e *FactrakDeficitExemption) error {
	return m.DB.Preload("User").First(e, id).Error
}

func (m *FactrakDeficitExemptionModel) HasApprovedExemption(userID uint, now time.Time) (bool, error) {
	season, year := FactrakDeficitSemester(now)
	var count int
	err := m.DB.Model(&FactrakDeficitExemption{}).
		Where("user_id = ? AND semester_season = ? AND semester_year = ? AND status = ?",
			userID, season, year, FactrakDeficitExemptionStatusApproved).
		Count(&count).Error
	return count > 0, err
}

func (m *FactrakDeficitExemptionModel) CreateExemptionRequest(userID uint, reason string, now time.Time) (*FactrakDeficitExemption, error) {
	season, year := FactrakDeficitSemester(now)

	tx := m.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.RollbackUnlessCommitted()

	// Lock the user's row so two requests at once can't both pass the check below
	var user User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Select("id").First(&user, userID).Error; err != nil {
		return nil, err
	}

	var count int
	err := tx.Model(&FactrakDeficitExemption{}).
		Where("user_id = ? AND semester_season = ? AND semester_year = ? AND status IN (?)", userID, season, year,
			[]FactrakDeficitExemptionStatus{FactrakDeficitExemptionStatusPending, FactrakDeficitExemptionStatusApproved}).
		Count(&count).Error
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrFactrakDeficitExemptionExists
	}

	e := &FactrakDeficitExemption{
		UserID:         userID,
		SemesterSeason: season,
		SemesterYear:   year,
		Reason:         reason,
		Status:         FactrakDeficitExemptionStatusPending,
	}
	if err := tx.Create(e).Error; err != nil {
		return nil, err
	}

	task, err := NewFactrakDeficitExemptionTask(&userID, e.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Create(task).Error; err != nil {
		return nil, err
	}

	return e, tx.Commit().Error
}

func (m *FactrakDeficitExemptionModel) SetExemptionStatus(id uint, status FactrakDeficitExemptionStatus) error {
	return m.DB.Model(&FactrakDeficitExemption{}).Where("id = ?", id).Update("status", status).Error
}
