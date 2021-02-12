package parse

import (
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
)

func Populate(d *api.WilliamsDiningAPI) (map[string]*search.Venue, error) {
	sidebar, err := d.FetchSidebar()
	if err != nil {
		return nil, err
	}

	// Mission
	mission, err := fetchMissionVenue(d, sidebar)
	if err != nil {
		return nil, err
	}

	// Paresky
	paresky, err := fetchPareskyVenue(d, sidebar)
	if err != nil {
		return nil, err
	}

	// Driscoll
	driscoll, err := fetchDriscollVenue(d, sidebar)
	if err != nil {
		return nil, err
	}

	// Eco Cafe
	/*ecoCafe, err := fetchEcoCafeVenue(d, sidebar)
	if err != nil {
		return nil, err
	}*/

	return map[string]*search.Venue{
		mission.Name: mission, paresky.Name: paresky, driscoll.Name: driscoll}, nil
}
