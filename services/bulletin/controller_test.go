package bulletin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"

	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
)

var (
	bulletin1 = models.Bulletin{
		BaseSchema: models.BaseSchema{
			ID: 1,
		},
		Title: "Test",
		Body:  "Hello WSO!",
		Type:  "lostAndFound",
	}
	bulletin2 = models.Bulletin{
		BaseSchema: models.BaseSchema{
			ID: 2,
		},
		Title: "Test",
		Body:  "Hello WSO!",
		Type:  "job",
	}
	bulletin3 = models.Bulletin{
		BaseSchema: models.BaseSchema{
			ID: 3,
		},
		Title: "Test",
		Body:  "Hello WSO!",
		Type:  "lostAndFound",
	}
)

func TestController_GetBulletinByID(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test bulletin into db
	err := db.FirstOrCreate(&bulletin1).Error
	assert.NoError(err)

	// Get test bulletin
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin := models.Bulletin{}
	err = json.Unmarshal(respData, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal(bulletin1.ID, respBulletin.ID)
	assert.Equal(bulletin1.Title, respBulletin.Title)
}

func TestController_FetchAllBulletins(t *testing.T) {

	t.Run("Fetches all Bulletins", func(t *testing.T) {
		// Setup (can copy and paste this basically)
		assert := testify.New(t)
		db := utils.SetupServiceTest(assert)
		router := gin.Default()
		SetupRouter(router, db)

		// Insert test bulletin into db
		err := db.FirstOrCreate(&bulletin1).Error
		assert.NoError(err)
		err = db.FirstOrCreate(&bulletin2).Error
		assert.NoError(err)
		err = db.FirstOrCreate(&bulletin3).Error
		assert.NoError(err)

		// Get test bulletin
		w, err := utils.DoHTTPReq(router, http.MethodGet, "/", nil)
		assert.NoError(err)

		// Status is okay
		assert.Equal(http.StatusOK, w.Code)

		// Decode response
		respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
		var respBulletin = []models.Bulletin{}
		err = json.Unmarshal(respData, &respBulletin)
		assert.NoError(err)

		// Check if all bulletins are obtained
		assert.Equal(len(respBulletin), 3)

		// Check if all the bulletins titles and bodies were obtained correctly
		assert.Equal(bulletin1.ID, respBulletin[0].ID)
		assert.Equal(bulletin1.Title, respBulletin[0].Title)
		assert.Equal(bulletin1.Body, respBulletin[0].Body)

		assert.Equal(bulletin2.ID, respBulletin[1].ID)
		assert.Equal(bulletin2.Title, respBulletin[1].Title)
		assert.Equal(bulletin2.Body, respBulletin[1].Body)

		assert.Equal(bulletin3.ID, respBulletin[2].ID)
		assert.Equal(bulletin3.Title, respBulletin[2].Title)
		assert.Equal(bulletin3.Body, respBulletin[2].Body)
	})

	t.Run("Fetches only the Lost and Found Bulletins",
		func(t *testing.T) {
			// Setup (can copy and paste this basically)
			assert := testify.New(t)
			db := utils.SetupServiceTest(assert)
			router := gin.Default()
			SetupRouter(router, db)

			// Insert test bulletin into db
			err := db.FirstOrCreate(&bulletin1).Error
			assert.NoError(err)
			err = db.FirstOrCreate(&bulletin2).Error
			assert.NoError(err)
			err = db.FirstOrCreate(&bulletin3).Error
			assert.NoError(err)

			// Get test bulletin
			w, err := utils.DoHTTPReq(router, http.MethodGet, "/?type=lostAndFound", nil)
			assert.NoError(err)

			// Status is okay
			assert.Equal(http.StatusOK, w.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
			var respBulletin = []models.Bulletin{}
			err = json.Unmarshal(respData, &respBulletin)
			assert.NoError(err)

			t.Log(respBulletin)

			// Check if all lost and found bulletins (1 and 3) are obtained
			assert.Equal(2, len(respBulletin))

			// Check if all the bulletins titles and bodies were obtained correctly
			assert.Equal(bulletin1.ID, respBulletin[0].ID)
			assert.Equal(bulletin1.Title, respBulletin[0].Title)
			assert.Equal(bulletin1.Body, respBulletin[0].Body)

			assert.Equal(bulletin3.ID, respBulletin[1].ID)
			assert.Equal(bulletin3.Title, respBulletin[1].Title)
			assert.Equal(bulletin3.Body, respBulletin[1].Body)
		})
}

func TestController_DeleteBulletin(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test bulletin into db
	err := db.FirstOrCreate(&bulletin1).Error
	assert.NoError(err)

	// Get test bulletin
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin := models.Bulletin{}
	err = json.Unmarshal(respData, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal(bulletin1.ID, respBulletin.ID)
	assert.Equal(bulletin1.Title, respBulletin.Title)

	// Delete test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodDelete, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Get test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status not found because Bulletin is deleted
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_UpdateBulletin(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test bulletin into db
	err := db.FirstOrCreate(&bulletin1).Error
	assert.NoError(err)

	// Get test bulletin
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin := models.Bulletin{}
	err = json.Unmarshal(respData, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal(bulletin1.ID, respBulletin.ID)
	assert.Equal(bulletin1.Title, respBulletin.Title)

	// Construct update params
	var jsonStr = []byte(`{"title":"ByeWSO."}`)

	// Update test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodPut, "/1", bytes.NewBuffer(jsonStr))
	assert.NoError(err)
	t.Log(w.HeaderMap)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Get test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin = models.Bulletin{}
	err = json.Unmarshal(respData, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal("ByeWSO.", respBulletin.Title)
}

// TODO: Tests: 1) Unable to update forbidden fields
