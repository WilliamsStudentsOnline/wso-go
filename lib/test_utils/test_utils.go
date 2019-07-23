package test_utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func SetupServiceTest(assert *assert.Assertions) *gorm.DB {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Env:          "test",
		GinMode:      "test",
		JWTRealm:     "wso-go-test",
		DatabaseType: "sqlite3",
		DatabaseArgs: ":memory:",
		Secrets: &config.Secrets{
			JWTSecretKey: "wso-jwt-test-secret",
		},
	}

	db := config.LoadDatabase(cfg)
	err := migrate.MigrateDB(db)
	assert.NoError(err)

	return db
}

func DoHTTPReq(router *gin.Engine, method, url string, body io.Reader) (*httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	router.ServeHTTP(w, req)
	return w, nil
}

func GetHTTPDataResp(assert *assert.Assertions, body []byte) []byte {
	resp := map[string]json.RawMessage{}
	err := json.Unmarshal(body, &resp)
	assert.NoError(err)

	return []byte(resp["data"])
}
