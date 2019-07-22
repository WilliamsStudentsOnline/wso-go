package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Instructor struct {
	URL  string
	Name string
}

type Meeting struct {
	Days     string
	Start    string
	Start12  string
	End      string
	End12    string
	FacDescr string
	Facil    string
}

type Attributes struct {
	Div1        bool
	Div2        bool
	Div3        bool
	DPE         bool
	QFR         bool
	WAC         bool
	PassFail    bool
	FifthCourse bool
}

type Course struct {
	Year              int
	Semester          string
	CourseID          int
	Department        string
	Number            int
	Section           string
	PeoplesoftNumber  int
	Consent           string
	GradingBasis      string
	GradingBasisDesc  string
	ClassType         string
	TitleLong         string
	TitleShort        string
	Instructors       []Instructor
	Meetings          []Meeting
	CourseAttributes  Attributes
	ClassFormat       string
	ClassReqEval      string
	ExtraInfo         string
	Prereqs           string
	DepartmentNotes   string
	DescriptionSearch string
	EnrlPref          string
}

type unparsedJSON []map[string]string

const FallSemesterID = 1201
const WinterSemesterID = 1202
const SpringSemesterID = 1203

func assertError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func ParseCatalog(catalog []byte) []Course {
	unparsedCourses := unparsedJSON{}
	err := json.Unmarshal(catalog, &unparsedCourses)
	assertError(err)

	courses := []Course{}

	for _, unparsed := range unparsedCourses {
		if unparsed["OFFERED"] != "Y" || unparsed["WMS_FACIL_DESCR1"] == "Cancelled" {
			continue
		}
		course := Course{}

		course.Year, err = strconv.Atoi(unparsed["WMS_ACAD_YEAR"])
		assertError(err)
		semID, err := strconv.Atoi(unparsed["STRM"])
		assertError(err)

		switch semID {
		case FallSemesterID:
			course.Semester = "FALL"
		case WinterSemesterID:
			course.Semester = "WINTER"
		case SpringSemesterID:
			course.Semester = "SPRING"
		default:
			course.Semester = "UNKNOWN"
		}

		course.CourseID, err = strconv.Atoi(unparsed["CRSE_ID"])
		assertError(err)
		course.Department = unparsed["SUBJECT"]
		course.Number, err = strconv.Atoi(unparsed["CATALOG_NBR"])
		assertError(err)

		// Tutorial sections start with 'T'
		course.Section = unparsed["CLASS_SECTION"]
		course.PeoplesoftNumber, err = strconv.Atoi(unparsed["CLASS_NBR"])
		assertError(err)

		// Options for CONSENT are 'N', ' ', 'D
		course.Consent = unparsed["CONSENT"]

		// Options for GRADING_BASIS are OPT,GRD,OPX,OPP, ,WPP,PF4,NON,PNP,XEG,PF5
		course.GradingBasis = unparsed["GRADING_BASIS"]
		passFail := false
		fifthCourse := false
		switch course.GradingBasis {
		case "WPP":
			course.GradingBasis = "W"
		case "GRD":
			course.GradingBasis = "No Pass/Fail and No Fifth Course"
		case "OPT":
			course.GradingBasis = "Pass/Fail Available, Fifth Course Available"
			passFail = true
			fifthCourse = true
		case "OPX":
			course.GradingBasis = "Pass/Fail Unavailable, Fifth Course Available"
			fifthCourse = true
		case "OPP":
			course.GradingBasis = "Pass/Fail Available, Fifth Course Unavailable"
			passFail = true
		default:
			course.GradingBasis = ""
		}

		ssrComponent := unparsed["SSR_COMPONENT"]
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

		course.TitleLong = unparsed["COURSE_TITLE_LONG"]
		course.TitleShort = unparsed["DESCR"]

		course.Instructors = []Instructor{}

		for i := 1; i <= 6; i++ {
			instructor := Instructor{}

			fn := unparsed["WMS_FIRST_NAME"+strconv.Itoa(i)]
			mn := unparsed["WMS_MID_NAME"+strconv.Itoa(i)]
			ln := unparsed["WMS_LAST_NAME"+strconv.Itoa(i)]

			if fn == " " {
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

		for i := 1; i <= 3; i++ {
			meeting := Meeting{}

			// Different options: MW, ,TR,MWF,W,TF,TBA,MR,T,M,R,M-F,F
			days := unparsed["WMS_STND_MTG_PAT"+strconv.Itoa(i)]
			if days == " " {
				continue
			} else if days == "TBA" {
				break
			}

			meeting.Days = days

			const twentyFourHour = "15:04"
			const twelveHour = "3:04pm"

			startT := unparsed["WMS_START_TIME"+strconv.Itoa(i)]
			if startT == " " {
				meeting.Start = ""
				meeting.Start12 = ""
			} else {
				startTime, err := time.Parse(twentyFourHour, startT)
				assertError(err)

				meeting.Start = startTime.Format(twentyFourHour)
				meeting.Start12 = startTime.Format(twelveHour)
			}

			endT := unparsed["WMS_END_TIME"+strconv.Itoa(i)]
			if endT == " " {
				meeting.End = ""
				meeting.End12 = ""
			} else {
				endTime, err := time.Parse(twentyFourHour, endT)
				assertError(err)

				meeting.End = endTime.Format(twentyFourHour)
				meeting.End12 = endTime.Format(twelveHour)
			}

			meeting.Facil = unparsed["WMS_FACIL_DESCR"+strconv.Itoa(i)]
			course.Meetings = append(course.Meetings, meeting)
		}

		unparsedAttributes := unparsed["WMS_ATTR_SRCH"]
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

		course.ClassFormat = unparsed["WMS_CLASS_FORMAT"]
		if unparsed["WMS_RQMT_EVAL"] == " " {
			course.ClassReqEval = ""
		} else {
			course.ClassReqEval = unparsed["WMS_RQMT_EVAL"]
		}

		course.ExtraInfo = unparsed["WMS_EXTRA_INFO"]
		if unparsed["WMS_EXTRA_INFO2"] != " " {
			course.ExtraInfo += "; " + unparsed["WMS_EXTRA_INFO2"]
		}

		course.Prereqs = unparsed["WMS_PREREQS"]
		course.DepartmentNotes = unparsed["WMS_DEPT_NOTES"]

		course.DescriptionSearch = unparsed["WMS_DESCR_SRCH"]
		course.EnrlPref = unparsed["WMS_ENRL_PREF"]

		courses = append(courses, course)
	}

	return courses
}

func grabCatalog() []byte {

	url := "https://catalog.williams.edu/wp-json/courses/v1/year/1920"

	catalogClient := http.Client{
		Timeout: time.Second * 30, // Maximum of 30 seconds
	}

	// Craft a GET request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assertError(err)

	// Send the GET request and get back the response
	res, getErr := catalogClient.Do(req)
	assertError(getErr)

	// Parse the body into []byte
	body, readErr := ioutil.ReadAll(res.Body)
	assertError(readErr)

	return body
}

func exportConstants() {
	// Most of these are hardcoded... Wonder if it makes more sense to edit the .json directly
	// rather than use this function
	constants := map[string]interface{}{
		"Palette":       [...]string{"#B3E5FC", "#F0F4C3", "#FFCCBC", "#CFD8DC", "#E1BEE7", "#B2DFDB"},
		"BorderPalette": [...]string{"#03A9F4", "#AFB42B", "#E64A19", "#455A64", "#7B17A2", "#00796B"},
		"Distributions": [...]string{"DPE", "QFR", "WAC"},
		"Divisions":     [...]string{"DIV1", "DIV2", "DIV3"},
		"Others":        [...]string{"OPP", "OPX"},
		"Levels":        [...]int{0, 1, 2, 3, 4},
		"ClassTypes":    [...]string{"Lecture", "Seminar", "Tutorial", "Studio", "Independent Study", "Laboratory"},
		"UpdateDate":    time.Now().Format(time.RFC850),
		"Dates": map[string]map[string]string{
			"Fall": map[string]string{
				"StartGCAL": "2019-09-05T",
				"Start":     "20190905",
				"End":       "20191207",
			},
			"Winter": map[string]string{
				"StartGCAL": "2020-01-06T",
				"Start":     "20200106",
				"End":       "20200130",
			},
			"Spring": map[string]string{
				"StartGCAL": "2020-02-05T",
				"Start":     "20200205",
				"End":       "20200515",
			},
		},
		"StartTimes": [...]string{"01:10", "08:00", "08:30", "08:55", "09:00", "09:55", "10:00", "11:00",
			"11:20", "12:00", "13:00", "13:10", "14:00", "14:10", "14:30", "14:35", "16:45", "18:30", "19:00"},
		"StartTimes12": [...]string{" 1:10AM", " 8:00AM", " 8:30AM", " 8:55AM", " 9:00AM", " 9:55AM", "10:00AM",
			"11:00AM", "11:20AM", "12:00PM", " 1:00PM", " 1:10PM", " 2:00PM", " 2:10PM", " 2:30PM", " 2:35PM", " 4:45PM",
			" 6:30PM", " 7:00PM"},
		"EndTimes": [...]string{"02:25", "08:50", "09:45", "09:50", "10:50", "11:10", "11:40", "11:50",
			"12:00", "12:15", "12:35", "12:40", "12:50", "13:45", "14:00", "14:25", "14:30", "14:35", "15:00",
			"15:40", "15:50", "16:00", "16:40", "17:00", "20:15", "20:30", "21:40", "22:00"},
		"EndTimes12": [...]string{" 2:25AM", " 8:50AM", " 9:45AM", " 9:50AM", "10:50AM", "11:10AM", "11:40AM",
			"11:50AM", "12:00PM", "12:15PM", "12:35PM", "12:40PM", "12:50PM", " 1:45PM", " 2:00PM", " 2:25PM",
			" 2:30PM", " 2:35PM", " 3:00PM", " 3:40PM", " 3:50PM", " 4:00PM", " 4:40PM", " 5:00PM", " 8:15PM",
			" 8:30PM", " 9:40PM", "10:00PM", "11:50PM"},
	}

	constantsJSON, err := json.Marshal(constants)
	assertError(err)
	err = ioutil.WriteFile("constants.json", constantsJSON, 0644)
	assertError(err)
}

func exportCatalog(courses []Course) {
	coursesJSON, err := json.Marshal(courses)
	assertError(err)
	err = ioutil.WriteFile("courses.json", coursesJSON, 0644)
	assertError(err)
}

func updateCatalog() {
	responseBody := grabCatalog()
	courses := ParseCatalog(responseBody)
	exportCatalog(courses)
	exportConstants()
}
