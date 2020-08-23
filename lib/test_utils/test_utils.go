package test_utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func SetupServiceTest(assert *assert.Assertions) *gorm.DB {
	gin.SetMode(gin.TestMode)
	cfg := SetupConfig()

	db := config.LoadDatabase(cfg, zap.NewNop().Sugar())
	db.SetLogger(gorm.Logger{LogWriter: log.New(os.Stdout, "\r\n", 0)})
	err := migrate.MigrateDB(db)
	assert.NoError(err)

	return db
}

func SetupConfig() *config.Config {
	return &config.Config{
		Env:          "test",
		GinMode:      "test",
		JWTRealm:     "wso-go-test",
		DatabaseType: "sqlite3",
		DatabaseArgs: ":memory:",
		Secrets: &config.Secrets{
			JWTSecretKey: "wso-jwt-test-secret",
		},
		SearchBackend: search.SearchBackendSQL,
	}
}

func SetupRouter(scopes ...string) *gin.Engine {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("scopes", scopes)
		c.Next()
	})

	return router
}

type TestEnv struct {
	Assert			*assert.Assertions
	DB					*gorm.DB
	Router			*gin.Engine
	Cfg					*config.Config
}

// Setup for testing, creating environment variables that work with the given scopes
func SetupTest(t *testing.T, scopes ...string) *TestEnv {
	assert := assert.New(t)
	db := SetupServiceTest(assert)
	router := SetupRouter(scopes...)
	cfg := SetupConfig()
	return &TestEnv{
		Assert:	assert,
		DB:			db,
		Router:	router,
		Cfg:		cfg,
	}
}

func AddUserContexts(router *gin.Engine, userID uint) {
	router.Use(func(c *gin.Context) {
		c.Set("id", userID)
		c.Next()
	})
}

func DoHTTPReq(router *gin.Engine, method, url string, body io.Reader) (*httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// If we have a body, set the Content Type of the request to JSON
	if body != nil {
		req.Header.Set("Content-Type", gin.MIMEJSON)
	}

	router.ServeHTTP(w, req)
	return w, nil
}

type APITestResp struct {
	Status          int                 `json:"status"`
	Data            json.RawMessage     `json:"data,omitempty"`
	Error           *services.RespError `json:"error,omitempty"`
	UpdateToken     bool                `json:"updateToken,omitempty"`
	PaginationTotal int                 `json:"paginationTotal,omitempty"`
}

func GetHTTPDataResp(assert *assert.Assertions, body []byte) APITestResp {
	resp := APITestResp{}
	err := json.Unmarshal(body, &resp)
	assert.NoError(err)

	return resp
}
