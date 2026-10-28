package test_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testMySQLOnce   sync.Once
	testMySQLErr    error
	testMySQLDBName string
	testTableNames  []string
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func testMySQLHost() string { return envOr("WSO_TEST_MYSQL_HOST", "127.0.0.1") }
func testMySQLUser() string { return envOr("WSO_TEST_MYSQL_USER", "root") }
func testMySQLPass() string {
	return envOr("WSO_TEST_MYSQL_PASSWORD", "secret-mysql-password")
}

func testMySQLPort() int {
	p, err := strconv.Atoi(envOr("WSO_TEST_MYSQL_PORT", "3307"))
	if err != nil || p <= 0 {
		return 3307
	}
	return p
}

func mysqlDSN(database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&multiStatements=true",
		testMySQLUser(), testMySQLPass(), testMySQLHost(), testMySQLPort(), database)
}

func unreachableMsg(err error) error {
	return fmt.Errorf(
		"test MariaDB unreachable at %s:%d (run `make itest`; is Docker up?): %w",
		testMySQLHost(), testMySQLPort(), err,
	)
}

// Create package DB once, apply Atlas migrations, cache app table names
func ensureTestMySQL() error {
	testMySQLOnce.Do(func() {
		testMySQLDBName = fmt.Sprintf("wso_test_%d", os.Getpid())

		root, err := gorm.Open("mysql", mysqlDSN(""))
		if err != nil {
			testMySQLErr = unreachableMsg(err)
			return
		}
		defer root.Close()

		if err := root.Exec(fmt.Sprintf("CREATE DATABASE `%s`", testMySQLDBName)).Error; err != nil {
			testMySQLErr = fmt.Errorf("create test database %s: %w", testMySQLDBName, err)
			return
		}

		db, err := gorm.Open("mysql", mysqlDSN(testMySQLDBName))
		if err != nil {
			testMySQLErr = unreachableMsg(err)
			return
		}
		defer db.Close()

		if err := migrate.MigrateDB(db); err != nil {
			testMySQLErr = fmt.Errorf("atlas migrate test database: %w", err)
			return
		}

		// Atlas uses datetime (second precision); widen to datetime(6) in tests so
		// cursor pagination / created_at ordering match GORM's time.Time like SQLite did
		if err := upgradeDateTimePrecision(db); err != nil {
			testMySQLErr = err
			return
		}

		tables, err := listAppTables(db)
		if err != nil {
			testMySQLErr = err
			return
		}
		testTableNames = tables
	})
	return testMySQLErr
}

func upgradeDateTimePrecision(db *gorm.DB) error {
	rows, err := db.Raw(`
		SELECT table_name, column_name
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		  AND data_type = 'datetime'
	`).Rows()
	if err != nil {
		return fmt.Errorf("list datetime columns: %w", err)
	}
	defer rows.Close()

	type col struct{ table, name string }
	var cols []col
	for rows.Next() {
		var c col
		if err := rows.Scan(&c.table, &c.name); err != nil {
			return err
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, c := range cols {
		q := fmt.Sprintf("ALTER TABLE `%s` MODIFY `%s` datetime(6)", c.table, c.name)
		if err := db.Exec(q).Error; err != nil {
			return fmt.Errorf("%s: %w", q, err)
		}
	}
	return nil
}

func listAppTables(db *gorm.DB) ([]string, error) {
	rows, err := db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_type = 'BASE TABLE'
		  AND table_name != 'schema_migrations'
		ORDER BY table_name
	`).Rows()
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func truncateAppTables(db *gorm.DB) error {
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return err
	}
	for _, name := range testTableNames {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", name)).Error; err != nil {
			_ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
			return fmt.Errorf("truncate %s: %w", name, err)
		}
	}
	return db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error
}

// Set up database for testing
func SetupServiceTest(assert *assert.Assertions) *gorm.DB {
	gin.SetMode(gin.TestMode)
	quietLogsInCI()

	err := ensureTestMySQL()
	assert.NoError(err)

	cfg := SetupConfig()
	db := config.LoadDatabase(cfg, zap.NewNop().Sugar())
	if os.Getenv("CI") == "" {
		db.SetLogger(gorm.Logger{LogWriter: log.New(os.Stdout, "\r\n", 0)})
	} else {
		db.LogMode(false)
	}

	assert.NoError(truncateAppTables(db))
	return db
}

// SetupServiceDB is SetupServiceTest without testify (examples / one-offs)
func SetupServiceDB() (*gorm.DB, error) {
	gin.SetMode(gin.TestMode)
	quietLogsInCI()

	if err := ensureTestMySQL(); err != nil {
		return nil, err
	}

	cfg := SetupConfig()
	db := config.LoadDatabase(cfg, zap.NewNop().Sugar())
	if os.Getenv("CI") == "" {
		db.SetLogger(gorm.Logger{LogWriter: log.New(os.Stdout, "\r\n", 0)})
	} else {
		db.LogMode(false)
	}
	if err := truncateAppTables(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func quietLogsInCI() {
	if os.Getenv("CI") == "" {
		return
	}
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
}

// Return test configuration values
func SetupConfig() *config.Config {
	dbName := testMySQLDBName
	if dbName == "" {
		dbName = fmt.Sprintf("wso_test_%d", os.Getpid())
	}

	return &config.Config{
		Env:           "test",
		GinMode:       "test",
		JWTRealm:      "wso-go-test",
		DatabaseType:  "mysql",
		DatabaseArgs:  mysqlDSN(dbName),
		MySQLUser:     testMySQLUser(),
		MySQLHost:     testMySQLHost(),
		MySQLPort:     testMySQLPort(),
		MySQLDatabase: dbName,
		MySQLArgs: map[string]string{
			"parseTime":       "true",
			"charset":         "utf8mb4",
			"multiStatements": "true",
		},
		Secrets: &config.Secrets{
			JWTSecretKey:  "wso-jwt-test-secret",
			MySQLPassword: testMySQLPass(),
		},
		SearchBackend: search.SearchBackendSQL,
	}
}

// Create a router with the given scopes
func SetupRouter(scopes ...string) *gin.Engine {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("scopes", scopes)
		c.Next()
	})

	return router
}

// Values needed for tests
type TestEnv struct {
	Assert *assert.Assertions
	DB     *gorm.DB
	Router *gin.Engine
	Cfg    *config.Config
}

// Setup for testing, creating environment variables that work with the given scopes
func SetupTest(t *testing.T, scopes ...string) *TestEnv {
	assert := assert.New(t)
	db := SetupServiceTest(assert)
	router := SetupRouter(scopes...)
	cfg := SetupConfig()
	return &TestEnv{
		Assert: assert,
		DB:     db,
		Router: router,
		Cfg:    cfg,
	}
}

// Add a user context to a router
func AddUserContexts(router *gin.Engine, userID uint) {
	router.Use(func(c *gin.Context) {
		c.Set("id", userID)
		c.Next()
	})
}

// Perform the correct HTTP request on the given router, with the correct body and parameters
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

// Data from an HTTP response
type APITestResp struct {
	Status          int                 `json:"status"`
	Data            json.RawMessage     `json:"data,omitempty"`
	Error           *services.RespError `json:"error,omitempty"`
	UpdateToken     bool                `json:"updateToken,omitempty"`
	PaginationTotal int                 `json:"paginationTotal,omitempty"`
}

// Retrieve the body of an HTTP response
func GetHTTPDataResp(assert *assert.Assertions, body []byte) APITestResp {
	resp := APITestResp{}
	err := json.Unmarshal(body, &resp)
	assert.NoError(err)

	return resp
}

// Assert that an HTTP response is successful and return the response's data
func GetGoodResp(assert *assert.Assertions, w *httptest.ResponseRecorder) APITestResp {
	// Response status is ok
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	return respData
}

// Assert that a response's error codes match the expected API Error's codes
func CheckRespError(assert *assert.Assertions, w *httptest.ResponseRecorder, apiErr *lib.APIError) {
	// Response status matches expected error
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)
}
