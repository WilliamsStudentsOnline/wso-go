package api

import "fmt"

type SideBarUnit struct {
	Success bool `json:"success"`
	Panels  []struct {
		ID   string `json:"id"`
		Html string `json:"html"`
	} `json:"panels"`
}

type DiningAPIUnit struct {
	Success         bool
	ItemPanel       string
	StaticPanel3    string
	UnitsPanel      string
	ChildUnitsPanel string
	MenuPanel       string
	CoursesPanel    string
	DisclaimerPanel string
}

func (d *DiningAPIUnit) GetItemPanel() (string, error) {
	if d.ItemPanel == "" {
		return "", fmt.Errorf("item panel empty or does not exist")
	}

	return d.ItemPanel, nil
}

func (d *DiningAPIUnit) GetMenuPanel() (string, error) {
	if d.MenuPanel == "" {
		return "", fmt.Errorf("menu panel empty or does not exist")
	}

	return d.MenuPanel, nil
}

func (d *DiningAPIUnit) GetChildUnitsPanel() (string, error) {
	if d.ChildUnitsPanel == "" {
		return "", fmt.Errorf("child units panel empty or does not exist")
	}

	return d.ChildUnitsPanel, nil
}

func (s *SideBarUnit) ToDiningAPIUnit() *DiningAPIUnit {
	d := DiningAPIUnit{
		Success: s.Success,
	}

	for _, panel := range s.Panels {
		switch panel.ID {
		case "itemPanel":
			d.ItemPanel = panel.Html
			break
		case "staticPanel3":
			d.StaticPanel3 = panel.Html
			break
		case "unitsPanel":
			d.UnitsPanel = panel.Html
			break
		case "childUnitsPanel":
			d.ChildUnitsPanel = panel.Html
			break
		case "menuPanel":
			d.MenuPanel = panel.Html
			break
		case "coursesPanel":
			d.CoursesPanel = panel.Html
			break
		case "disclaimerPanel":
			d.DisclaimerPanel = panel.Html
			break
		}
	}

	return &d
}
