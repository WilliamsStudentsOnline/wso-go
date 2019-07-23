package lib

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Instructor struct {
	URL  string
	Name string `json:"name"`
}

type Meeting struct {
	Days     string `json:"days"`
	Start    string `json:"start"`
	Start12  string `json:"start12"`
	End      string `json:"end"`
	End12    string `json:"end12"`
	FacDescr string `json:"facDescr"`
	Facil    string `json:"facil"`
}

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

type Course struct {
	Year              int          `json:"year"`
	Semester          string       `json:"semester"`
	CourseID          int          `json:"courseID"`
	Department        string       `json:"departmnet"`
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

type exportCourses struct {
	Courses    []Course `json:"courses"`
	UpdateTime string   `json:"updateTime"`
}

type unparsedJSON []map[string]string

const (
	fallSemesterID   = 1201
	winterSemesterID = 1202
	springSemesterID = 1203
)

func ParseCatalog(catalog []byte) ([]Course, error) {
	unparsedCourses := unparsedJSON{}
	err := json.Unmarshal(catalog, &unparsedCourses)
	if err != nil {
		return nil, err
	}

	courses := []Course{}

	for _, unparsed := range unparsedCourses {
		if unparsed["OFFERED"] != "Y" || unparsed["WMS_FACIL_DESCR1"] == "Cancelled" {
			continue
		}
		course := Course{}

		course.Year, err = strconv.Atoi(unparsed["WMS_ACAD_YEAR"])
		if err != nil {
			return nil, err
		}
		semID, err := strconv.Atoi(unparsed["STRM"])
		if err != nil {
			return nil, err
		}

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

		course.CourseID, err = strconv.Atoi(unparsed["CRSE_ID"])
		if err != nil {
			return nil, err
		}
		course.Department = unparsed["SUBJECT"]
		course.Number, err = strconv.Atoi(unparsed["CATALOG_NBR"])
		if err != nil {
			return nil, err
		}

		// Tutorial sections start with 'T'
		course.Section = unparsed["CLASS_SECTION"]
		course.PeoplesoftNumber, err = strconv.Atoi(unparsed["CLASS_NBR"])
		if err != nil {
			return nil, err
		}

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
				if err != nil {
					return nil, err
				}

				meeting.Start = startTime.Format(twentyFourHour)
				meeting.Start12 = startTime.Format(twelveHour)
			}

			endT := unparsed["WMS_END_TIME"+strconv.Itoa(i)]
			if endT == " " {
				meeting.End = ""
				meeting.End12 = ""
			} else {
				endTime, err := time.Parse(twentyFourHour, endT)
				if err != nil {
					return nil, err
				}

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

	return courses, nil
}

func grabCatalog() ([]byte, error) {

	url := "https://catalog.williams.edu/wp-json/courses/v1/year/1920"

	catalogClient := http.Client{
		Timeout: time.Second * 30, // Maximum of 30 seconds
	}

	// Craft a GET request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Send the GET request and get back the response
	res, err := catalogClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse the body into []byte
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
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
	exportCatalog(courses)

	return nil
}
