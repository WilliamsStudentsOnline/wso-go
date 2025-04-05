package parse

import (
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
)

func fetchDriscollVenue(d *api.WilliamsDiningAPI, sidebar map[string]int) (*search.Venue, error) {
	driscollDiningHall := search.NewDiningHall("Driscoll")

	dApiUnit, err := d.SelectUnitFromSideBar(sidebar[driscollDiningHall.Name])
	if err != nil {
		return nil, err
	}

	childUnitsData, err := dApiUnit.GetChildUnitsPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallMultiMenu(childUnitsData, driscollDiningHall, d)
	if err != nil {
		return nil, err
	}

	return &search.Venue{
		Name:        driscollDiningHall.Name,
		DiningHalls: []*search.DiningHall{driscollDiningHall},
	}, nil
}
