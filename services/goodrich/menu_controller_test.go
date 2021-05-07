package goodrich

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListMenu(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrich)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	m1 := models.GoodrichMenuItem{
		Title:         "Bagel",
		Description:   "It's a bagel",
		Price:         1.69,
		Available:     true,
		QuantityLimit: false,
	}
	m2 := models.GoodrichMenuItem{
		Title:         "Coffee",
		Description:   "you drink it",
		Price:         3.41,
		Available:     true,
		QuantityLimit: false,
	}
	m3 := models.GoodrichMenuItem{
		Title:         "Avocado",
		Description:   "a rare specialty",
		Price:         4.29,
		Available:     false,
		QuantityLimit: false,
	}
	assert.NoError(db.Create(&m1).Create(&m2).Create(&m3).Error)

	/* Test 1: works normally */
	// Get test menu
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/menu", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.GoodrichMenuItem
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct menu items
	assert.Len(resp, 2)
	assert.Equal(m1.Title, resp[0].Title)
	assert.Equal(m2.Title, resp[1].Title)

	/* Test 2: ?all=true */
	// Get test menu
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/menu?all=true", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.GoodrichMenuItem{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct menu items
	assert.Len(resp, 3)
	assert.Equal(m3.Title, resp[0].Title) // (A)vocado
	assert.Equal(m1.Title, resp[1].Title) // (B)agel
	assert.Equal(m2.Title, resp[2].Title) // (C)offee
}
