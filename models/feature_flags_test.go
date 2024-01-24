// models_test.go
package models

import (
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/models"
)

func TestAdd(t *testing.T) {

	f1 := FeatureFlags{
		Name:   "About",
		Status: DISABLED,
	}
	assert.NoError(db.Create(&p1).Error)

	assert.NoError(AddFeatureFlag("Wiki", ENABLED))

	// Add more test cases as needed
}
