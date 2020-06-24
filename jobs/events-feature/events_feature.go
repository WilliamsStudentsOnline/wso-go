package events_update

import (
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"

	//"strconv"
	//"strings"
	"time"
	//"unicode"
)

const (
	// EVENTS_URL stores the endpoint for the daily announcements
	EVENTS_URL = "https://events.williams.edu/wp-json/wms/events/v1/list/dm"
)

type RawAnnouncements struct {
	HeaderName string
	announcements []RawGeneralAnnouncement
}

type Cats struct {
	TermID         string `json:"term_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	TermGroup      string `json:"term_group"`
	TermTaxonomyID string `json:"term_taxonomy_id"`
	Taxonomy       string `json:"taxonomy"`
	Description    string `json:"description"`
	Parent         string `json:"parent"`
	Count          string `json:"count"`
	Filter         string `json:"filter"`
	TermOrder      string `json:"term_order"`
}
type CSS struct {
	TableWrap   string `json:"tablewrap"`
	OL          string `json:"ol"`
	LITop       string `json:"li_top"`
	LIBottom    string `json:"li_bottom"`
	HR          string `json:"hr"`
	A           string `json:"a"`
	Strong      string `json:"strong"`
	IMG         string `json:"img"`
	Table       string `json:"table"`
	Hero        string `json:"hero"`
	Logo        string `json:"logo"`
	ViewOnline  string `json:"view_online"`
	JumpLink    string `json:"jump_link"`
	BTB         string `json:"btb"`
	Heading     string `json:"heading"`
	TOC         string `json:"toc"`
	Title       string `json:"title"`
	BodyText    string `json:"bodytext"`
	Link        string `json:"link"`
	Count       string `json:"count"`
	TextWIndent string `json:"text_w_indent"`
}

type RawGeneralAnnouncement struct {
	ID                  string `json:"post_author"`
	PostDate            string `json:"post_date"`
	PostDateGMT         string `json:"post_date_gmt"`
	PostContent         string `json:"post_content"`
	PostTitle           string `json:"post_title"`
	PostExcerpt         string `json:"post_excerpt"`
	PostStatus          string `json:"post_status"`
	CommentStatus       string `json:"comment_status"`
	PingStatus          string `json:"ping_status"`
	PostPassword        string `json:"post_password"`
	PostName            string `json:"post_name"`
	ToPing              string `json:"to_ping"`
	Pinged              string `json:"pinged"`
	PostModified        string `json:"post_modified"`
	PostModifiedGMT     string `json:"post_modified_gmt"`
	PostContentFiltered string `json:"post_content_filtered"`
	PostParent          string `json:"post_parent"`
	GUID                string `json:"guid"`
	MenuOrder           string `json:"menu_order"`
	PostType            string `json:"post_type"`
	PostMimeType        string `json:"post_mime_type"`
	CommentCount        string `json:"comment_count"`
	Filter              string `json:"filter"`
	EventStartDate      string `json:"EventStartDate"`
	EventEndDate        string `json:"EventEndDate"`
	Init                string `json:"init"`
	Content             string `json:"content"`
	EditLink            string `json:"edit_link"`
	EventURL            string `json:"event_url"`
	Permalink           string `json:"permalink"`
	User                string `json:"user"`
	Author              string `json:"author"`
	AuthorEmail         string `json:"author_email"`
	Type                string `json:"type"`
	Private             string `json:"private"`
	LDAPDepartment      string `json:"ldap_department"`
	DeptsArr            string `json:"depts_arr"`
	Depts               string `json:"depts"`
	Orgs                string `json:"orgs"`
	Cats                Cats  `json:"cats"`
	Category            string `json:"category"`
	CategoryHTML        string `json:"category_html"`
	Title               string `json:"title"`
	Cost                string `json:"cost"`
	Venue               string `json:"venue"`
	VenuePhone          string `json:"venue_phone"`
	VenueRoom           string `json:"venue_room"`
	Map                 string `json:"map"`
	FacebookURL         string `json:"facebook_url"`
	WebsiteURL          string `json:"website_url"`
	TwitterHash         string `json:"twitter_hash"`
	TwitterURL          string `json:"twitter_url"`
	Tags                string `json:"tags"`
	Recurring           string `json:"recurring"`
	RecurrenceURL       string `json:"recurrence_url"`
	CTDSite             string `json:"ctd_site"`
	OrganizersLabel     string `json:"organizers_label"`
	Organizers          string `json:"organizers"`
	OrganizerPhone      string `json:"organizer_phone"`
	OrganizerEmail      string `json:"organizer_email"`
	OrganizerWebsite    string `json:"organizer_website"`
	IsDM                string `json:"is_dm"`
	DMText              string `json:"dm_text"`
	DMDates             string `json:"dm_dates"`
	EventDates          string `json:"event_dates"`
	StartDateTime       string `json:"StartDateTime"`
	TimeFormatted       string `json:"time_formatted"`
	StartDate           string `json:"start_date"`
	StartTS             string `json:"start_ts"`
	EndTS               string `json:"end_ts"`
	ThumbID             string `json:"thumb_id"`
	IMG                 string `json:"img"`
	ThumbURL            string `json:"thumb_url"`
	ThumbURLUncropped   string `json:"thumb_url_uncropped"`
	ThumbURLMedium      string `json:"thumb_url_medium"`
	Debug               string `json:"debug"`
	CSS                 CSS    `json:"css"`
	ArrowDown           string `json:"arrow_down"`
	ArrowUp             string `json:"arrow_up"`
	Hash                string `json:"hash"`
	HashLink            string `json:"hash_link"`
}

type DailyMessage struct {
	ID               uint `json:"id"`
	Description      string
	Title            string
	ShortDescription string // dm_text
	Author           string
	AuthorEmail      string
	Department       string
	Category         string
	Venue            string
}

func GetRawDailyMessages() ([]RawAnnouncements, error) {
	eventsClient := &http.Client{
		Timeout: time.Second * 30, // maximum of 30s
	}

	url := EVENTS_URL
	fmt.Println("URL: " + url + "")

	// send a GET request and get back the response
	res, err := eventsClient.Get(url)
	if err != nil {
		return nil, err
	}
	// defer res.Body.Close()
	//fmt.Printf("%+v\n", res.Body)
	// body, err := ioutil.ReadAll(res.Body)
	//
	// fmt.Println(body)

	//populate the array of RawGeneralAnnouncements
	var rawAnnouncements []RawAnnouncements
	err = json.NewDecoder(res.Body).Decode(&rawAnnouncements)

	if err != nil {
		fmt.Println("returning badly")
		return nil, err
	}

	fmt.Printf("rawAnnouncements!!! %+v", rawAnnouncements)
	return rawAnnouncements, nil
}

// func ParseDailyMessages(rawMessages []RawGeneralAnnouncement) ([]DailyMessage, err){
//   var messages []DailyMessage
//   for rawMessage, _ := range rawMessages {
//       var message DailyMessage
//
//       messages = append(messages,message)
//   }
//   return
// }
// func GetDailyMessages() (map[string]DailyMessage, error) {
//   return nil, nil
// }
