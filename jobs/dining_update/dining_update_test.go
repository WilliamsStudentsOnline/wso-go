package dining_update

import "testing"

func TestNutriSliceMenuType(t *testing.T) {
	vi := VendorInfo{
		Hours: map[string]VendorInfoHoursWeek{
			"monday": {
				"lunch": {
					Open:               "11:00am",
					Close:              "3:00pm",
					NutriSliceMenuType: "82-grill-lunch",
				},
				"dinner": {
					Open:  "4:00pm",
					Close: "8:00pm",
				},
			},
		},
	}

	if got := nutriSliceMenuType(vi, "lunch"); got != "82-grill-lunch" {
		t.Fatalf("lunch slug = %q, want 82-grill-lunch", got)
	}
	if got := nutriSliceMenuType(vi, "dinner"); got != "dinner" {
		t.Fatalf("dinner slug = %q, want dinner (fallback to key)", got)
	}
}
