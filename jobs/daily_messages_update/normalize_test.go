package daily_messages_update

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadFixture(t *testing.T, name string) Export {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	now := time.Date(2026, 11, 20, 7, 15, 0, 0, eastern)
	out, skipped, err := Normalize(raw, "2026-11-20", now)
	require.Empty(t, skipped)
	require.NoError(t, err)
	return out
}

func findMessage(t *testing.T, out Export, id int) Message {
	t.Helper()
	for _, c := range out.Categories {
		for _, m := range c.Messages {
			if m.ID == id {
				return m
			}
		}
	}
	t.Fatalf("message %d not found", id)
	return Message{}
}

func TestNormalizeEmpty(t *testing.T) {
	out := loadFixture(t, "empty.json")
	assert.NotNil(t, out.Categories)
	assert.Len(t, out.Categories, 0)
	assert.Equal(t, "2026-11-20", out.Date)
	assert.Equal(t, "2026-11-20T07:15:00-05:00", out.UpdateTime)

	b, err := json.Marshal(out)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"categories":[]`)
}

func TestNormalizeFeed(t *testing.T) {
	out := loadFixture(t, "feed.json")

	var names []string
	for _, c := range out.Categories {
		names = append(names, c.Name)
	}
	assert.Equal(t, []string{
		"General Announcements",
		"Lectures, Films, Readings, Panels",
		"Campus Life/Student Organizations",
		"On and Around Campus",
		"Academic/Teaching/Research",
		"Arts at Williams",
		"Environment and Sustainability",
		"Information Technology",
	}, names)

	vogue := findMessage(t, out, 204142)
	assert.Equal(t, "62_center", vogue.Type)
	assert.Equal(t, "Benji Hart Vogue Class and Workshop!", vogue.Title)
	assert.Equal(t, "https://events.williams.edu/files/2026/11/Benji-Flyer.png", vogue.ImageURL)
	assert.Equal(t, "https://forms.gle/91rw7SqjAtjoyf1i8", vogue.MoreURL)
	assert.True(t, strings.HasSuffix(vogue.Summary, "Join …"), vogue.Summary)
	assert.Equal(t, []string{"2026-11-20", "2026-11-30"}, vogue.RunDates)
	assert.Equal(t, []string{"faculty", "staff", "students"}, vogue.Audiences)
	require.Len(t, vogue.Departments, 5)
	assert.Equal(t, Department{Slug: "62-center", Name: "'62 Center"}, vogue.Departments[0])
	require.NotNil(t, vogue.Event)
	assert.Equal(t, "’62 Center for Theatre & Dance", vogue.Event.Venue)
	assert.Equal(t, "Dance Studio", vogue.Event.Room)
	assert.Equal(t, "2026-12-02T19:00:00-05:00", vogue.Event.Start)
	assert.Equal(t, "2026-12-02T21:30:00-05:00", vogue.Event.End)
	assert.False(t, vogue.Event.AllDay)
	assert.Equal(t, "7:00 pm - 9:30 pm", vogue.Event.TimeFormatted)

	notice := findMessage(t, out, 204398)
	assert.Equal(t, "announcement", notice.Type)
	assert.Nil(t, notice.Event)
	assert.Equal(t, "", notice.ImageURL)
	assert.Equal(t, "", notice.TicketURL)
	assert.Equal(t, []string{"2026-11-20", "2026-11-23"}, notice.RunDates)
	assert.Equal(t, "https://events.williams.edu/event/daily-messages-over-the-thanksgiving-holiday-12/", notice.URL)
	assert.Equal(t, Author{Name: "Noelle Lemoine", Email: "nlemoine@williams.edu", Department: "Communications"}, notice.Author)
	assert.True(t, strings.HasPrefix(notice.Body, "Daily Messages will not be approved"))
	assert.NotContains(t, notice.Body, "<p>")
	assert.Contains(t, notice.BodyHTML, "<p>")
	assert.NotContains(t, notice.Summary, "<p>")

	single := findMessage(t, out, 204382)
	assert.Equal(t, []string{"2026-11-20"}, single.RunDates)

	for _, c := range out.Categories {
		for _, m := range c.Messages {
			assert.NotNil(t, m.Audiences, m.ID)
			assert.NotNil(t, m.Departments, m.ID)
			assert.NotNil(t, m.RunDates, m.ID)
			assert.NotContains(t, m.URL, "utm_", m.ID)
			assert.NotContains(t, m.Title, "&#", m.ID)
		}
	}
}

func TestNormalizeTolerantFields(t *testing.T) {
	raw := []byte(`{
		"A": [
			{"ID": 1, "type": "event", "title": false, "content": null, "dm_text": false, "permalink": false,
			 "author": false, "orgs": null, "depts_arr": false, "terms": [], "meta": [], "all_day": "1"},
			{"ID": "bad"},
			{"ID": 2, "type": "announcement", "title": "ok", "terms": {"daily_message_audience": false},
			 "meta": {"acf_daily_message_run_dates_0": [{"value": "garbage"}], "acf_daily_message_run_dates_1": [{"value": "20261120"}], "other": "x"}}
		],
		"B": {"0": {"ID": 3, "type": "announcement", "title": "zero"}, "2": {"ID": 4, "type": "announcement", "title": "two"}}
	}`)
	out, skipped, err := Normalize(raw, "2026-11-20", testNow)
	require.NoError(t, err)
	require.Len(t, skipped, 1)

	require.Len(t, out.Categories, 2)
	require.Len(t, out.Categories[0].Messages, 2)
	first := out.Categories[0].Messages[0]
	assert.Equal(t, 1, first.ID)
	assert.Equal(t, "", first.Title)
	assert.Equal(t, []Department{}, first.Departments)
	assert.Equal(t, []string{}, first.Audiences)
	require.NotNil(t, first.Event)
	assert.True(t, first.Event.AllDay)
	assert.Equal(t, "", first.Event.Start)

	assert.Equal(t, []string{"2026-11-20"}, out.Categories[0].Messages[1].RunDates)

	require.Len(t, out.Categories[1].Messages, 2)
	assert.Equal(t, 3, out.Categories[1].Messages[0].ID)
	assert.Equal(t, 4, out.Categories[1].Messages[1].ID)
}
