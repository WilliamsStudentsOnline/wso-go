package models

import (
	"encoding/json"
	"errors"
)

// FlagStatus represents the possible states of a feature flag
type FlagStatus string

const (
	FlagEnabled  FlagStatus = "ENABLED"
	FlagDisabled FlagStatus = "DISABLED"
)

type FeatureFlags struct {
	BaseSchema

	Name   string     `gorm:"uniqueIndex;not null" json:"name"`
	Status FlagStatus `gorm:"not null" json:"status"`
}

func (*FeatureFlags) TableName() string {
	return "feature_flags"
}

func (fs *FlagStatus) UnmarshalJSON(b []byte) error {
	type F FlagStatus
	var status = (*F)(fs)
	err := json.Unmarshal(b, &status)
	if err != nil {
		return err
	}

	switch *fs {
	case FlagEnabled, FlagDisabled:
		return nil

	}
	return errors.New("invalid feature flag status")

}
