package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures/local"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	admin "github.com/WilliamsStudentsOnline/wso-go/services/admin"
	. "github.com/WilliamsStudentsOnline/wso-go/services/user"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func photoFlagSetup(t *testing.T) (*testify.Assertions, *gorm.DB, *config.Config, *local.Backend) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()
	cfg.PictureBackend = pictures.PictureBackendLocal
	cfg.PictureLocalPath = t.TempDir()

	backend, err := local.NewBackend(cfg.PictureLocalPath, zaptest.NewLogger(t).Sugar())
	require.NoError(t, err)
	return assert, db, cfg, backend
}

func userRouter(t *testing.T, db *gorm.DB, cfg *config.Config, userID uint) *gin.Engine {
	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, userID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	return router
}

func photoHidden(t *testing.T, db *gorm.DB, userID uint) bool {
	var u models.User
	require.NoError(t, db.First(&u, userID).Error)
	return u.PhotoHidden != nil && *u.PhotoHidden
}

func TestFlagUserPhoto(t *testing.T) {
	assert, db, cfg, backend := photoFlagSetup(t)

	target := models.User{Type: models.UserTypeStudent, Name: "Target", UnixID: "target"}
	reporter := models.User{Type: models.UserTypeStudent, Name: "Reporter", UnixID: "reporter"}
	noPhoto := models.User{Type: models.UserTypeStudent, Name: "No Photo", UnixID: "nophoto"}
	hidden := models.User{Type: models.UserTypeStudent, Name: "Hidden", UnixID: "hidden", PhotoHidden: lib.BoolToPtr(true)}
	require.NoError(t, db.Create(&target).Error)
	require.NoError(t, db.Create(&reporter).Error)
	require.NoError(t, db.Create(&noPhoto).Error)
	require.NoError(t, db.Create(&hidden).Error)
	require.NoError(t, backend.SaveUserPhotoBoth(target.UnixID, utils.CreateTestImage(100, 100)))
	require.NoError(t, backend.SaveUserPhotoBoth(hidden.UnixID, utils.CreateTestImage(100, 100)))

	targetRouter := userRouter(t, db, cfg, target.ID)
	w, err := utils.DoHTTPReq(targetRouter, http.MethodPost, fmt.Sprintf("/%d/photo/flag", target.ID), nil)
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorUserCannotFlagOwnPhoto)

	reporterRouter := userRouter(t, db, cfg, reporter.ID)
	for _, u := range []models.User{noPhoto, hidden} {
		w, err = utils.DoHTTPReq(reporterRouter, http.MethodPost, fmt.Sprintf("/%d/photo/flag", u.ID), nil)
		require.NoError(t, err)
		utils.CheckRespError(assert, w, lib.ErrorUserPhotoNotFound)
	}

	w, err = utils.DoHTTPReq(reporterRouter, http.MethodPost, fmt.Sprintf("/%d/photo/flag", target.ID), nil)
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)

	w, err = utils.DoHTTPReq(reporterRouter, http.MethodPost, fmt.Sprintf("/%d/photo/flag", target.ID), nil)
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorUserPhotoFlagOpen)

	var tasks []*models.ModeratorTask
	require.NoError(t, db.Find(&tasks).Error)
	require.Len(t, tasks, 1)
	assert.Equal(models.ModeratorTaskTypeUserPhotoReview, tasks[0].TaskType)
	assert.Equal(models.TaskReference(models.ModeratorTaskTypeUserPhotoReview, target.ID), tasks[0].TaskReference)
}

func TestFlagUserPhoto_ApproveHidesPhoto(t *testing.T) {
	assert, db, cfg, backend := photoFlagSetup(t)

	target := models.User{Type: models.UserTypeStudent, Name: "Target", UnixID: "target"}
	reporter := models.User{Type: models.UserTypeStudent, Name: "Reporter", UnixID: "reporter"}
	require.NoError(t, db.Create(&target).Error)
	require.NoError(t, db.Create(&reporter).Error)
	require.NoError(t, backend.SaveUserPhotoBoth(target.UnixID, utils.CreateTestImage(100, 100)))

	reporterRouter := userRouter(t, db, cfg, reporter.ID)
	w, err := utils.DoHTTPReq(reporterRouter, http.MethodPost, fmt.Sprintf("/%d/photo/flag", target.ID), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)

	adminRouter := utils.SetupRouter(auth.ScopeAdminAll)
	utils.AddUserContexts(adminRouter, 99)
	admin.SetupRouter(adminRouter, db, cfg, zaptest.NewLogger(t).Sugar())

	var task models.ModeratorTask
	require.NoError(t, db.Where("task_type = ?", models.ModeratorTaskTypeUserPhotoReview).First(&task).Error)

	resolveBody, _ := json.Marshal(admin.ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err = utils.DoHTTPReq(adminRouter, http.MethodPost,
		fmt.Sprintf("/tasks/%d/resolve", task.ID), bytes.NewReader(resolveBody))
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)

	assert.True(photoHidden(t, db, target.ID))
	exists, err := backend.DoesUserPhotoExists(target.UnixID)
	require.NoError(t, err)
	assert.True(exists)
}

func TestUploadProfilePhoto_ClearsPhotoHidden(t *testing.T) {
	assert, db, cfg, _ := photoFlagSetup(t)

	user := models.User{Type: models.UserTypeStudent, Name: "Stu", UnixID: "stu", PhotoHidden: lib.BoolToPtr(true)}
	require.NoError(t, db.Create(&user).Error)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "photo.jpg")
	require.NoError(t, err)
	require.NoError(t, jpeg.Encode(part, utils.CreateTestImage(100, 100), nil))
	require.NoError(t, mw.Close())

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/%d/photo", user.ID), &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	userRouter(t, db, cfg, user.ID).ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)

	assert.False(photoHidden(t, db, user.ID))
}
