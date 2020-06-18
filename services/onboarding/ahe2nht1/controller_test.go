package ahe2nht1

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestController_GetUserByUnix(t *testing.T) {

	require := assert.New(t)
	db := utils.SetupServiceTest(require)

	u1 := models.User{
		Name:      "Test 1",
		UnixID:    "unix1",
		ClassYear: lib.IntToPtr(2022),
	}
	require.NoError(db.Create(&u1).Error) // puts user in DB!
	// init routing
	router := utils.SetupRouter(auth.ScopeUsers) // creat test router
	// make it seem like requests are coming from user1
	utils.AddUserContexts(router, u1.ID)
	// creates testing config
	cfg := utils.SetupConfig()
	// creates testing logger
	logger := zap.S()
	// set up router with testing parameters
	SetupRouter(router, db, cfg, logger)

	// ACTUALLY TEST
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/unix1", nil) //GET request for unix
	require.NoError(err)                                             // ensure no failure

	//Decode response
	resp := utils.GetHTTPDataResp(require, w.Body.Bytes()) //decodes into rest API format
	require.Nil(resp.Error)                                // ensure no error from resp

	// store decoded response in respUser
	respUser := models.User{}                  // initialize an empty user
	err = json.Unmarshal(resp.Data, &respUser) // put the data into user struct API
	require.NoError(err)                       // ensure no errors from dumping into user struct

	// An example test that ensures that the unixID from the user
	// we inserted into the DB and the unixID we got from the API are the same.
	require.Equal(u1.UnixID, respUser.UnixID) // check if they are the same
	// TEST NAME : Ensure that the name inserted into db and name retrieved are the same
	require.Equal(u1.Name, respUser.Name)
	// TEST YEAR : Ensure that the class year of the user inserted into db and the year received are the same
	require.Equal(u1.ClassYear, respUser.ClassYear)

}
