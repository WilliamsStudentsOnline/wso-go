package dining_update

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type VendorInfo struct {
	Name            string `json:"name" yaml:"name"`
	Eats4EphsUnitID string `json:"eats4EphsUnitID" yaml:"eats_4_ephs_unit_id"`
	Operating       bool   `json:"operating" yaml:"operating"`
	OnlineOrder     bool   `json:"onlineOrder" yaml:"online_order"`
	// key is day of week, value is list of meals to hours
	Hours map[string]VendorInfoHoursWeek `json:"hours" yaml:"hours"`
}

// key is meal, value is hours
type VendorInfoHoursWeek map[string]VendorInfoHoursMeals

type VendorInfoHoursMeals struct {
	Open  string `json:"open" yaml:"open"`
	Close string `json:"close" yaml:"close"`
}

func ReadVendorInfo(path string) (map[string]VendorInfo, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	out := make(map[string]VendorInfo)
	err = yaml.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
