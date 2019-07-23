package lib

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Instructor holds the url and name of the isntructors
type Instructor struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Meeting holds the information relevant to the weekly class meetings
type Meeting struct {
	Days  string `json:"days"`
	Start string `json:"start"`
	End   string `json:"end"`
	Facil string `json:"facil"`
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
	Year              int          `json:"year"`
	Semester          string       `json:"semester"`
	CourseID          string       `json:"courseID"`
	Department        string       `json:"department"`
	Number            int          `json:"number"`
	Section           string       `json:"section"`
	PeoplesoftNumber  int          `json:"peoplesoftNumber"`
	Consent           string       `json:"consent"`
	GradingBasis      string       `json:"gradingBasis"`
	GradingBasisDesc  string       `json:"gradingBasisDesc"`
	ClassType         string       `json:"classType"`
	TitleLong         string       `json:"titleLong"`
	TitleShort        string       `json:"titleShort"`
	Instructors       []Instructor `json:"instructors"`
	Meetings          []Meeting    `json:"meetings"`
	CourseAttributes  Attributes   `json:"courseAttributes"`
	ClassFormat       string       `json:"classFormat"`
	ClassReqEval      string       `json:"classReqEval"`
	ExtraInfo         string       `json:"extraInfo"`
	Prereqs           string       `json:"prereqs"`
	DepartmentNotes   string       `json:"departmentNotes"`
	DescriptionSearch string       `json:"descriptionSearch"`
	EnrlPref          string       `json:"enrlPref"`
}

type uCourse struct {
	AcademicYear         int    `json:"WMS_ACAD_YEAR,string"`
	Offered              string `json:"Offered"`
	STRM                 int    `json:"STRM,string"`
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
	WMSINSTROTH          string `json:"WMS_INSTR_OTH"` // No idea what this is
	PreReqs              string `json:"WMS_PREREQS"`
	EnrollmentPreference string `json:"WMS_ENRL_PREF"`
	DepartmentNotes      string `json:"WMS_DEPT_NOTES"`
	MaterialFee          string `json:"WMS_MATL_FEE"`
	ExperientialLearning string `json:"WMS_EXP_ENRL"`
	EnrollmentLimit      string `json:"WMS_ENRL_LIMIT"`
	WMSNC                string `json:"WMS_NC"` // No idea what this is
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

const (
	fallSemesterID   = 1201
	winterSemesterID = 1202
	springSemesterID = 1203
	twentyFourHour   = "15:04"
)

// capitalize capitalizes the first letter of the string
func capitalize(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

// trimCapitalize trims leading/following white spaces and Capitalizes the first letter of the string
func trimCapitalize(str string) string {
	return capitalize(strings.TrimSpace(str))
}

// trimTitle trims leading/following white spaces and Title Cases the string
func trimTitle(str string) string {
	return strings.Title(strings.TrimSpace(str))
}

// ParseCatalog processes the raw byte data from the JSON endpoint to obtain Course objects
func ParseCatalog(catalog []byte) ([]Course, error) {
	var unparsedCourses = []uCourse{}
	err := json.Unmarshal(catalog, &unparsedCourses)
	if err != nil {
		return nil, err
	}

	courses := []Course{}

	for _, unparsed := range unparsedCourses {
		if unparsed.Offered != "Y" || unparsed.Facility1 == "Cancelled" {
			continue
		}
		course := Course{}

		course.Year = unparsed.AcademicYear

		semID := unparsed.STRM
		switch semID {
		case fallSemesterID:
			course.Semester = "FALL"
		case winterSemesterID:
			course.Semester = "WINTER"
		case springSemesterID:
			course.Semester = "SPRING"
		default:
			course.Semester = "UNKNOWN"
		}

		course.CourseID = strings.TrimSpace(unparsed.CourseID)
		course.Department = strings.TrimSpace(unparsed.Subject)
		course.Number = unparsed.CatalogNumber

		// Tutorial sections start with 'T'
		course.Section = strings.TrimSpace(unparsed.ClassSection)
		course.PeoplesoftNumber = unparsed.ClassNumber

		// Options for Consent are 'N', ' ', 'D
		course.Consent = strings.TrimSpace(unparsed.Consent)

		// Options for GRADING_BASIS are OPT,GRD,OPX,OPP, ,WPP,PF4,NON,PNP,XEG,PF5
		course.GradingBasis = strings.TrimSpace(unparsed.GradingBasis)
		passFail := false
		fifthCourse := false
		switch course.GradingBasis {
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
		case "IND":
			course.ClassType = "Independent Study"
		case "LAB":
			course.ClassType = "Laboratory"
		default:
			course.ClassType = ssrComponent
		}

		course.TitleLong = trimTitle(unparsed.CourseTitleLong)
		course.TitleShort = trimTitle(unparsed.Description)

		course.Instructors = []Instructor{}

		names := []struct {
			firstName  string
			middleName string
			lastName   string
		}{
			{
				unparsed.FirstName1, unparsed.MiddleName1, unparsed.LastName1,
			},
			{
				unparsed.FirstName2, unparsed.MiddleName2, unparsed.LastName2,
			},
			{
				unparsed.FirstName3, unparsed.MiddleName3, unparsed.LastName3,
			},
			{
				unparsed.FirstName4, unparsed.MiddleName4, unparsed.LastName4,
			},
			{
				unparsed.FirstName5, unparsed.MiddleName5, unparsed.LastName5,
			},
			{
				unparsed.FirstName6, unparsed.MiddleName6, unparsed.LastName6,
			},
		}

		for _, instructorName := range names {
			instructor := Instructor{}

			fn := strings.TrimSpace(instructorName.firstName)
			mn := strings.TrimSpace(instructorName.middleName)
			ln := strings.TrimSpace(instructorName.lastName)

			if fn == "" {
				continue
			}

			name := fn
			if mn != "" {
				name += " " + mn
			}
			name += " " + ln

			instructor.Name = name

			// @TODO include factrak search
			instructor.URL = ""

			course.Instructors = append(course.Instructors, instructor)
		}

		course.Meetings = []Meeting{}
		unparsedMeetings := []struct {
			days     string
			start    string
			end      string
			facility string
		}{
			{
				unparsed.StandardMeeting1, unparsed.StartTime1, unparsed.EndTime1, unparsed.Facility1,
			},
			{
				unparsed.StandardMeeting2, unparsed.StartTime2, unparsed.EndTime2, unparsed.Facility2,
			},
			{
				unparsed.StandardMeeting3, unparsed.StartTime3, unparsed.EndTime3, unparsed.Facility3,
			},
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
				startTime, err := time.Parse(twentyFourHour, startT)
				if err != nil {
					return nil, err
				}

				meeting.Start = startTime.Format(twentyFourHour)
			}

			endT := strings.TrimSpace(unparsedMeeting.end)
			if endT == "" {
				meeting.End = ""
			} else {
				endTime, err := time.Parse(twentyFourHour, endT)
				if err != nil {
					return nil, err
				}

				meeting.End = endTime.Format(twentyFourHour)
			}

			meeting.Facil = trimTitle(unparsedMeeting.facility)
			course.Meetings = append(course.Meetings, meeting)
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
		course.EnrlPref = trimCapitalize(unparsed.EnrollmentPreference)

		courses = append(courses, course)
	}

	return courses, nil
}

func grabCatalog() ([]byte, error) {

	url := "https://catalog.williams.edu/wp-json/courses/v1/year/1920"

	catalogClient := &http.Client{
		Timeout: time.Second * 30, // Maximum of 30 seconds
	}

	// Send the GET request and get back the response
	res, err := catalogClient.Get(url)
	if err != nil {
		return nil, err
	}

	// Parse the body into []byte
	return ioutil.ReadAll(res.Body)
}

func exportCatalog(courses []Course) error {
	var catalog = exportCourses{}
	catalog.Courses = courses
	catalog.UpdateTime = time.Now().Format(time.RFC850)

	catalogJSON, err := json.Marshal(catalog)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("courses.json", catalogJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

func updateCatalog() error {
	responseBody, err := grabCatalog()
	if err != nil {
		return err
	}
	courses, err := ParseCatalog(responseBody)
	if err != nil {
		return err
	}
	return exportCatalog(courses)
}
