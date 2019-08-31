package events_update

// https://events.williams.edu/wp-json/wms/events/v1/list/dm

type RawDailyMessageEvent struct {
	ID                  int           `json:"ID"`
	PostAuthor          string        `json:"post_author"`
	PostDate            string        `json:"post_date"`
	PostDateGmt         string        `json:"post_date_gmt"`
	PostContent         string        `json:"post_content"`
	PostTitle           string        `json:"post_title"`
	PostExcerpt         string        `json:"post_excerpt"`
	PostStatus          string        `json:"post_status"`
	CommentStatus       string        `json:"comment_status"`
	PingStatus          string        `json:"ping_status"`
	PostPassword        string        `json:"post_password"`
	PostName            string        `json:"post_name"`
	ToPing              string        `json:"to_ping"`
	Pinged              string        `json:"pinged"`
	PostModified        string        `json:"post_modified"`
	PostModifiedGmt     string        `json:"post_modified_gmt"`
	PostContentFiltered string        `json:"post_content_filtered"`
	PostParent          int           `json:"post_parent"`
	GUID                string        `json:"guid"`
	MenuOrder           int           `json:"menu_order"`
	PostType            string        `json:"post_type"`
	PostMimeType        string        `json:"post_mime_type"`
	CommentCount        string        `json:"comment_count"`
	Filter              string        `json:"filter"`
	EventStartDate      string        `json:"EventStartDate"`
	EventEndDate        string        `json:"EventEndDate"`
	Init                bool          `json:"init"`
	Content             string        `json:"content"`
	EditLink            interface{}   `json:"edit_link"`
	EventURL            string        `json:"event_url"`
	Permalink           string        `json:"permalink"`
	User                string        `json:"user"`
	Author              string        `json:"author"`
	AuthorEmail         string        `json:"author_email"`
	Type                string        `json:"type"`
	Private             interface{}   `json:"private"`
	LdapDepartment      string        `json:"ldap_department"`
	DeptsArr            []interface{} `json:"depts_arr"`
	Depts               string        `json:"depts"`
	Orgs                string        `json:"orgs"`
	Cats                []struct {
		TermID         int    `json:"term_id"`
		Name           string `json:"name"`
		Slug           string `json:"slug"`
		TermGroup      int    `json:"term_group"`
		TermTaxonomyID int    `json:"term_taxonomy_id"`
		Taxonomy       string `json:"taxonomy"`
		Description    string `json:"description"`
		Parent         int    `json:"parent"`
		Count          int    `json:"count"`
		Filter         string `json:"filter"`
		TermOrder      string `json:"term_order"`
	} `json:"cats"`
	Category          string        `json:"category"`
	CategoryHTML      string        `json:"category_html"`
	Title             string        `json:"title"`
	Cost              string        `json:"cost"`
	Venue             interface{}   `json:"venue"`
	VenuePhone        string        `json:"venue_phone"`
	VenueRoom         string        `json:"venue_room"`
	Map               string        `json:"map"`
	FacebookURL       string        `json:"facebook_url"`
	WebsiteURL        string        `json:"website_url"`
	TwitterHash       string        `json:"twitter_hash"`
	TwitterURL        string        `json:"twitter_url"`
	Tags              bool          `json:"tags"`
	Recurring         bool          `json:"recurring"`
	RecurrenceURL     string        `json:"recurrence_url"`
	CtdSite           []interface{} `json:"ctd_site"`
	OrganizersLabel   string        `json:"organizers_label"`
	Organizers        interface{}   `json:"organizers"`
	OrganizerPhone    string        `json:"organizer_phone"`
	OrganizerEmail    string        `json:"organizer_email"`
	OrganizerWebsite  string        `json:"organizer_website"`
	IsDm              bool          `json:"is_dm"`
	DmText            string        `json:"dm_text"`
	DmDates           []string      `json:"dm_dates"`
	EventDates        []string      `json:"event_dates"`
	StartDateTime     string        `json:"StartDateTime"`
	TimeFormatted     string        `json:"time_formatted"`
	StartDate         string        `json:"start_date"`
	StartTs           string        `json:"start_ts"`
	EndTs             string        `json:"end_ts"`
	ThumbID           bool          `json:"thumb_id"`
	Img               string        `json:"img"`
	ThumbURL          bool          `json:"thumb_url"`
	ThumbURLUncropped bool          `json:"thumb_url_uncropped"`
	ThumbURLMedium    bool          `json:"thumb_url_medium"`
	Debug             bool          `json:"debug"`
	CSS               struct {
		Tablewrap   string `json:"tablewrap"`
		Ol          string `json:"ol"`
		LiTop       string `json:"li_top"`
		LiBottom    string `json:"li_bottom"`
		Hr          string `json:"hr"`
		A           string `json:"a"`
		Strong      string `json:"strong"`
		Img         string `json:"img"`
		Table       string `json:"table"`
		Hero        string `json:"hero"`
		Logo        string `json:"logo"`
		ViewOnline  string `json:"view_online"`
		JumpLink    string `json:"jump_link"`
		Btb         string `json:"btb"`
		Heading     string `json:"heading"`
		Toc         string `json:"toc"`
		Title       string `json:"title"`
		Bodytext    string `json:"bodytext"`
		Link        string `json:"link"`
		Count       string `json:"count"`
		TextWIndent string `json:"text_w_indent"`
	} `json:"css"`
	ArrowDown string `json:"arrow_down"`
	ArrowUp   string `json:"arrow_up"`
	Hash      string `json:"hash"`
	HashLink  string `json:"hash_link"`
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

func GetDailyMessages() (map[string]DailyMessage, error) {

}
