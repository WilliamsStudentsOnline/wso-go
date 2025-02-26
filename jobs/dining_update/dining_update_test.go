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
    
	t.Logf("Starting TestDriscoll")

	t.Log("Creating Williams Dining API...")
	d, err := api.CreateWilliamsDiningAPI()
	if err != nil {
        	t.Fatalf("Failed to create API: %v", err)
    	}
	assert.NoError(err)
	
    	if d == nil {
		t.Fatal("API client is nil despite no error")
    	}
	
	t.Log("Calling parse.Populate...")
	venues, err := parse.Populate(d)
	
	fmt.Println(err)
	assert.NoError(err)

	fmt.Println("venues", venues)
}
