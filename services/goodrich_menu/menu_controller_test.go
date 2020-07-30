package goodrich_menu

import (
	"encoding/json"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"net/http"
	"testing"
)

func TestController_ListMenuItems(t *testing.T) {

	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichAdmin, auth.ScopeGoodrichOrder)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	//db.Table("goodrich_menu_items").CreateTable(&models.MenuItem{})

	m1 := models.MenuItem{
		BaseSchema: models.BaseSchema{
			ID: 1,
		},
		Title:   "Item1",
		Description:  "Test item 1",
		Price: 3.00,
		Available: true,
	}

	// m2 will have no description
	m2 := models.MenuItem{
		BaseSchema: models.BaseSchema{
			ID: 2,
		},
		Title:   "Item2",
		Price: 4.00,
		Available: false,
	}

	assert.NoError(db.Create(&m1).Create(&m2).Error)

	// Get test menu items
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/goodrich/menu"+"?includeAvailable=false", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.MenuItem
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if menu items are correct (ordered by date)
	assert.Len(resp, 2)
	assert.Equal(m1.Title, resp[0].Title)
	assert.Equal(m2.Title, resp[1].Title)

	//TODO: fix, getting 404 error

}