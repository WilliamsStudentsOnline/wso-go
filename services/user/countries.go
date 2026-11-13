package user

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed countries.json states.json
var countryData embed.FS

var validCountries = loadNames("countries.json")
var validStates = loadNames("states.json")

func loadNames(filename string) []string {
	data, err := countryData.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var countries []string
	if err := json.Unmarshal(data, &countries); err != nil {
		panic(err)
	}
	return countries
}

var stateAbbreviations = map[string]string{
	"AL": "Alabama", "AK": "Alaska", "AZ": "Arizona", "AR": "Arkansas",
	"CA": "California", "CO": "Colorado", "CT": "Connecticut", "DE": "Delaware",
	"DC": "District of Columbia", "FL": "Florida", "GA": "Georgia", "HI": "Hawaii",
	"ID": "Idaho", "IL": "Illinois", "IN": "Indiana", "IA": "Iowa", "KS": "Kansas",
	"KY": "Kentucky", "LA": "Louisiana", "ME": "Maine", "MD": "Maryland",
	"MA": "Massachusetts", "MI": "Michigan", "MN": "Minnesota", "MS": "Mississippi",
	"MO": "Missouri", "MT": "Montana", "NE": "Nebraska", "NV": "Nevada",
	"NH": "New Hampshire", "NJ": "New Jersey", "NM": "New Mexico", "NY": "New York",
	"NC": "North Carolina", "ND": "North Dakota", "OH": "Ohio", "OK": "Oklahoma",
	"OR": "Oregon", "PA": "Pennsylvania", "RI": "Rhode Island", "SC": "South Carolina",
	"SD": "South Dakota", "TN": "Tennessee", "TX": "Texas", "UT": "Utah",
	"VT": "Vermont", "VA": "Virginia", "WA": "Washington", "WV": "West Virginia",
	"WI": "Wisconsin", "WY": "Wyoming",
}

func canonicalUSState(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if state, ok := stateAbbreviations[strings.ToUpper(trimmed)]; ok {
		return state, true
	}
	for _, state := range validStates {
		if strings.EqualFold(trimmed, state) {
			return state, true
		}
	}
	return "", false
}

// canonicalCountry returns the display name for a country, ignoring case and
// surrounding whitespace. The short US names are retained for existing clients.
func canonicalCountry(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if strings.EqualFold(trimmed, "US") || strings.EqualFold(trimmed, "USA") || strings.EqualFold(trimmed, "United States of America") {
		return "United States", true
	}
	for _, country := range validCountries {
		if strings.EqualFold(trimmed, country) {
			return country, true
		}
	}
	return "", false
}
