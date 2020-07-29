package catalog_update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/WilliamsStudentsOnline/wso-go/lib/search/factrak"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	// CatalogURL stores the endpoint for the catalog
	CatalogURL      = "https://catalog.williams.edu/wp-json/courses/v1/year"
	DraftCatalogURL = "https://catalog.draft.williams.edu/wp-json/courses/v1/year"
	hourFormat      = "15:04"
)

// Instructor holds the url and name of the instructors
type Instructor struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// Meeting holds the information relevant to the weekly class meetings
type Meeting struct {
	Days     string `json:"days"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Facility string `json:"facility"`
}

// Attributes consolidates the divisional/distributional/additional options as boolean variables
type Attributes struct {
	Div1        bool `json:"div1"`
	Div2        bool `json:"div2"`
	Div3        bool `json:"div3"`
	DPE         bool `json:"dpe"`
	QFR         bool `json:"qfr"`
	WAC         bool `json:"wac"`
	PassFail    bool `json:"passFail"`
	FifthCourse bool `json:"fifthCourse"`
}

// Course represents the parsed useful information of a Williams Course
type Course struct {
	Year                 int           `json:"year"`
	Semester             string        `json:"semester"`
	CourseID             string        `json:"courseID"`
	Department           string        `json:"department"`
	Number               int           `json:"number"`
	Section              string        `json:"section"`
	SectionType          string        `json:"sectionType"`
	PeoplesoftNumber     int           `json:"peoplesoftNumber"`
	Consent              string        `json:"consent"`
	GradingBasisDesc     string        `json:"gradingBasisDesc"`
	ClassType            string        `json:"classType"`
	TitleLong            string        `json:"titleLong"`
	TitleShort           string        `json:"titleShort"`
	Instructors          []*Instructor `json:"instructors"`
	Meetings             []*Meeting    `json:"meetings"`
	CourseAttributes     Attributes    `json:"courseAttributes"`
	ClassFormat          string        `json:"classFormat"`
	ClassReqEval         string        `json:"classReqEval"`
	ExtraInfo            string        `json:"extraInfo"`
	Prereqs              string        `json:"prereqs"`
	DepartmentNotes      string        `json:"departmentNotes"`
	DescriptionSearch    string        `json:"descriptionSearch"`
	EnrolmentPreferences string        `json:"enrolmentPreferences"`
	CrossListing         []string      `json:"crossListing"`
	Components           []string      `json:"components"`

	// JSON will not marshal these
	crossListingMap map[string]bool `json:"-"`
	componentsMap   map[string]bool `json:"-"`
}

// RawCourse represents the unparsed course information we get from the catalog endpoint.
type RawCourse struct {
	AcademicYear         int    `json:"WMS_ACAD_YEAR,string"`
	Offered              string `json:"OFFERED"`
	Semester             int    `json:"STRM,string"`
	CourseID             string `json:"CRSE_ID"`
	EffectiveDate        string `json:"EFFDT"`
	Subject              string `json:"SUBJECT"`
	CatalogNumber        int    `json:"CATALOG_NBR,string"`
	CourseLetter         string `json:"WMS_CRSE_LETTER"`
	ClassSection         string `json:"CLASS_SECTION"`
	ClassNumber          int    `json:"CLASS_NBR,string"`
	Consent              string `json:"CONSENT"`
	GradingBasis         string `json:"GRADING_BASIS"`
	SSRComponent         string `json:"SSR_COMPONENT"`
	Description          string `json:"DESCR"`
	UnitsMinimum         string `json:"UNITS_MINIMUM"`
	CourseTitleLong      string `json:"COURSE_TITLE_LONG"`
	FirstName1           string `json:"WMS_FIRST_NAME1"`
	MiddleName1          string `json:"WMS_MID_NAME1"`
	LastName1            string `json:"WMS_LAST_NAME1"`
	URL1                 string `json:"URL_1"`
	FirstName2           string `json:"WMS_FIRST_NAME2"`
	MiddleName2          string `json:"WMS_MID_NAME2"`
	LastName2            string `json:"WMS_LAST_NAME2"`
	URL2                 string `json:"URL_2"`
	FirstName3           string `json:"WMS_FIRST_NAME3"`
	MiddleName3          string `json:"WMS_MID_NAME3"`
	LastName3            string `json:"WMS_LAST_NAME3"`
	URL3                 string `json:"URL_3"`
	FirstName4           string `json:"WMS_FIRST_NAME4"`
	MiddleName4          string `json:"WMS_MID_NAME4"`
	LastName4            string `json:"WMS_LAST_NAME4"`
	URL4                 string `json:"URL_4"`
	FirstName5           string `json:"WMS_FIRST_NAME5"`
	MiddleName5          string `json:"WMS_MID_NAME5"`
	LastName5            string `json:"WMS_LAST_NAME5"`
	URL5                 string `json:"URL_5"`
	FirstName6           string `json:"WMS_FIRST_NAME6"`
	MiddleName6          string `json:"WMS_MID_NAME6"`
	LastName6            string `json:"WMS_LAST_NAME6"`
	URL6                 string `json:"URL_6"`
	StandardMeeting1     string `json:"WMS_STND_MTG_PAT1"`
	StartTime1           string `json:"WMS_START_TIME1"`
	EndTime1             string `json:"WMS_END_TIME1"`
	Facility1            string `json:"WMS_FACIL_DESCR1"`
	StandardMeeting2     string `json:"WMS_STND_MTG_PAT2"`
	StartTime2           string `json:"WMS_START_TIME2"`
	EndTime2             string `json:"WMS_END_TIME2"`
	Facility2            string `json:"WMS_FACIL_DESCR2"`
	StandardMeeting3     string `json:"WMS_STND_MTG_PAT3"`
	StartTime3           string `json:"WMS_START_TIME3"`
	EndTime3             string `json:"WMS_END_TIME3"`
	Facility3            string `json:"WMS_FACIL_DESCR"`
	AttributesSearch     string `json:"WMS_ATTR_SRCH"`
	ClassFormat          string `json:"WMS_CLASS_FORMAT"`
	Evaluation           string `json:"WMS_RQMT_EVAL"`
	ExtraInfo            string `json:"WMS_EXTRA_INFO"`
	ExtraInfo2           string `json:"WMS_EXTRA_INFO2"`
	PreReqs              string `json:"WMS_PREREQS"`
	EnrollmentPreference string `json:"WMS_ENRL_PREF"`
	DepartmentNotes      string `json:"WMS_DEPT_NOTES"`
	MaterialFee          string `json:"WMS_MATL_FEE"`
	ExperientialLearning string `json:"WMS_EXP_ENRL"`
	EnrollmentLimit      string `json:"WMS_ENRL_LIMIT"`
	Campus               string `json:"CAMPUS"`
	Description140       string `json:"WMS_DESCR140"`
	ShortDescription     string `json:"WMS_SHORT_DESCR"`
	DistributionNote1    string `json:"WMS_DISTRIB_NT1"`
	DistributionNote2    string `json:"WMS_DISTRIB_NT2"`
	DistributionNote3    string `json:"WMS_DISTRIB_NT3"`
	DescriptionSearch    string `json:"WMS_DESCR_SRCH"`
	DistributionNotes    string `json:"WMS_DISTRIB_NOTES"`
}

type exportCourses struct {
	Courses    []Course `json:"courses"`
	UpdateTime string   `json:"updateTime"`
}

// ParseCatalog processes the raw byte data from the JSON endpoint to obtain Course objects
func ParseCatalog(catalog []RawCourse, fallSemID, winterSemID, springSemID int, log *zap.SugaredLogger, searchFactrak factrak.SearchFactrak) ([]Course, error) {
	// Initialize the slice this way in order to ensure it will never respond as a nil slice
	courses := []Course{}

	// Map names to IDs of profs we've found in this cache.
	// This speeds up lookup times for searching for profs significantly
	profIDCache := make(map[string]uint)

	for _, unparsed := range catalog {
		if unparsed.Offered != "Y" || unparsed.Facility1 == "Cancelled" {
			continue
		}
		course := Course{}

		course.Year = unparsed.AcademicYear

		semID := unparsed.Semester
		switch semID {
		case fallSemID:
			course.Semester = "Fall"
		case winterSemID:
			course.Semester = "Winter"
		case springSemID:
			course.Semester = "Spring"
		default:
			course.Semester = "Unknown"
		}

		course.CourseID = strings.TrimSpace(unparsed.CourseID)
		course.Department = strings.TrimSpace(unparsed.Subject)
		course.Number = unparsed.CatalogNumber

		// Tutorial sections start with 'T'
		course.Section = strings.TrimSpace(unparsed.ClassSection)
		course.PeoplesoftNumber = unparsed.ClassNumber

		// Parse if remote or hybrid or in-person
		if len(course.Section) > 0 {
			switch unicode.ToUpper(rune(course.Section[0])) {
			case 'R':
				course.SectionType = "remote"
			case 'H':
				course.SectionType = "hybrid"
			default:
				course.SectionType = "in-person"
			}
		}

		// Options for Consent are 'N', ' ', 'D
		course.Consent = strings.TrimSpace(unparsed.Consent)

		// Options for GRADING_BASIS are OPT,GRD,OPX,OPP, ,WPP,PF4,NON,PNP,XEG,PF5
		passFail := false
		fifthCourse := false
		switch strings.TrimSpace(unparsed.GradingBasis) {
		case "WPP":
			course.GradingBasisDesc = "Winter Study"
		case "GRD":
			course.GradingBasisDesc = "No Pass/Fail and No Fifth Course"
		case "OPT":
			course.GradingBasisDesc = "Pass/Fail Available, Fifth Course Available"
			passFail = true
			fifthCourse = true
		case "OPX":
			course.GradingBasisDesc = "Pass/Fail Unavailable, Fifth Course Available"
			fifthCourse = true
		case "OPP":
			course.GradingBasisDesc = "Pass/Fail Available, Fifth Course Unavailable"
			passFail = true
		case "PNP":
			course.GradingBasisDesc = "Pass/Fail Option Only"
			passFail = true
			fifthCourse = true
		case "PF4":
			course.GradingBasisDesc = "Pass/Fail Option Only"
			passFail = true
		case "NON":
			course.GradingBasisDesc = "Non-Graded"
		}

		ssrComponent := strings.TrimSpace(unparsed.SSRComponent)
		switch ssrComponent {
		case "LEC":
			course.ClassType = "Lecture"
		case "SEM":
			course.ClassType = "Seminar"
		case "TUT":
			course.ClassType = "Tutorial"
		case "STU":
			course.ClassType = "Studio"
		case "HON":
			course.ClassType = "Honors"
		case "CON":
			course.ClassType = "Conference"
		case "IND":
			course.ClassType = "Independent Study"
		case "LAB":
			course.ClassType = "Laboratory"
		default:
			course.ClassType = ssrComponent
		}

		course.TitleLong = trimTitle(unparsed.CourseTitleLong)
		course.TitleShort = trimTitle(unparsed.Description)

		// Instructors

		names := []struct {
			firstName  string
			middleName string
			lastName   string
		}{
			{unparsed.FirstName1, unparsed.MiddleName1, unparsed.LastName1},
			{unparsed.FirstName2, unparsed.MiddleName2, unparsed.LastName2},
			{unparsed.FirstName3, unparsed.MiddleName3, unparsed.LastName3},
			{unparsed.FirstName4, unparsed.MiddleName4, unparsed.LastName4},
			{unparsed.FirstName5, unparsed.MiddleName5, unparsed.LastName5},
			{unparsed.FirstName6, unparsed.MiddleName6, unparsed.LastName6},
		}

		// Parse instructors

		for _, instructorName := range names {
			instructor := Instructor{}

			fn := strings.TrimSpace(instructorName.firstName)
			mn := strings.TrimSpace(instructorName.middleName)
			ln := strings.TrimSpace(instructorName.lastName)

			if fn == "" {
				break
			}

			name := fn
			if mn != "" {
				name += " " + mn
			}
			name += " " + ln

			instructor.Name = name

			// Factrak search:
			if searchFactrak != nil {
				// Look up prof name in cache
				cachedID, ok := profIDCache[instructor.Name]
				if ok {
					// If hit, use it
					instructor.ID = cachedID
				} else {
					// If miss, search in DB via factrak search engine and then add it to the cache
					var err error
					instructor.ID, err = SearchProfessor(searchFactrak, fn, mn, ln)
					if err != nil {
						log.With("error", err).Error("Failed to find professor from factrak")
					}
					profIDCache[instructor.Name] = instructor.ID
				}
			}

			course.Instructors = append(course.Instructors, &instructor)
		}

		// Class Meetings

		unparsedMeetings := []struct {
			days     string
			start    string
			end      string
			facility string
		}{
			{unparsed.StandardMeeting1, unparsed.StartTime1, unparsed.EndTime1, unparsed.Facility1},
			{unparsed.StandardMeeting2, unparsed.StartTime2, unparsed.EndTime2, unparsed.Facility2},
			{unparsed.StandardMeeting3, unparsed.StartTime3, unparsed.EndTime3, unparsed.Facility3},
		}

		for _, unparsedMeeting := range unparsedMeetings {
			meeting := Meeting{}

			// Different options: MW, ,TR,MWF,W,TF,TBA,MR,T,M,R,M-F,F
			days := strings.TrimSpace(unparsedMeeting.days)
			if days == "" {
				continue
			} else if days == "TBA" {
				break
			}

			meeting.Days = days

			startT := strings.TrimSpace(unparsedMeeting.start)
			if startT == "" {
				meeting.Start = ""
			} else {
				startTime, err := time.Parse(hourFormat, startT)
				if err != nil {
					return nil, err
				}

				meeting.Start = startTime.Format(hourFormat)
			}

			endT := strings.TrimSpace(unparsedMeeting.end)
			if endT == "" {
				meeting.End = ""
			} else {
				endTime, err := time.Parse(hourFormat, endT)
				if err != nil {
					return nil, err
				}

				meeting.End = endTime.Format(hourFormat)
			}

			meeting.Facility = trimTitle(unparsedMeeting.facility)
			course.Meetings = append(course.Meetings, &meeting)
		}

		unparsedAttributes := unparsed.AttributesSearch
		course.CourseAttributes = Attributes{
			Div1:        strings.Contains(unparsedAttributes, "DIV_D1"),
			Div2:        strings.Contains(unparsedAttributes, "DIV_D2"),
			Div3:        strings.Contains(unparsedAttributes, "DIV_D3"),
			DPE:         strings.Contains(unparsedAttributes, "DPE_DPE"),
			QFR:         strings.Contains(unparsedAttributes, "QFR_QFR"),
			WAC:         strings.Contains(unparsedAttributes, "WAC_WAC"),
			PassFail:    passFail,
			FifthCourse: fifthCourse,
		}

		course.ClassFormat = trimTitle(unparsed.ClassFormat)
		course.ClassReqEval = trimCapitalize(unparsed.Evaluation)

		course.ExtraInfo = strings.TrimSpace(unparsed.ExtraInfo)
		if strings.TrimSpace(unparsed.ExtraInfo2) != "" {
			course.ExtraInfo += "; " + strings.TrimSpace(unparsed.ExtraInfo2)
		}

		course.Prereqs = trimCapitalize(unparsed.PreReqs)
		course.DepartmentNotes = trimCapitalize(unparsed.DepartmentNotes)

		course.DescriptionSearch = trimCapitalize(unparsed.DescriptionSearch)
		course.EnrolmentPreferences = trimCapitalize(unparsed.EnrollmentPreference)

		courses = append(courses, course)
	}

	UpdateCrossListing(courses)

	return courses, nil
}

// UpdateCrossListing takes the parsed array of courses and updates their cross-listing information.
func UpdateCrossListing(courses []Course) {
	// Sort courses by CourseID, since cross-listed courses all have the same CourseID, so we
	// only need to do one pass through the array
	sort.SliceStable(courses, func(i, j int) bool {
		return courses[i].CourseID < courses[j].CourseID
	})

	currCourse := ""

	// These maps act as sets to collect and store the information throughout cross-listings
	var crossListings = map[string]bool{}
	var components = map[string]bool{}

	for i := range courses {
		// If the current course has a different CourseID, then we must have finished finding all
		// cross-listings
		if currCourse != courses[i].CourseID {
			currCourse = courses[i].CourseID
			crossListings = make(map[string]bool)
			components = make(map[string]bool)
		}

		crossListings[courses[i].Department+" "+strconv.Itoa(courses[i].Number)] = true
		components[courses[i].ClassType] = true
		courses[i].crossListingMap = crossListings
		courses[i].componentsMap = components
	}

	// After the maps are populated, go through the array again to populate the exportable arrays.
	for i := range courses {
		courses[i].CrossListing = make([]string, len(courses[i].crossListingMap))
		j := 0
		for k := range courses[i].crossListingMap {
			courses[i].CrossListing[j] = k
			j++
		}

		courses[i].Components = make([]string, len(courses[i].componentsMap))
		j = 0
		for k := range courses[i].componentsMap {
			courses[i].Components[j] = k
			j++
		}
	}

	// Sort by course code before returning
	sort.SliceStable(courses, func(i, j int) bool {
		return (courses[i].Department < courses[j].Department) ||
			(courses[i].Department == courses[j].Department && courses[i].Number < courses[j].Number)
	})
}

func SearchProfessor(search factrak.SearchFactrak, fn, mn, ln string) (uint, error) {
	completeName := fn // includes middle name
	partialName := fn  // doesnt include middle name
	if mn != "" {
		completeName += " " + mn
	}
	completeName += " " + ln
	partialName += " " + ln

	var completeNameProfs []*models.User
	err := search.SearchProfessors(completeName, &completeNameProfs, nil)
	if err != nil {
		return 0, err
	}

	var partialNameProfs []*models.User
	err = search.SearchProfessors(partialName, &partialNameProfs, nil)
	if err != nil {
		return 0, err
	}

	if len(completeNameProfs) == 1 {
		return completeNameProfs[0].ID, nil
	} else if len(partialNameProfs) == 1 {
		return partialNameProfs[0].ID, nil
	} else {
		return 0, nil
	}
}

// GetCatalog fetches the json from the CatalogURL endpoint and parses it into an array of RawCourses.
func GetCatalog(academicYear int, draft bool) ([]RawCourse, error) {
	catalogClient := &http.Client{
		Timeout: time.Second * 30, // Maximum of 30 seconds
	}

	var url string
	if draft {
		url = fmt.Sprintf("%s/%d", DraftCatalogURL, academicYear)
	} else {
		url = fmt.Sprintf("%s/%d", CatalogURL, academicYear)
	}

	// Send the GET request and get back the response
	res, err := catalogClient.Get(url)
	if err != nil {
		return nil, err
	}

	var rawCourses []RawCourse
	err = json.NewDecoder(res.Body).Decode(&rawCourses)
	if err != nil {
		return nil, err
	}

	return rawCourses, nil
}

// SaveCatalog writes the array of courses and the update time into the provided writer.
func SaveCatalog(w io.Writer, courses []Course) error {
	var catalog = exportCourses{}
	catalog.Courses = courses
	catalog.UpdateTime = time.Now().Format(time.RFC850)

	return json.NewEncoder(w).Encode(catalog)
}

func AttachDBProfessors(courses []*Course, db *gorm.DB) error {
	for _, course := range courses {
		for _, instructor := range course.Instructors {
			prof := models.User{}
			err := db.Where("type = ?", models.UserTypeProfessor).
				Where("name = ?", instructor.Name).
				First(&prof).Error
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					instructor.ID = 0
					continue
				} else {
					return err
				}
			}

			instructor.ID = prof.ID
		}
	}

	return nil
}

// capitalize capitalizes the first letter of the string.
func capitalize(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

// trimCapitalize trims leading/following white spaces and Capitalizes the first letter of the string.
func trimCapitalize(str string) string {
	return capitalize(strings.TrimSpace(str))
}

// trimTitle trims leading/following white spaces and Title Cases the string.
func trimTitle(str string) string {
	return strings.Title(strings.TrimSpace(str))
}
