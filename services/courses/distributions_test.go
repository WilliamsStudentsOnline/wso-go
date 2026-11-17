package courses_test

import (
	"net/http"
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	"github.com/stretchr/testify/require"
)

func TestMyDistributions(t *testing.T) {
	assert, env, _ := seedMine(t)
	// Both terms count as earned; a someday course only with planned
	for _, req := range []AddMyCourseRequest{
		{CrseID: "100136", Strm: 1251, ClassNbrs: []int{1001}},
		{CrseID: "100136", Strm: 1271, ClassNbrs: []int{2001}},
	} {
		code, body := send(t, env, http.MethodPost, "/mine", req)
		require.Equal(t, http.StatusCreated, code, string(body))
	}
	code, body := send(t, env, http.MethodGet, "/mine/distributions", nil)
	require.Equal(t, http.StatusOK, code, string(body))
	var d DistributionsResponse
	data(t, assert, body, &d)
	assert.Equal(2, d.Earned.Div3)
	assert.Equal(2, d.Earned.QFR)
	assert.Equal(3, d.Needed.Div1)
	assert.Equal(2, d.Needed.WS)
}
