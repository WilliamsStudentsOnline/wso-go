package events_update

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"log"
	//"strconv"
	//"strings"
	"time"
	//"unicode"
)

const (
	// EVENTS_URL stores the endpoint for the daily announcements
	EVENTS_URL = "https://events.williams.edu/wp-json/wms/events/v1/list/dm"
)


type Cats struct {
	TermID         int `json:"term_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	TermGroup      int `json:"term_group"`
	TermTaxonomyID int `json:"term_taxonomy_id"`
	Taxonomy       string `json:"taxonomy"`
	Description    string `json:"description"`
	Parent         int `json:"parent"`
	Count          int `json:"count"`
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


type RawCategory struct {
	RawCategoryName string
	RawDailyMessages []RawDailyMessage
}

type RawDailyMessage struct {
	ID                  uint     `json:"ID"`
	PostAuthor          string `json:"post_author"`
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
	PostParent          int `json:"post_parent"`
	GUID                string `json:"guid"`
	MenuOrder           int `json:"menu_order"`
	PostType            string `json:"post_type"`
	PostMimeType        string `json:"post_mime_type"`
	CommentCount        string `json:"comment_count"`
	Filter              string `json:"filter"`
	EventStartDate      string `json:"EventStartDate"`
	EventEndDate        string `json:"EventEndDate"`
	Init                bool `json:"init"`
	Content             string `json:"content"`
	EditLink            interface{}  `json:"edit_link"`
	EventURL            string `json:"event_url"`
	Permalink           string `json:"permalink"`
	User                string `json:"user"`
	Author              string `json:"author"`
	AuthorEmail         string `json:"author_email"`
	Type                string `json:"type"`
	Private             interface{}  `json:"private"`
	LDAPDepartment      string `json:"ldap_department"`
	DeptsArr            []interface{} `json:"depts_arr"`
	Depts               string `json:"depts"`
	Orgs                string `json:"orgs"`
	Cats                []Cats  `json:"cats"`
	Category            string `json:"category"`
	CategoryHTML        string `json:"category_html"`
	Title               string `json:"title"`
	Cost                string `json:"cost"`
	Venue               interface{} `json:"venue"`
	VenuePhone          string `json:"venue_phone"`
	VenueRoom           string `json:"venue_room"`
	Map                 string `json:"map"`
	FacebookURL         string `json:"facebook_url"`
	WebsiteURL          string `json:"website_url"`
	TwitterHash         string `json:"twitter_hash"`
	TwitterURL          string `json:"twitter_url"`
	Tags                bool `json:"tags"`
	Recurring           bool `json:"recurring"`
	RecurrenceURL       string `json:"recurrence_url"`
	CTDSite             []interface{} `json:"ctd_site"`
	OrganizersLabel     string `json:"organizers_label"`
	Organizers          interface{} `json:"organizers"`
	OrganizerPhone      string `json:"organizer_phone"`
	OrganizerEmail      string `json:"organizer_email"`
	OrganizerWebsite    string `json:"organizer_website"`
	IsDM                bool `json:"is_dm"`
	DMText              string `json:"dm_text"`
	DMDates             []string  `json:"dm_dates"`
	EventDates          []string  `json:"event_dates"`
	StartDateTime       string `json:"StartDateTime"`
	TimeFormatted       string `json:"time_formatted"`
	StartDate           string `json:"start_date"`
	StartTS             string `json:"start_ts"`
	EndTS               string `json:"end_ts"`
	ThumbID             bool `json:"thumb_id"`
	IMG                 string `json:"img"`
	ThumbURL            bool `json:"thumb_url"`
	ThumbURLUncropped   bool `json:"thumb_url_uncropped"`
	ThumbURLMedium      bool `json:"thumb_url_medium"`
	Debug               bool `json:"debug"`
	CSS                 CSS    `json:"css"`
	ArrowDown           string `json:"arrow_down"`
	ArrowUp             string `json:"arrow_up"`
	Hash                string `json:"hash"`
	HashLink            string `json:"hash_link"`
}


type MessageCategory struct {
	CategoryName string `json:"category_name"`
	DailyMessages []DailyMessage `json:"daily_messages"`
}
type DailyMessage struct {
	ID               uint `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	ShortDescription string `json:"short_description"`// dm_text
	Author           string `json:"author"`
	AuthorEmail      string `json:"author_email"`
	Department       string `json:"dept"`
	// Category         string UNNCESSARY SINCE WE HAVE PARENT STRUCT
	Venue            interface{} `json:"venue"`
}

func GetRawCategories() ([]RawCategory, error) {
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
	defer res.Body.Close()

	// decode get request's body into a raw data map
	var rawData map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&rawData)
	if err != nil {
		return nil, err
	}

	// define an array of RawCategory that will contain -- key (category name )and an []RDM
	var rawCategories []RawCategory
	// convert raw data map into an array of RawCategory
	for k, v := range rawData {
		var rawCategory RawCategory
		// define the header name as the key
		rawCategory.RawCategoryName = k

		// convert interface{} to JSON string
		JSONString, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		// convert JSON string to an array of RDM structs
    var rawDailyMessages []RawDailyMessage
    json.Unmarshal(JSONString, &rawDailyMessages)
		rawCategory.RawDailyMessages = rawDailyMessages

		// append this fully written RawCategory to rawCategories to return
		rawCategories = append(rawCategories, rawCategory)
	}
	return rawCategories, nil
}

func ParseDailyMessages(rawCategories []RawCategory) ([]MessageCategory, error){
	var categories []MessageCategory
	// iterate through each rawCategory in rawCategories
  for _ , rawCategory := range rawCategories {
			// create a new MessageCategory for each rawCategory you see
			var category MessageCategory
			category.CategoryName = rawCategory.RawCategoryName
			// traverse each raw daily message in the rawCategory's array of RawDailyMessages
			for _ ,rawDailyMessage := range rawCategory.RawDailyMessages {
				// create a new DailyMessage for each RawDailyMessage you see
				var message DailyMessage
				message.ID = rawDailyMessage.ID
				message.Title = rawDailyMessage.PostTitle
				message.ShortDescription = rawDailyMessage.DMText
				message.Description = rawDailyMessage.PostContent
				message.Author = rawDailyMessage.Author
				message.AuthorEmail = rawDailyMessage.AuthorEmail
				message.Department = rawDailyMessage.LDAPDepartment
				message.Venue = rawDailyMessage.Venue
				// append this message to its appropriate category
				category.DailyMessages = append(category.DailyMessages, message)
			}
    	// append category to category array
			categories = append(categories, category)
  }

  return categories, nil
}

func DumpCategories(categories []MessageCategory){
	for _, category := range categories {
		fmt.Println("category_name: " + category.CategoryName + "\n")
		for _, message := range category.DailyMessages {
			fmt.Printf("\tid: %d\n",message.ID)
			fmt.Println("\ttitle: " + message.Title)
			fmt.Println("\tshort_description: " + message.ShortDescription)
			fmt.Println("\tdescription: " + message.Description)
			fmt.Println("\tauthor: " + message.Author)
			fmt.Println("\tautor_email: " + message.AuthorEmail)
			fmt.Println("\tdept: " + message.Department)
			fmt.Print("\tvenue: ")
			describe(message.Venue)
			fmt.Println("**************************************************************")
		}
		fmt.Println("\n")
	}
}

func describe(i interface{}) {
	fmt.Printf("(%v, %T)\n", i, i)
}

// takes the Menu struct, exports data the JSON and writes it to a local file
//	returns date as well
func WriteToJSON(categories []MessageCategory) (string, error) {
	eventsJSON, err:= json.Marshal(categories)
	if err != nil{
		return "", err
	}
	err = ioutil.WriteFile("events_data.json", eventsJSON, 0644)
	if err != nil {
		return "", err
	}
	dt := time.Now()

	return dt.Format("01-02-2006 15:04:05 Mon"), err
}
