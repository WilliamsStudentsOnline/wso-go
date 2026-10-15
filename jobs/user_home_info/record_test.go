package user_home_info

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTSV(t *testing.T) {
	input := "" +
		"EMAIL_ADDR\tSUBOX\tDORM\tENTRY\tCITY\tSTATE\tCOUNTRY\tPOSTAL\r\n" +
		"jhc6@williams.edu\t2642 Paresky\tWilliams\tW1\tBranford\tCT\tUnited States\t06405-5401\r\n" +
		"av16@williams.edu\t1024 Paresky\tArmstrong\tAP2\tSingapore\t \tSingapore\t268851\r\n" +
		"mgd4@williams.edu\t2604 Paresky\tOff-Campus\t \tGoshen\tNY\tUnited States\t10924-1526\r\n" +
		"gk12@williams.edu\t \t \t \tSeoul\t \tKorea, Republic of\t06544\r\n"

	records, err := ParseTSV(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, records, 4)

	assert.Equal(t, Record{
		UnixID:      "jhc6",
		Email:       "jhc6@williams.edu",
		SUBox:       "2642 Paresky",
		Dorm:        "Williams",
		Entry:       "W1",
		HomeTown:    "Branford",
		HomeState:   "CT",
		HomeCountry: "United States",
		HomeZip:     "06405-5401",
	}, records[0])

	assert.Equal(t, "av16", records[1].UnixID)
	assert.Equal(t, "", records[1].HomeState)
	assert.Equal(t, "Singapore", records[1].HomeTown)
	assert.Equal(t, "Singapore", records[1].HomeCountry)

	assert.Equal(t, "Off-Campus", records[2].Dorm)
	assert.Equal(t, "", records[2].Entry)

	assert.Equal(t, "gk12", records[3].UnixID)
	assert.Equal(t, "", records[3].SUBox)
	assert.Equal(t, "", records[3].Dorm)
	assert.Equal(t, "Seoul", records[3].HomeTown)
	assert.Equal(t, "Korea, Republic of", records[3].HomeCountry)
}

func TestParseTSVRejectsBadHeader(t *testing.T) {
	input := "EMAIL\tSUBOX\tDORM\tENTRY\tCITY\tSTATE\tCOUNTRY\tPOSTAL\n"
	_, err := ParseTSV(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "header")
}

func TestParseTSVRejectsNonWilliamsEmail(t *testing.T) {
	input := "" +
		"EMAIL_ADDR\tSUBOX\tDORM\tENTRY\tCITY\tSTATE\tCOUNTRY\tPOSTAL\n" +
		"someone@example.com\t1 Paresky\tWest\t \tTown\tMA\tUnited States\t01267\n"
	_, err := ParseTSV(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "@williams.edu")
}

func TestNullableString(t *testing.T) {
	assert.Nil(t, NullableString(""))
	assert.Equal(t, "Boston", NullableString("Boston"))
}

func TestUnixFromEmail(t *testing.T) {
	unix, err := unixFromEmail("  ABC1@Williams.EDU ")
	require.NoError(t, err)
	assert.Equal(t, "abc1", unix)

	_, err = unixFromEmail("not-an-email")
	require.Error(t, err)
}
