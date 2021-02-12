package parse

import (
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
)

func fetchEcoCafeVenue(d *api.WilliamsDiningAPI, sidebar map[string]int) (*search.Venue, error) {
	ecoCafeDiningHall := search.NewDiningHall("Eco Cafe")

	dApiUnit, err := d.SelectUnitFromSideBar(sidebar[ecoCafeDiningHall.Name])
	if err != nil {
		return nil, err
	}

	childUnitsData, err := dApiUnit.GetMenuPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallDirectDailyMenu(childUnitsData, ecoCafeDiningHall, d)
	if err != nil {
		return nil, err
	}

	return &search.Venue{
		Name:        ecoCafeDiningHall.Name,
		DiningHalls: []*search.DiningHall{ecoCafeDiningHall},
	}, nil
}
