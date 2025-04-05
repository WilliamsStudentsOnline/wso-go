package parse

import (
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
)

func fetchMissionVenue(d *api.WilliamsDiningAPI, sidebar map[string]int) (*search.Venue, error) {
	missionDiningHall := search.NewDiningHall("Mission")

	dApiUnit, err := d.SelectUnitFromSideBar(sidebar[missionDiningHall.Name])
	if err != nil {
		return nil, err
	}

	childUnitsData, err := dApiUnit.GetChildUnitsPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallMultiMenu(childUnitsData, missionDiningHall, d)
	if err != nil {
		return nil, err
	}

	return &search.Venue{
		Name:        missionDiningHall.Name,
		DiningHalls: []*search.DiningHall{missionDiningHall},
	}, nil
}
