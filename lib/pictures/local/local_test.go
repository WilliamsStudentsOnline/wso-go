package local_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures/local"
	"github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestUserPhoto(t *testing.T) {
	assert := testify.New(t)
	logger := zaptest.NewLogger(t).Sugar()

	unixID := "ys5" // example unixID
	tempPath := t.TempDir()
	backend, err := local.NewBackend(tempPath, logger)
	assert.NoError(err)

	// check that no user image exists now
	exists, err := backend.DoesUserPhotoExists(unixID)
	if assert.NoError(err) {
		assert.False(exists)
	}

	// Test 1: save an image that's smaller than the max size
	img_smaller := test_utils.CreateTestImage(100, 100)
	err = backend.SaveUserPhotoBoth(unixID, img_smaller)
	assert.NoError(err)

	// Test 2: save an image that's larger than the max size
	img_larger := test_utils.CreateTestImage(1024, 1024)
	err = backend.SaveUserPhotoBoth(unixID, img_larger)
	assert.NoError(err)

	// Test 3: save an image that's not square
	img_not_square := test_utils.CreateTestImage(800, 200)
	err = backend.SaveUserPhotoBoth(unixID, img_not_square)
	assert.NoError(err)

	// check that the file exists now
	exists, err = backend.DoesUserPhotoExists(unixID)
	if assert.NoError(err) {
		assert.True(exists)
	}
}
