package dining_update

const (
	foodTruckFullBelly = "full-belly-fewd"
	foodTruckBerkshire = "berkshire-catering-co"
	foodTruckOpenTime  = "11:00am"
	foodTruckCloseTime = "2:00pm"
)

// foodTruckNames is the display name for each truck vendor id.
var foodTruckNames = map[string]string{
	foodTruckFullBelly: "Full Belly Fewd",
	foodTruckBerkshire: "Berkshire Catering Co",
}

// foodTruckSchedule maps YYYY-MM-DD (year-inclusive) to the truck vendor ids
// present that day. Days with no trucks are omitted (e.g. Spring Break).
var foodTruckSchedule = map[string][]string{
	// Fall 2026
	"2026-09-08": {foodTruckFullBelly},
	"2026-09-10": {foodTruckBerkshire},
	"2026-09-15": {foodTruckFullBelly},
	"2026-09-17": {foodTruckFullBelly, foodTruckBerkshire},
	"2026-09-22": {foodTruckFullBelly},
	"2026-09-24": {foodTruckBerkshire},
	"2026-09-29": {foodTruckFullBelly},
	"2026-10-01": {foodTruckBerkshire},
	"2026-10-06": {foodTruckFullBelly},
	"2026-10-08": {foodTruckBerkshire},
	"2026-10-13": {foodTruckFullBelly},
	"2026-10-15": {foodTruckBerkshire},
	"2026-10-20": {foodTruckFullBelly},
	"2026-10-22": {foodTruckBerkshire},
	"2026-10-27": {foodTruckFullBelly},
	"2026-10-29": {foodTruckBerkshire},
	"2026-11-03": {foodTruckFullBelly},
	"2026-11-05": {foodTruckBerkshire},
	"2026-11-10": {foodTruckFullBelly},
	"2026-11-12": {foodTruckBerkshire},
	"2026-11-17": {foodTruckFullBelly},
	"2026-11-19": {foodTruckBerkshire},

	// Spring 2027
	"2027-02-16": {foodTruckFullBelly},
	"2027-02-18": {foodTruckBerkshire},
	"2027-02-23": {foodTruckFullBelly},
	"2027-02-25": {foodTruckBerkshire},
	"2027-03-02": {foodTruckFullBelly},
	"2027-03-04": {foodTruckBerkshire},
	"2027-03-09": {foodTruckFullBelly},
	"2027-03-11": {foodTruckBerkshire},
	"2027-03-16": {foodTruckFullBelly},
	"2027-03-18": {foodTruckBerkshire},
	// 2027-03-23 .. 2027-04-01: Spring Break — no trucks
	"2027-04-06": {foodTruckFullBelly},
	"2027-04-08": {foodTruckBerkshire},
	"2027-04-13": {foodTruckFullBelly},
	"2027-04-15": {foodTruckBerkshire},
	"2027-04-20": {foodTruckFullBelly},
	"2027-04-22": {foodTruckBerkshire},
	"2027-04-27": {foodTruckFullBelly},
	"2027-04-29": {foodTruckBerkshire},
	"2027-05-04": {foodTruckFullBelly},
	"2027-05-06": {foodTruckBerkshire},
	"2027-05-11": {foodTruckFullBelly},
	"2027-05-13": {foodTruckBerkshire},
	"2027-05-18": {foodTruckFullBelly},
}

// applyFoodTrucks injects scheduled food-truck vendors into existing date
// entries. Dates with no schedule entry are left unchanged; no new date keys
// are created, so output stays tied to the normal dining week.
func applyFoodTrucks(vendorsByDate map[string]map[string]Vendor) {
	for dateStr, vendors := range vendorsByDate {
		truckIDs, ok := foodTruckSchedule[dateStr]
		if !ok {
			continue
		}
		for _, id := range truckIDs {
			vendors[id] = makeFoodTruckVendor(id)
		}
	}
}

func makeFoodTruckVendor(id string) Vendor {
	name := foodTruckNames[id]
	if name == "" {
		name = id
	}
	return Vendor{
		Name:        name,
		OnlineOrder: false,
		Operating:   true,
		Meals: map[string]*Meal{
			"lunch": {
				Name: "lunch",
				Hours: &Hours{
					Open:  foodTruckOpenTime,
					Close: foodTruckCloseTime,
				},
			},
		},
	}
}
