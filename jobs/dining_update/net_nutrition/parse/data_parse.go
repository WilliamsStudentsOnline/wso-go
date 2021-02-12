package parse

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/api"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dining_update/net_nutrition/search"
	"strconv"
	"strings"
)

func parseChildUnitsToNameId(data string) (map[string]int, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	childUnits := make(map[string]int)

	doc.Find(".cbo_nn_childUnitsCell a").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())

		onclick, exists := s.Attr("onclick")

		if !exists {
			return
		}
		onclickSubMatch := api.RegexpGetId.FindStringSubmatch(onclick)
		if len(onclickSubMatch) < 2 {
			return
		}

		id, err := strconv.Atoi(onclickSubMatch[1])
		if err != nil {
			return
		}

		childUnits[name] = id
	})

	return childUnits, nil
}

func parseMenuItems(data string) (*search.Menu, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	menuItems := make(search.Menu)

	currentMenuHeader := "Unknown"
	menuItems[currentMenuHeader] = []string{}

	doc.Find(".cbo_nn_itemGridTable tbody tr").Has("td").Each(func(i int, s *goquery.Selection) {
		// Case if subheader
		if s.Find(".cbo_nn_itemGroupRow").Length() > 0 {
			currentMenuHeader = strings.TrimSpace(s.Find(".cbo_nn_itemGroupRow").Text())
			menuItems[currentMenuHeader] = []string{}
			return
		}

		// Case if item
		menuItems[currentMenuHeader] = append(menuItems[currentMenuHeader], strings.TrimSpace(s.Find("td.cbo_nn_itemHover").Text()))
	})

	if len(menuItems["Unknown"]) == 0 {
		delete(menuItems, "Unknown")
	}

	return &menuItems, nil
}

func parseDailyMenuData(data string) (map[string]map[string]int, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	dailyMenu := make(map[string]map[string]int)

	doc.Find("table.cbo_nn_menuTable td.cbo_nn_menuCell > table > tbody").Each(func(_ int, s *goquery.Selection) {
		dateHeader := s.ChildrenFiltered("tr").First().Text()
		dailyMenu[dateHeader] = make(map[string]int)

		s.Find("td.cbo_nn_menuLinkCell a").Each(func(_ int, childSel *goquery.Selection) {
			name := strings.TrimSpace(childSel.Text())

			onclick, exists := childSel.Attr("onclick")

			if !exists {
				return
			}
			onclickSubMatch := api.RegexpGetId.FindStringSubmatch(onclick)
			if len(onclickSubMatch) < 2 {
				return
			}

			id, err := strconv.Atoi(onclickSubMatch[1])
			if err != nil {
				return
			}

			dailyMenu[dateHeader][name] = id
		})
	})

	return dailyMenu, nil
}
