package pic_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures/local"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services/auth"
	"github.com/WilliamsStudentsOnline/wso-go/services/auth/api"
	"github.com/WilliamsStudentsOnline/wso-go/services/pic"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type picEnv struct {
	router *gin.Engine
	token  string
	dir    string
}

func setupPic(t *testing.T, users ...*models.User) picEnv {
	t.Helper()
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()
	cfg.JWTAPITimeoutHours = 1
	cfg.PictureBackend = pictures.PictureBackendLocal
	cfg.PictureLocalPath = t.TempDir()
	log := zaptest.NewLogger(t).Sugar()

	for _, u := range users {
		require.NoError(t, db.Create(u).Error)
	}
	viewer := models.User{Type: models.UserTypeStudent, Name: "Viewer", UnixID: "viewer"}
	require.NoError(t, db.Create(&viewer).Error)

	mw, err := api.LoadMiddleware(cfg, db, log)
	require.NoError(t, err)
	token, _, err := mw.TokenGenerator(&auth.AuthenticatorPayload{
		TokenLevel: auth.TokenLevelUser,
		User:       &viewer,
	})
	require.NoError(t, err)

	router := gin.New()
	group := router.Group("")
	group.Use(mw.MiddlewareFunc())
	pic.SetupRouter(group.Group("/pic"), db, cfg, log)

	return picEnv{router: router, token: token, dir: cfg.PictureLocalPath}
}

func (e picEnv) get(path string, withToken bool) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if withToken {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}
	e.router.ServeHTTP(w, req)
	return w
}

func savePhoto(t *testing.T, dir, unixID string) {
	t.Helper()
	backend, err := local.NewBackend(dir, zaptest.NewLogger(t).Sugar())
	require.NoError(t, err)
	require.NoError(t, backend.SaveUserPhotoBoth(unixID, utils.CreateTestImage(100, 100)))
}

func TestGetPicture_ServesUserPhoto(t *testing.T) {
	assert := testify.New(t)
	env := setupPic(t)
	savePhoto(t, env.dir, "stu")

	for _, path := range []string{"/pic/large/stu.jpg", "/pic/thumb/stu.jpg", "/pic/user/large/stu.jpg", "/pic/user/thumb/stu.jpg"} {
		w := env.get(path, true)
		assert.Equal(http.StatusOK, w.Code, path)
		assert.Equal("image/jpeg", w.Header().Get("Content-Type"), path)
		assert.Equal("private, max-age=300", w.Header().Get("Cache-Control"), path)
		assert.NotEmpty(w.Header().Get("Last-Modified"), path)
	}
}

func TestGetPicture_RequiresAuth(t *testing.T) {
	assert := testify.New(t)
	env := setupPic(t)
	savePhoto(t, env.dir, "stu")

	assert.Equal(http.StatusUnauthorized, env.get("/pic/user/large/stu.jpg", false).Code)
}

func TestGetPicture_HiddenPhoto(t *testing.T) {
	assert := testify.New(t)
	hidden := &models.User{Type: models.UserTypeStudent, Name: "Hidden", UnixID: "hidden", PhotoHidden: lib.BoolToPtr(true)}
	env := setupPic(t, hidden)
	savePhoto(t, env.dir, "hidden")

	for _, path := range []string{"/pic/large/hidden.jpg", "/pic/user/thumb/hidden.jpg"} {
		assert.Equal(http.StatusNotFound, env.get(path, true).Code, path)
	}
}

func TestGetPicture_MissingAndTraversal(t *testing.T) {
	assert := testify.New(t)
	env := setupPic(t)

	secret := filepath.Join(filepath.Dir(env.dir), "secret.txt")
	require.NoError(t, os.WriteFile(secret, []byte("secret"), 0600))
	t.Cleanup(func() { os.Remove(secret) })

	assert.Equal(http.StatusNotFound, env.get("/pic/user/large/nobody.jpg", true).Code)
	assert.Equal(http.StatusNotFound, env.get("/pic/user", true).Code)
	for _, path := range []string{"/pic/../secret.txt", "/pic/user/../../secret.txt", "/pic/%2e%2e/secret.txt"} {
		w := env.get(path, true)
		assert.NotEqual(http.StatusOK, w.Code, path)
		assert.NotContains(w.Body.String(), "secret", path)
	}
}
