package parse

import (
	"fmt"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
)

func fetchPareskyVenue(d *api.WilliamsDiningAPI, sidebar map[string]int) (*search.Venue, error) {
	paresky := search.Venue{
		Name: "Paresky Student Center",
	}

	dApiUnit, err := d.SelectUnitFromSideBar(sidebar[paresky.Name])
	if err != nil {
		return nil, err
	}

	childUnitsData, err := dApiUnit.GetChildUnitsPanel()
	if err != nil {
		return nil, err
	}

	childUnits, err := parseChildUnitsToNameId(childUnitsData)
	if err != nil {
		return nil, err
	}

	whitmans, err := fetchPareskyWhitmansDining(d, childUnits)
	if err != nil {
		return nil, err
	}

	/*whitmansLateNight, err := fetchPareskyWhitmansLateNightDining(d, childUnits)
	if err != nil {
		return nil, err
	}*/

	freshAndGo, err := fetchPareskyFreshAndGoDining(d, childUnits)
	if err != nil {
		return nil, err
	}

	/*leeSnackBar, err := fetchPareskyLeeSnackBar(d, childUnits)
	if err != nil {
		return nil, err
	}*/

	/*grill82, err := fetchParesky82Grill(d, childUnits)
	if err != nil {
		return nil, err
	}*/

	paresky.DiningHalls = []*search.DiningHall{
		whitmans,
		freshAndGo,
	}
	return &paresky, nil
}

func loadPareskyDiningHall(name string, nameToId map[string]int, d *api.WilliamsDiningAPI) (*search.DiningHall, *api.DiningAPIUnit, error) {
	dh := search.NewDiningHall(name)

	id, ok := nameToId[name]
	if !ok {
		return nil, nil, fmt.Errorf("missing name in map name=%s", name)
	}

	dApiUnit, err := d.SelectUnitFromChildUnitsList(id)
	if err != nil {
		return nil, nil, err
	}

	return dh, dApiUnit, nil
}

func fetchPareskyWhitmansDining(d *api.WilliamsDiningAPI, nameToId map[string]int) (*search.DiningHall, error) {
	whitmansDiningHall, dApiUnit, err := loadPareskyDiningHall("Whitmans'", nameToId, d)
	if err != nil {
		return nil, err
	}

	childUnitsData, err := dApiUnit.GetChildUnitsPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallMultiMenu(childUnitsData, whitmansDiningHall, d)
	if err != nil {
		return nil, err
	}

	return whitmansDiningHall, nil
}

func fetchPareskyWhitmansLateNightDining(d *api.WilliamsDiningAPI, nameToId map[string]int) (*search.DiningHall, error) {
	whitmansLateNightDiningHall, dApiUnit, err := loadPareskyDiningHall("Whitmans' Late Night", nameToId, d)
	if err != nil {
		return nil, err
	}

	childUnitsData, err := dApiUnit.GetChildUnitsPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallMultiMenu(childUnitsData, whitmansLateNightDiningHall, d)
	if err != nil {
		return nil, err
	}

	return whitmansLateNightDiningHall, nil
}

func fetchPareskyFreshAndGoDining(d *api.WilliamsDiningAPI, nameToId map[string]int) (*search.DiningHall, error) {
	freshAndGoDiningHall, dApiUnit, err := loadPareskyDiningHall("Fresh & Go", nameToId, d)
	if err != nil {
		return nil, err
	}

	menuData, err := dApiUnit.GetItemPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallDirectConstantMenu(menuData, freshAndGoDiningHall, d)
	if err != nil {
		return nil, err
	}

	return freshAndGoDiningHall, nil
}

func fetchPareskyLeeSnackBar(d *api.WilliamsDiningAPI, nameToId map[string]int) (*search.DiningHall, error) {
	leeSnackBarDiningHall, dApiUnit, err := loadPareskyDiningHall("Lee Snack Bar", nameToId, d)
	if err != nil {
		return nil, err
	}

	menuData, err := dApiUnit.GetMenuPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallDirectConstantMenu(menuData, leeSnackBarDiningHall, d)
	if err != nil {
		return nil, err
	}

	return leeSnackBarDiningHall, nil
}

func fetchParesky82Grill(d *api.WilliamsDiningAPI, nameToId map[string]int) (*search.DiningHall, error) {
	grill82DiningHall, dApiUnit, err := loadPareskyDiningHall("'82 Grill", nameToId, d)
	if err != nil {
		return nil, err
	}

	itemData, err := dApiUnit.GetItemPanel()
	if err != nil {
		return nil, err
	}

	err = parseDiningHallDirectConstantMenu(itemData, grill82DiningHall, d)
	if err != nil {
		return nil, err
	}

	return grill82DiningHall, nil
}
