// models_test.go
package models_test

import (
	"fmt"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestFeatureFlags_AddAndChangeFlags(t *testing.T) {

	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewFeatureFlagsModel(db, zaptest.NewLogger(t).Sugar())

	assert.NoError(m.AddFeatureFlag("Wiki", models.FlagEnabled))

	var count int64
	if err := db.Model(&models.FeatureFlags{}).Where("name = ?", "Wiki").Count(&count).Error; err != nil {
		fmt.Println("Error:", err)
	}

	assert.Equal(int64(1), count, "Expected feature flag to be added")

	var res FeatureFlags
	assert.NoError(m.GetFeatureFlagByName("Wiki", &res))
	assert.Equal(res.Name, "Wiki")
	assert.Equal(res.Status, models.FlagEnabled)
	assert.NoError(m.SetFeatureFlag("Wiki", models.FlagDisabled))
	var res2 FeatureFlags

	assert.NoError(m.GetFeatureFlagByName("Wiki", &res2))
	assert.Equal(res2.Status, models.FlagDisabled)

}

func TestFeatureFlags_GetAllFlags(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewFeatureFlagsModel(db, zaptest.NewLogger(t).Sugar())

	assert.NoError(m.AddFeatureFlag("t1", models.FlagEnabled))
	assert.NoError(m.AddFeatureFlag("t2", models.FlagEnabled))
	assert.NoError(m.AddFeatureFlag("t3", models.FlagEnabled))
	assert.NoError(m.AddFeatureFlag("t4", models.FlagDisabled))
	myNames := []string{"t1", "t2", "t3", "t4"}
	myStatus := []FlagStatus{models.FlagEnabled, models.FlagEnabled, models.FlagEnabled, models.FlagDisabled}

	var res []*FeatureFlags
	assert.NoError(m.GetAllFeatureFlags(&res))
	for i := range res {
		assert.Equal(myNames[i], res[i].Name)
		assert.Equal(myStatus[i], res[i].Status)

	}

}
