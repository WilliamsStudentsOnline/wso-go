package models

import (
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
func (m *FeatureFlagsModel) GetFeatureFlagByName(name string) (*FeatureFlags, error) {
	var flag FeatureFlags

	if err := m.DB.Where("name = ?", name).First(&flag).Error; err != nil {
		return nil, err
	}

	return &flag, nil
}

func (m *FeatureFlagsModel) AddFeatureFlag(name string, status FlagStatus) error {
	flag := FeatureFlags{
		Name:   name,
		Status: status,
	}

	return m.DB.Create(&flag).Error
}

// SetFeatureFlag sets the status of the feature flag.
func (ff *FeatureFlags) SetFeatureFlag(status FlagStatus) {
	ff.Status = status
}
