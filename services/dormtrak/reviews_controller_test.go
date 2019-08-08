package dormtrak_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	testify "github.com/stretchr/testify/assert"
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
	SetupRouter(router, db)

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
		assert.Equal(orderedReviews[i].ID, orderedReviews[i].ID)
		if orderedReviews[i].Comment != nil {
			assert.Equal(*orderedReviews[i].Comment, *orderedReviews[i].Comment)
		} else {
			assert.Nil(orderedReviews[i].Comment)
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
		fmt.Sprintf("/reviews?offset=%s&limit=%d", reviews[4].CreatedAt.Format(time.RFC3339Nano), 2), nil)
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

func generateReviewTestComment() *string {
	randBytes := make([]byte, 100)
	for i := 0; i < 100; i++ {
		randBytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	str := string(randBytes)
	return &str
}
