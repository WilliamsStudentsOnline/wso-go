package api

import (
	"fmt"
	"net/http"
	"time"
)

const WilliamsDiningBaseUrl = "https://williamsdining.api.nutrislice.com/menu/api/"

type NutriSliceAPI struct {
	baseUrl string
	client  *http.Client
}

func CreateNutriSliceAPI() *NutriSliceAPI {
	return &NutriSliceAPI{
		baseUrl: WilliamsDiningBaseUrl,
		client:  NewHTTPClient(15 * time.Second),
	}
}

func (a *NutriSliceAPI) WeeklyMenuURL(diningHallSlug, meal string, d time.Time) string {
	return fmt.Sprintf("%sweeks/school/%s/menu-type/%s/%d/%02d/%02d/",
		a.baseUrl, diningHallSlug, meal, d.Year(), int(d.Month()), d.Day())
}

// returns raw bytes of the nutrislice JSON for a given dining hall, meal, and date
func (a *NutriSliceAPI) GetWeeklyMenu(diningHallSlug, meal string, d time.Time) ([]byte, error) {
	url := a.WeeklyMenuURL(diningHallSlug, meal, d)
	return DoGet(a.client, url)
}
