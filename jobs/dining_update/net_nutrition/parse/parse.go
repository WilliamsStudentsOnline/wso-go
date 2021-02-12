package parse

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
)

func parseDiningHallMultiMenu(menuOptionsData string, dh *search.DiningHall, d *api.WilliamsDiningAPI) error {
	menuOptions, err := parseChildUnitsToNameId(menuOptionsData)
	if err != nil {
		return err
	}

	for name, id := range menuOptions {
		dApiUnit, err := d.SelectUnitFromChildUnitsList(id)
		if err != nil {
			return err
		}

		if dApiUnit.ItemPanel != "" {
			cm, err := parseConstantMenu(dApiUnit.ItemPanel, name, d)
			if err != nil {
				return err
			}

			dh.Menus[name] = cm
		} else if dApiUnit.MenuPanel != "" {
			dm, err := parseDailyMenu(dApiUnit.MenuPanel, name, d)
			if err != nil {
				return err
			}
			dh.Menus[name] = dm
		}
	}

	return nil
}

func parseDiningHallDirectDailyMenu(dailyMenuData string, dh *search.DiningHall, d *api.WilliamsDiningAPI) error {
	dm, err := parseDailyMenu(dailyMenuData, dh.Name, d)
	if err != nil {
		return err
	}

	dh.Menus[dh.Name] = dm

	return nil
}

func parseDiningHallDirectConstantMenu(constMenuData string, dh *search.DiningHall, d *api.WilliamsDiningAPI) error {
	dm, err := parseConstantMenu(constMenuData, dh.Name, d)
	if err != nil {
		return err
	}

	dh.Menus[dh.Name] = dm

	return nil
}

func parseDailyMenu(dailyMenuData string, menuName string, d *api.WilliamsDiningAPI) (*search.DailyMenu, error) {
	dailyMenuStructure, err := parseDailyMenuData(dailyMenuData)
	if err != nil {
		return nil, err
	}

	dm := search.DailyMenu{
		Name: menuName,
		Days: make(map[string]search.DayMeals),
	}

	for date, meals := range dailyMenuStructure {
		dayMeals := make(map[string]*search.Menu)

		for meal, mealId := range meals {
			dApiUnit, err := d.SelectMenu(mealId)
			if err != nil {
				return nil, err
			}

			itemPanel, err := dApiUnit.GetItemPanel()
			if err != nil {
				return nil, err
			}

			mealItems, err := parseMenuItems(itemPanel)
			if err != nil {
				return nil, err
			}

			meal = strings.ToLower(meal)

			dayMeals[meal] = mealItems
		}

		dm.Days[date] = dayMeals
	}

	return &dm, nil
}

func parseConstantMenu(constMenuData string, menuName string, d *api.WilliamsDiningAPI) (*search.ConstantMenu, error) {
	menuItems, err := parseMenuItems(constMenuData)
	if err != nil {
		return nil, err
	}

	return &search.ConstantMenu{
		Name: menuName,
		Menu: menuItems,
	}, nil
}
