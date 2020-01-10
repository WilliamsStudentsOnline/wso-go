package models_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/db/seed"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestEphmatchProfileModel_DeleteProfile(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewEphmatchProfileModel(db, zaptest.NewLogger(t).Sugar())

	// Create user
	user := seed.GenerateStudent()
	assert.NoError(db.Create(user).Error)

	// Create Profile
	profile := &EphmatchProfile{
		UserID:      user.ID,
		Gender:      lib.StrToPtr(EphmatchProfileGenderNB),
		Description: lib.StrToPtr("hello world 123 abc foo bar"),
	}
	assert.NoError(db.Create(profile).Error)

	// Get the profile and assert deleted is false
	res := &EphmatchProfile{}
	assert.NoError(m.GetProfileByID(user.ID, res))
	assert.Equal(profile.Description, res.Description)
	assert.False(res.Deleted)

	// Delete the profile
	assert.NoError(m.DeleteProfile(res))

	// Assert does not exist for regular get profile
	res = &EphmatchProfile{}
	retErr := m.GetProfileByID(user.ID, res)
	assert.Error(retErr)
	assert.True(gorm.IsRecordNotFoundError(retErr))

	// *** Assert for all get profiles ***

	// Assert does not exist for self scoped get profile
	res = &EphmatchProfile{}
	retErr = nil
	retErr = m.GetSelfProfileByIDScopedNoDefault(user.ID, res)
	assert.Error(retErr)
	assert.True(gorm.IsRecordNotFoundError(retErr))

	// Assert does not exist for get all profiles
	var resArr []*EphmatchProfile
	assert.NoError(m.GetAllProfiles(&resArr, nil))
	assert.Empty(resArr)

	// Assert does exist and deleted is true for self get profile
	res = &EphmatchProfile{}
	assert.NoError(m.GetSelfProfileByID(user.ID, res))
	assert.Equal(profile.Description, res.Description)
	assert.NotNil(res.DeletedAt)
	assert.True(res.Deleted)
}

func TestEphmatchProfileModel_CreateOrUpdateProfileUnscoped(t *testing.T) {
	db := utils.SetupServiceTest(testify.New(t))

	m := NewEphmatchProfileModel(db, zaptest.NewLogger(t).Sugar())

	t.Run("Create Profile", func(t *testing.T) {
		assert := testify.New(t)

		user := seed.GenerateStudent()
		assert.NoError(db.Create(user).Error)

		// Create Profile
		newProfile := EphmatchProfile{
			UserID:      user.ID,
			Gender:      lib.StrToPtr(EphmatchProfileGenderNB),
			Description: lib.StrToPtr("hello world 123 abc foo bar"),
		}

		var profile EphmatchProfile
		assert.NoError(m.CreateOrUpdateProfileUnscoped(user.ID, newProfile, &profile))

		var res EphmatchProfile
		assert.NoError(db.Model(EphmatchProfile{}).Where(EphmatchProfile{UserID: newProfile.UserID}).First(&res).Error)

		assert.Equal(profile.ID, res.ID)
		assert.Equal(profile.Gender, res.Gender)
		assert.Equal(profile.Description, res.Description)
		assert.Nil(res.DeletedAt)
		assert.Equal(newProfile.Gender, res.Gender)
		assert.Equal(newProfile.Description, res.Description)
	})

	t.Run("Create Profile with empty description and gender", func(t *testing.T) {
		assert := testify.New(t)

		user := seed.GenerateStudent()
		assert.NoError(db.Create(user).Error)

		// Create Profile
		newProfile := EphmatchProfile{
			UserID: user.ID,
		}

		var profile EphmatchProfile
		assert.NoError(m.CreateOrUpdateProfileUnscoped(user.ID, newProfile, &profile))

		var res EphmatchProfile
		assert.NoError(db.Model(EphmatchProfile{}).Where(EphmatchProfile{UserID: newProfile.UserID}).First(&res).Error)

		assert.Equal(profile.ID, res.ID)
		assert.Equal(profile.Gender, res.Gender)
		assert.Equal(profile.Description, res.Description)
		assert.Nil(res.DeletedAt)
		assert.Equal(newProfile.Gender, res.Gender)
		assert.Equal(newProfile.Description, res.Description)
	})

	t.Run("Update deleted profile", func(t *testing.T) {
		assert := testify.New(t)

		user := seed.GenerateStudent()
		assert.NoError(db.Create(user).Error)

		// Create Profile
		newProfile := EphmatchProfile{
			UserID:      user.ID,
			Gender:      lib.StrToPtr(EphmatchProfileGenderNB),
			Description: lib.StrToPtr("hello world 123 abc foo bar"),
		}

		var profile EphmatchProfile
		assert.NoError(m.CreateOrUpdateProfileUnscoped(user.ID, newProfile, &profile))

		// Delete it
		assert.NoError(m.DeleteProfile(&profile))

		profile = EphmatchProfile{}
		newProfileDeleted := EphmatchProfile{
			UserID: user.ID,
		}
		assert.NoError(m.CreateOrUpdateProfileUnscoped(user.ID, newProfileDeleted, &profile))

		var res EphmatchProfile
		assert.NoError(db.Model(EphmatchProfile{}).Where(EphmatchProfile{UserID: newProfile.UserID}).First(&res).Error)

		assert.Equal(profile.ID, res.ID)
		assert.Equal(profile.Gender, res.Gender)
		assert.Equal(profile.Description, res.Description)
		assert.Nil(res.DeletedAt)
		assert.Equal(newProfile.Gender, res.Gender)
		assert.Equal(newProfile.Description, res.Description)
	})

}
