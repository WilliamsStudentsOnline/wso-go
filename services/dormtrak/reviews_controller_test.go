package dormtrak_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListReviews(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}

	d1 := models.Dorm{
		Neighborhood: &models.Neighborhood{
			Name: "Currier",
		},
		Name: "East",
	}
	d2 := models.Dorm{
		Neighborhood: &models.Neighborhood{
			Name: "Dodd",
		},
		Name: "Dodd",
	}
	dr1 := models.DormRoom{
		Dorm:   &d2,
		Number: "2.CONST",
	}

	assert.NoError(db.Create(&u1).Create(&u2).Create(&d1).Create(&d2).Create(&dr1).Error)

	reviews := []*models.DormtrakReview{
		{
			User: &u1,
			DormRoom: &models.DormRoom{
				Dorm:   &d1,
				Number: "1.1",
			},
			Comment: generateReviewTestComment(),
		},
		{
			User: &u1,
			DormRoom: &models.DormRoom{
				Dorm:   &d2,
				Number: "2.2",
			},
			Comment: generateReviewTestComment(),
		},
		{
			User: &u1,
			DormRoom: &models.DormRoom{
				Dorm:   &d1,
				Number: "1.2",
			},
			Comment: generateReviewTestComment(),
		},
		{
			User: &u2,
			DormRoom: &models.DormRoom{
				Dorm:   &d1,
				Number: "1.3",
			},
			Comment: generateReviewTestComment(),
		},
		{
			User:     &u1,
			DormRoom: &dr1,
			Comment:  generateReviewTestComment(),
		},
		{
			User:     &u2,
			DormRoom: &dr1,
			Comment:  generateReviewTestComment(),
		},
		{
			User: &u1,
			DormRoom: &models.DormRoom{
				Dorm:   &d1,
				Number: "1.4",
			},
			Comment: nil,
		},
	}

	for i := range reviews {
		assert.NoError(db.Create(&reviews[i]).Error)
	}

	// We then reverse the list, as by default we order by creation
	orderedReviews := make([]*models.DormtrakReview, len(reviews))
	copy(orderedReviews, reviews)
	for i := len(orderedReviews)/2 - 1; i >= 0; i-- {
		opp := len(orderedReviews) - 1 - i
		orderedReviews[i], orderedReviews[opp] = orderedReviews[opp], orderedReviews[i]
	}

	// Setup routing. Be user 1
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, u1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Test 1: Get all reviews
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/reviews", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.DormtrakReview
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check that reviews are correct
	assert.Len(resp, 7)
	for i := range resp {
		assert.Equal(orderedReviews[i].ID, resp[i].ID)
		if orderedReviews[i].Comment != nil {
			assert.Equal(*orderedReviews[i].Comment, *resp[i].Comment)
		} else {
			assert.Nil(resp[i].Comment)
		}

		// Ensure we hide userID if not user 1 (self)
		if orderedReviews[i].UserID != u1.ID {
			assert.Zero(resp[i].UserID)
			assert.Nil(resp[i].User)
		}
	}

	// Test 2: Get all reviews with dormID
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews?dormID=%d", d1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.DormtrakReview{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check that reviews are correct
	assert.Len(resp, 4)
	assert.Equal(reviews[6].ID, resp[0].ID)
	assert.Equal(reviews[3].ID, resp[1].ID)
	assert.Equal(reviews[2].ID, resp[2].ID)
	assert.Equal(reviews[0].ID, resp[3].ID)

	// Test 3: Get all reviews with dormRoomID
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews?dormRoomID=%d", dr1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.DormtrakReview{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check that reviews are correct
	assert.Len(resp, 2)
	assert.Equal(reviews[5].ID, resp[0].ID)
	assert.Equal(reviews[4].ID, resp[1].ID)

	// Test 4: Get all reviews with userID
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews?userID=%d", u1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.DormtrakReview{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check that reviews are correct
	assert.Len(resp, 5)
	assert.Equal(reviews[6].ID, resp[0].ID)
	assert.Equal(reviews[4].ID, resp[1].ID)
	assert.Equal(reviews[2].ID, resp[2].ID)
	assert.Equal(reviews[1].ID, resp[3].ID)
	assert.Equal(reviews[0].ID, resp[4].ID)

	// Test 5: Get all reviews with userID that is not self (fail)
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews?userID=%d", u2.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Test 6: Get all reviews that have a comment
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews?commented=true"), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.DormtrakReview{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check that reviews are correct
	assert.Len(resp, 6)
	for i := range resp {
		assert.Equal(orderedReviews[i].ID, orderedReviews[i].ID)
	}

	// Test 7: Get all reviews with userID AND dormID
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews?userID=%d&dormID=%d", u1.ID, d1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.DormtrakReview{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check that reviews are correct
	assert.Len(resp, 3)
	assert.Equal(reviews[6].ID, resp[0].ID)
	assert.Equal(reviews[2].ID, resp[1].ID)
	assert.Equal(reviews[0].ID, resp[2].ID)

	// Test 7: Get all reviews with pagination
	w, err = utils.DoHTTPReq(router, http.MethodGet,
		fmt.Sprintf("/reviews?start=%s&limit=%d", reviews[4].CreatedAt.Format(time.RFC3339Nano), 2), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.DormtrakReview{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check that reviews are correct. Expect the 4th and 5th review, as we gave it the 3rd review (by time)
	assert.Len(resp, 2)
	assert.Equal(reviews[3].ID, resp[0].ID)
	assert.Equal(reviews[2].ID, resp[1].ID)
}

func TestController_GetReview(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}

	dr1 := models.DormRoom{
		Dorm: &models.Dorm{
			Neighborhood: &models.Neighborhood{
				Name: "Currier",
			},
			Name: "East",
		},
		Number: "DormRoom.1",
	}

	assert.NoError(db.Create(&u1).Create(&u2).Create(&dr1).Error)

	reviews := []*models.DormtrakReview{
		{
			User:     &u1,
			DormRoom: &dr1,
			Comment:  generateReviewTestComment(),
		},
		{
			User:     &u2,
			DormRoom: &dr1,
			Comment:  generateReviewTestComment(),
		},
	}

	for i := range reviews {
		assert.NoError(db.Create(&reviews[i]).Error)
	}

	// Setup routing. Be user 1
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, u1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Test 1: Get review
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews/%d", reviews[0].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.DormtrakReview
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	assert.Equal(reviews[0].ID, resp.ID)
	assert.Equal(*reviews[0].Comment, *resp.Comment)
	// Preloads dorm room
	assert.Equal(reviews[0].DormRoom.ID, resp.DormRoom.ID)
	// Preloads dorm
	assert.Equal(reviews[0].DormRoom.Dorm.ID, resp.DormRoom.Dorm.ID)
	// Contains user info, as self
	assert.Equal(reviews[0].UserID, resp.UserID)

	// Test 2: Get review created by other user
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews/%d", reviews[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.DormtrakReview{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	assert.Equal(reviews[1].ID, resp.ID)
	assert.Equal(*reviews[1].Comment, *resp.Comment)
	// Preloads dorm room
	assert.Equal(reviews[1].DormRoom.ID, resp.DormRoom.ID)
	// Preloads dorm
	assert.Equal(reviews[1].DormRoom.Dorm.ID, resp.DormRoom.Dorm.ID)
	// Missing user info, as self
	assert.Zero(resp.UserID)
	assert.Nil(resp.User)

	// Test 3: Fail on missing review
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/reviews/%d", 42), nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_CreateReview(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	dr1 := models.DormRoom{
		Dorm: &models.Dorm{
			Neighborhood: &models.Neighborhood{
				Name: "Currier",
			},
			Name: "East",
		},
		Number: "103",
	}
	dr2 := models.DormRoom{
		Dorm: &models.Dorm{
			Neighborhood: &models.Neighborhood{
				Name: "Dodd",
			},
			Name: "Hubbel",
		},
		Number: "1A",
	}
	u1 := models.User{
		Type:     models.UserTypeStudent,
		Name:     "User 1",
		UnixID:   "u1",
		DormRoom: &dr1,
	}
	u2 := models.User{
		Type:     models.UserTypeStudent,
		Name:     "User 2",
		UnixID:   "u2",
		DormRoom: &dr2,
	}
	dtr1 := models.DormtrakReview{
		User:     &u2,
		DormRoom: &dr2,
		Comment:  generateReviewTestComment(),
	}
	assert.NoError(db.Create(&dr1).Create(&dr2).Create(&u1).Create(&u2).Create(&dtr1).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeDormtrakWrite)
	utils.AddUserContexts(router, u1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// First, we run tests on validations

	// Test 1: error on missing dorm room ID
	params := ReviewCreateParams{}
	createReviewExpectError(assert, router, params, lib.ErrorRequestDataValidationFailed)

	// Test 2: error on unowned dorm room
	params = ReviewCreateParams{DormRoomID: u2.DormRoomID}
	createReviewExpectError(assert, router, params, lib.ErrorReviewDormNotOwner)

	// Test 3: error on existing review
	params = ReviewCreateParams{DormRoomID: u2.DormRoomID, Comment: generateReviewTestComment()}
	// Setup bad student router
	r1 := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeDormtrakWrite)
	utils.AddUserContexts(r1, u2.ID)
	SetupRouter(r1, db, cfg, zaptest.NewLogger(t).Sugar())
	createReviewExpectError(assert, r1, params, lib.ErrorReviewAlreadyExists)

	// Test 4: create review
	params = ReviewCreateParams{DormRoomID: u1.DormRoomID, Comment: generateReviewTestComment(),
		BedAdjustable:    lib.BoolToPtr(false),
		Closet:           lib.StrToPtr("foobar"),
		ThermostatAccess: lib.BoolToPtr(true),
		Location:         lib.IntToPtr(4),
	}

	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Post review
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/reviews", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	var resp models.DormtrakReview
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert these values:
	assert.Equal(*params.Comment, *resp.Comment)
	assert.Equal(*params.DormRoomID, resp.DormRoomID)
	assert.Equal(u1.ID, resp.UserID)
	assert.Equal(*params.BedAdjustable, *resp.BedAdjustable)
	assert.Equal(*params.Closet, *resp.Closet)
	assert.Equal(*params.ThermostatAccess, *resp.ThermostatAccess)
	assert.Equal(*params.Location, *resp.Location)
	// Assert that we preload dorm room, dorm, and neighborhood
	assert.Equal(dr1.ID, resp.DormRoom.ID)
	assert.Equal(dr1.Dorm.ID, resp.DormRoom.Dorm.ID)
	assert.Equal(dr1.Dorm.Neighborhood.ID, resp.DormRoom.Dorm.Neighborhood.ID)
	// Assert that dorm room and dorm were updated with new review data
	assert.Equal(*params.Closet, *resp.DormRoom.Closet)
	assert.Equal(*params.ThermostatAccess, *resp.DormRoom.ThermostatAccess)
	assert.Equal(float64(*params.Location), *resp.DormRoom.Dorm.Location)
}

func TestController_UpdateReview(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}
	review := models.DormtrakReview{
		User: &u1,
		DormRoom: &models.DormRoom{
			Number: "103",
			Dorm: &models.Dorm{
				Neighborhood: &models.Neighborhood{
					Name: "Currier",
				},
				Name: "East",
			},
		},
		Comment:          generateReviewTestComment(),
		BedAdjustable:    lib.BoolToPtr(false),
		Closet:           lib.StrToPtr("foobar"),
		ThermostatAccess: lib.BoolToPtr(true),
		Location:         lib.IntToPtr(4),
		Wifi:             lib.IntToPtr(7),
	}
	assert.NoError(db.Create(&u1).Create(&u2).Create(&review).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeDormtrakWrite)
	utils.AddUserContexts(router, u1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// First, we run tests on validations

	// Test 1: error on out of range values (upper bound)
	params := ReviewUpdateParams{Location: lib.IntToPtr(100)}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	// Get bad review (expect failure)
	apiErr := lib.ErrorRequestDataValidationFailed
	w, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/reviews/%d", review.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 2: error on out of range values (lower bound)
	params = ReviewUpdateParams{Location: lib.IntToPtr(-1)}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	// Get bad review (expect failure)
	apiErr = lib.ErrorRequestDataValidationFailed
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/reviews/%d", review.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 3: error on missing review
	params = ReviewUpdateParams{Location: lib.IntToPtr(3)}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	// Get bad review (expect failure)
	apiErr = lib.ErrorRecordNotFound
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/reviews/%d", 42), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 4: error on bad user
	params = ReviewUpdateParams{Location: lib.IntToPtr(3)}
	r1 := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeDormtrakWrite)
	utils.AddUserContexts(r1, u2.ID)
	SetupRouter(r1, db, cfg, zaptest.NewLogger(t).Sugar())
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	// Get bad review (expect failure)
	apiErr = lib.ErrorMustBeSelf
	w, err = utils.DoHTTPReq(r1, http.MethodPatch, fmt.Sprintf("/reviews/%d", review.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 5: actually update and work
	params = ReviewUpdateParams{
		Comment:          generateReviewTestComment(),
		BedAdjustable:    lib.BoolToPtr(true),
		Closet:           lib.StrToPtr("baz"),
		ThermostatAccess: lib.BoolToPtr(false),
		Location:         lib.IntToPtr(6),
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)

	// Do HTTP
	w, err = utils.DoHTTPReq(router,
		http.MethodPatch, fmt.Sprintf("/reviews/%d", review.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	// Parse API response
	var resp models.DormtrakReview
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	// Assert for response
	assert.Equal(review.ID, resp.ID)
	assert.Equal(*params.Comment, *resp.Comment)
	assert.Equal(*params.BedAdjustable, *resp.BedAdjustable)
	assert.Equal(*params.Closet, *resp.Closet)
	assert.Equal(*params.ThermostatAccess, *resp.ThermostatAccess)
	assert.Equal(*params.Location, *resp.Location)
	assert.Equal(*review.Wifi, *resp.Wifi)
	// Preload/ensure room and dorm updated
	assert.Equal(*params.Closet, *resp.DormRoom.Closet)
	assert.Equal(*params.ThermostatAccess, *resp.DormRoom.ThermostatAccess)
	assert.Equal(float64(*params.Location), *resp.DormRoom.Dorm.Location)
	assert.Equal(float64(*review.Wifi), *resp.DormRoom.Dorm.Wifi)
	assert.NotNil(resp.DormRoom.Dorm.Neighborhood)
}

func TestController_DeleteReview(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}
	review := models.DormtrakReview{
		User: &u1,
		DormRoom: &models.DormRoom{
			Number: "103",
			Dorm: &models.Dorm{
				Neighborhood: &models.Neighborhood{
					Name: "Currier",
				},
				Name: "East",
			},
		},
		Comment:          generateReviewTestComment(),
		BedAdjustable:    lib.BoolToPtr(false),
		Closet:           lib.StrToPtr("foobar"),
		ThermostatAccess: lib.BoolToPtr(true),
		Location:         lib.IntToPtr(4),
		Wifi:             lib.IntToPtr(7),
	}
	assert.NoError(db.Create(&u1).Create(&u2).Create(&review).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeDormtrakWrite)
	utils.AddUserContexts(router, u1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// First, we run tests on validations

	// Test 1: error on missing review
	apiErr := lib.ErrorRecordNotFound
	w, err := utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/reviews/%d", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 2: error on bad user
	r1 := utils.SetupRouter(auth.ScopeDormtrak, auth.ScopeDormtrakWrite)
	utils.AddUserContexts(r1, u2.ID)
	SetupRouter(r1, db, cfg, zaptest.NewLogger(t).Sugar())
	apiErr = lib.ErrorMustBeSelf
	w, err = utils.DoHTTPReq(r1, http.MethodDelete, fmt.Sprintf("/reviews/%d", review.ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 3: actually update and work
	w, err = utils.DoHTTPReq(router,
		http.MethodDelete, fmt.Sprintf("/reviews/%d", review.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	// Parse API response
	var resp models.DormtrakReview
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	// Assert for response
	assert.Equal(review.ID, resp.ID)

	// Check to make sure it is not in db
	var count int
	assert.NoError(db.Model(models.NewDormtrakReview(review.ID)).Count(&count).Error)
	assert.Zero(count)
}

func generateReviewTestComment() *string {
	randBytes := make([]byte, 100)
	for i := 0; i < 100; i++ {
		randBytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	str := string(randBytes)
	return &str
}

func createReviewExpectError(assert *testify.Assertions, router *gin.Engine, params ReviewCreateParams, apiErr *lib.APIError) {
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Get bad review (expect failure)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/reviews", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	respErrCode := utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode
	assert.Equal(apiErr.Code, respErrCode)
}
