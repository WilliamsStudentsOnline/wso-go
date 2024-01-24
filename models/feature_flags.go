package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// FeatureFlagsModel represents the model for feature flags.
type FeatureFlagsModel struct {
	*BaseModel
}

// NewFeatureFlagsModel creates a new instance of FeatureFlagsModel.
func NewFeatureFlagsModel(db *gorm.DB, log *zap.SugaredLogger) *FeatureFlagsModel {
	return &FeatureFlagsModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// GetAllFeatureFlags retrieves all feature flags from the database.
func (m *FeatureFlagsModel) GetAllFeatureFlags(flags *[]*FeatureFlags) error {
	err := m.DB.Find(flags).Error
	return err
}

// GetFeatureFlagByName retrieves a feature flag by its name.
func (m *FeatureFlagsModel) GetFeatureFlagByName(name string, flag *FeatureFlags) (err error) {
	err = m.DB.Where("name = ?", name).First(flag).Error
	return
}

func (m *FeatureFlagsModel) AddFeatureFlag(name string, status FlagStatus) error {
	flag := FeatureFlags{
		Name:   name,
		Status: status,
	}

	// Check if a feature flag with the same name already exists
	var existingFlag FeatureFlags
	err := m.DB.Where("name = ?", name).First(&existingFlag).Error
	if err == nil {
		// A feature flag with the same name already exists
		return errors.New("feature flag with the same name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// An unexpected error occurred
		return err
	}

	// If no existing feature flag found, create a new one
	return m.DB.Create(&flag).Error
}

func (m *FeatureFlagsModel) SetFeatureFlag(name string, status FlagStatus) error {
	// Check if a feature flag with the same name already exists
	var existingFlag FeatureFlags
	err := m.DB.Where("name = ?", name).First(&existingFlag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Feature flag with the specified name doesn't exist, throw an error
			return errors.New("feature flag with the specified name does not exist")
		}
		// An unexpected error occurred
		return err
	}

	// A feature flag with the same name already exists, modify its status
	existingFlag.Status = status
	return m.DB.Save(&existingFlag).Error

}
