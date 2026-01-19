package dining_update

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type VendorInfo struct {
	Name            string `json:"name" yaml:"name"`
	NutriSliceSlug  string `json:"NutriSliceSlug" yaml:"nutrislice_slug"`
	Operating       bool   `json:"operating" yaml:"operating"`
	OnlineOrder     bool   `json:"onlineOrder" yaml:"online_order"`
	// key is day of week, value is list of meals to hours
	Hours map[string]VendorInfoHoursWeek `json:"hours" yaml:"hours"`
}

// key is meal, value is hours
type VendorInfoHoursWeek map[string]VendorInfoHoursMeals

type VendorInfoHoursMeals struct {
	Open        string `json:"open" yaml:"open"`
	Close       string `json:"close" yaml:"close"`
	DisplayName string `json:"display_name,omitempty" yaml:"display_name,omitempty"`
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
