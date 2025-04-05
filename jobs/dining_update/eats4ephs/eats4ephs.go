package eats4ephs

import (
	"encoding/json"
	"net/http"
)

const MenuAPI = "https://dining.williams.edu/wp-json/dining/menus"

type MenuItem struct {
	ServiceUnit  string `json:"service_unit"`
	UnitID       string `json:"unitid"`
	Course       string `json:"course"`
	FormalName   string `json:"formal_name"`
	Meal         string `json:"meal"`
	PortionSize  string `json:"portion_size"`
	NetNutrition string `json:"net_nutrition"`
}

func GetDailyMenu() ([]MenuItem, error) {
	resp, err := http.Get(MenuAPI)
	if err != nil {
		return nil, err
	}

	var menu []MenuItem
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&menu)
	return menu, err
}
