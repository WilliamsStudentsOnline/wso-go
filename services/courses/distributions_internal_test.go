package courses

import (
	"testing"

	testify "github.com/stretchr/testify/assert"
)

func TestTally(t *testing.T) {
	d := tally([]entryCounts{
		{Subject: "MATH", Div3: true, QFR: true},
		{Subject: "MATH", Div3: true},
		{Subject: "MATH", Div3: true}, // a third course in one subject doesn't count for the division
		{Subject: "ENGL", Div1: true, WAC: true},
	})
	testify.Equal(t, PlanDistributions{Div1: 1, Div3: 2, WS: 1, QFR: 1}, d)
}
