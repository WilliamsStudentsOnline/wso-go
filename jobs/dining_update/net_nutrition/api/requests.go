package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

func (d *WilliamsDiningAPI) doUnitRequest(r *http.Request) (*DiningAPIUnit, error) {
	resp, err := d.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	sideBarUnit := &SideBarUnit{}
	if err = json.Unmarshal(bodyBytes, sideBarUnit); err != nil {
		return nil, err
	}

	if sideBarUnit.Success == false {
		return nil, fmt.Errorf("query response: success=false")
	}

	return sideBarUnit.ToDiningAPIUnit(), resp.Body.Close()
}

func (d *WilliamsDiningAPI) SelectMenu(menuId int) (*DiningAPIUnit, error) {
	req, err := http.NewRequest(http.MethodPost, d.baseUrl + "/Menu/SelectMenu", bytes.NewBufferString(url.Values{
		"menuOid": []string{strconv.Itoa(menuId)},
	}.Encode()))
	if err != nil {
		return nil, err
	}

	return d.doUnitRequest(req)
}

func (d *WilliamsDiningAPI) SelectUnitFromSideBar(unitId int) (*DiningAPIUnit, error) {
	req, err := http.NewRequest(http.MethodPost, d.baseUrl + "/Unit/SelectUnitFromSideBar", bytes.NewBufferString(url.Values{
		"unitOid": []string{strconv.Itoa(unitId)},
	}.Encode()))
	if err != nil {
		return nil, err
	}

	return d.doUnitRequest(req)
}

func (d *WilliamsDiningAPI) SelectUnitFromChildUnitsList(unitId int) (*DiningAPIUnit, error) {
	req, err := http.NewRequest(http.MethodPost, d.baseUrl + "/Unit/SelectUnitFromChildUnitsList", bytes.NewBufferString(url.Values{
		"unitOid": []string{strconv.Itoa(unitId)},
	}.Encode()))
	if err != nil {
		return nil, err
	}

	return d.doUnitRequest(req)
}

func (d *WilliamsDiningAPI) FetchSidebar() (map[string]int, error) {
	req, err := http.NewRequest(http.MethodGet, d.baseUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	sidebar := make(map[string]int)

	doc.Find("#sideUnitPanel .cbo_nn_sideUnitCell a").Each(func(i int, s *goquery.Selection) {
		name := s.Text()

		onclick, exists := s.Attr("onclick")

		if !exists {
			return
		}
		onclickSubMatch := RegexpGetId.FindStringSubmatch(onclick)
		if len(onclickSubMatch) < 2 {
			return
		}

		id, err := strconv.Atoi(onclickSubMatch[1])
		if err != nil {
			return
		}

		sidebar[name] = id
	})

	return sidebar, nil
}
