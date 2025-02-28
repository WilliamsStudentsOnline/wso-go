package dining_update

import (
	"fmt"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/parse"
	testify "github.com/stretchr/testify/assert"
)

func TestDriscoll(t *testing.T) {
	assert := testify.New(t)

	d, err := api.CreateWilliamsDiningAPI()
	if err != nil {
		// CI/CD test marked as bot request and blocked by Williams
		if err.Error() == "403" {
			t.Skip("Skipping test due to Cloudflare firewall")
		} else {
			assert.NoError(err)
		}
	}

	venues, err := parse.Populate(d)
	fmt.Println(err)
	assert.NoError(err)

	fmt.Println("venues", venues)
}
