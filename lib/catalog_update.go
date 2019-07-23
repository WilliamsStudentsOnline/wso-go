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

type name struct {
	firstName  string
	middleName string
	lastName   string
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

type uMeeting struct {
	days     string
	start    string
	end      string
	facility string
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

type uCourse struct {
	WMSACADYEAR     string `json:"WMS_ACAD_YEAR"`
	OFFERED         string `json:"OFFERED"`
	STRM            string `json:"STRM"`
	CRSEID          string `json:"CRSE_ID"`
	EFFDT           string `json:"EFFDT"`
	SUBJECT         string `json:"SUBJECT"`
	CATALOGNBR      string `json:"CATALOG_NBR"`
	WMSCRSELETTER   string `json:"WMS_CRSE_LETTER"`
	CLASSSECTION    string `json:"CLASS_SECTION"`
	CLASSNBR        string `json:"CLASS_NBR"`
	CONSENT         string `json:"CONSENT"`
	GRADINGBASIS    string `json:"GRADING_BASIS"`
	SSRCOMPONENT    string `json:"SSR_COMPONENT"`
	DESCR           string `json:"DESCR"`
	UNITSMINIMUM    string `json:"UNITS_MINIMUM"`
	COURSETITLELONG string `json:"COURSE_TITLE_LONG"`
	WMSFIRSTNAME1   string `json:"WMS_FIRST_NAME1"`
	WMSMIDNAME1     string `json:"WMS_MID_NAME1"`
	WMSLASTNAME1    string `json:"WMS_LAST_NAME1"`
	URL1            string `json:"URL_1"`
	WMSFIRSTNAME2   string `json:"WMS_FIRST_NAME2"`
	WMSMIDNAME2     string `json:"WMS_MID_NAME2"`
	WMSLASTNAME2    string `json:"WMS_LAST_NAME2"`
	URL2            string `json:"URL_2"`
	WMSFIRSTNAME3   string `json:"WMS_FIRST_NAME3"`
	WMSMIDNAME3     string `json:"WMS_MID_NAME3"`
	WMSLASTNAME3    string `json:"WMS_LAST_NAME3"`
	URL3            string `json:"URL_3"`
	WMSFIRSTNAME4   string `json:"WMS_FIRST_NAME4"`
	WMSMIDNAME4     string `json:"WMS_MID_NAME4"`
	WMSLASTNAME4    string `json:"WMS_LAST_NAME4"`
	URL4            string `json:"URL_4"`
	WMSFIRSTNAME5   string `json:"WMS_FIRST_NAME5"`
	WMSMIDNAME5     string `json:"WMS_MID_NAME5"`
	WMSLASTNAME5    string `json:"WMS_LAST_NAME5"`
	URL5            string `json:"URL_5"`
	WMSFIRSTNAME6   string `json:"WMS_FIRST_NAME6"`
	WMSMIDNAME6     string `json:"WMS_MID_NAME6"`
	WMSLASTNAME6    string `json:"WMS_LAST_NAME6"`
	URL6            string `json:"URL_6"`
	WMSSTNDMTGPAT1  string `json:"WMS_STND_MTG_PAT1"`
	WMSSTARTTIME1   string `json:"WMS_START_TIME1"`
	WMSENDTIME1     string `json:"WMS_END_TIME1"`
	WMSFACILDESCR1  string `json:"WMS_FACIL_DESCR1"`
	WMSSTNDMTGPAT2  string `json:"WMS_STND_MTG_PAT2"`
	WMSSTARTTIME2   string `json:"WMS_START_TIME2"`
	WMSENDTIME2     string `json:"WMS_END_TIME2"`
	WMSFACILDESCR2  string `json:"WMS_FACIL_DESCR2"`
	WMSSTNDMTGPAT3  string `json:"WMS_STND_MTG_PAT3"`
	WMSSTARTTIME3   string `json:"WMS_START_TIME3"`
	WMSENDTIME3     string `json:"WMS_END_TIME3"`
	WMSFACILDESCR3  string `json:"WMS_FACIL_DESCR3"`
	WMSATTRSRCH     string `json:"WMS_ATTR_SRCH"`
	WMSCLASSFORMAT  string `json:"WMS_CLASS_FORMAT"`
	WMSRQMTEVAL     string `json:"WMS_RQMT_EVAL"`
	WMSEXTRAINFO    string `json:"WMS_EXTRA_INFO"`
	WMSEXTRAINFO2   string `json:"WMS_EXTRA_INFO2"`
	WMSINSTROTH     string `json:"WMS_INSTR_OTH"`
	WMSPREREQS      string `json:"WMS_PREREQS"`
	WMSENRLPREF     string `json:"WMS_ENRL_PREF"`
	WMSDEPTNOTES    string `json:"WMS_DEPT_NOTES"`
	WMSMATLFEE      string `json:"WMS_MATL_FEE"`
	WMSEXPENRL      string `json:"WMS_EXP_ENRL"`
	WMSENRLLIMIT    string `json:"WMS_ENRL_LIMIT"`
	WMSNC           string `json:"WMS_NC"`
	CAMPUS          string `json:"CAMPUS"`
	WMSDESCR140     string `json:"WMS_DESCR140"`
	WMSSHORTDESCR   string `json:"WMS_SHORT_DESCR"`
	WMSDISTRIBNT1   string `json:"WMS_DISTRIB_NT1"`
	WMSDISTRIBNT2   string `json:"WMS_DISTRIB_NT2"`
	WMSDISTRIBNT3   string `json:"WMS_DISTRIB_NT3"`
	WMSDESCRSRCH    string `json:"WMS_DESCR_SRCH"`
	WMSDISTRIBNOTES string `json:"WMS_DISTRIB_NOTES"`
}

type exportCourses struct {
	Courses    []Course `json:"courses"`
	UpdateTime string   `json:"updateTime"`
}

const (
	fallSemesterID   = 1201
	winterSemesterID = 1202
	springSemesterID = 1203
)

func ParseCatalog(catalog []byte) ([]Course, error) {
	var unparsedCourses = []uCourse{}
	err := json.Unmarshal(catalog, &unparsedCourses)
	if err != nil {
		return nil, err
	}

	courses := []Course{}

	for _, unparsed := range unparsedCourses {
		if unparsed.OFFERED != "Y" || unparsed.WMSFACILDESCR1 == "Cancelled" {
			continue
		}
		course := Course{}

		course.Year, err = strconv.Atoi(unparsed.WMSACADYEAR)
		if err != nil {
			return nil, err
		}
		semID, err := strconv.Atoi(unparsed.STRM)
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

		course.CourseID, err = strconv.Atoi(unparsed.CRSEID)
		if err != nil {
			return nil, err
		}
		course.Department = unparsed.SUBJECT
		course.Number, err = strconv.Atoi(unparsed.CATALOGNBR)
		if err != nil {
			return nil, err
		}

		// Tutorial sections start with 'T'
		course.Section = unparsed.CLASSSECTION
		course.PeoplesoftNumber, err = strconv.Atoi(unparsed.CLASSNBR)
		if err != nil {
			return nil, err
		}

		// Options for CONSENT are 'N', ' ', 'D
		course.Consent = unparsed.CONSENT

		// Options for GRADING_BASIS are OPT,GRD,OPX,OPP, ,WPP,PF4,NON,PNP,XEG,PF5
		course.GradingBasis = unparsed.GRADINGBASIS
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

		ssrComponent := unparsed.SSRCOMPONENT
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

		course.TitleLong = unparsed.COURSETITLELONG
		course.TitleShort = unparsed.DESCR

		course.Instructors = []Instructor{}

		names := []name{
			{
				unparsed.WMSFIRSTNAME1, unparsed.WMSMIDNAME1, unparsed.WMSLASTNAME1,
			},
			{
				unparsed.WMSFIRSTNAME2, unparsed.WMSMIDNAME2, unparsed.WMSLASTNAME2,
			},
			{
				unparsed.WMSFIRSTNAME3, unparsed.WMSMIDNAME3, unparsed.WMSLASTNAME3,
			},
			{
				unparsed.WMSFIRSTNAME4, unparsed.WMSMIDNAME4, unparsed.WMSLASTNAME4,
			},
			{
				unparsed.WMSFIRSTNAME5, unparsed.WMSMIDNAME5, unparsed.WMSLASTNAME5,
			},
			{
				unparsed.WMSFIRSTNAME6, unparsed.WMSMIDNAME6, unparsed.WMSLASTNAME6,
			},
		}

		for _, instructorName := range names {
			instructor := Instructor{}

			fn := instructorName.firstName
			mn := instructorName.middleName
			ln := instructorName.lastName

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
		unparsedMeetings := []uMeeting{
			{
				unparsed.WMSSTNDMTGPAT1, unparsed.WMSSTARTTIME1, unparsed.WMSENDTIME1, unparsed.WMSFACILDESCR1,
			},
			{
				unparsed.WMSSTNDMTGPAT2, unparsed.WMSSTARTTIME2, unparsed.WMSENDTIME2, unparsed.WMSFACILDESCR2,
			},
			{
				unparsed.WMSSTNDMTGPAT3, unparsed.WMSSTARTTIME3, unparsed.WMSENDTIME3, unparsed.WMSFACILDESCR3,
			},
		}

		for _, unparsedMeeting := range unparsedMeetings {
			meeting := Meeting{}

			// Different options: MW, ,TR,MWF,W,TF,TBA,MR,T,M,R,M-F,F
			days := unparsedMeeting.days
			if days == " " {
				continue
			} else if days == "TBA" {
				break
			}

			meeting.Days = days

			const twentyFourHour = "15:04"
			const twelveHour = "3:04pm"

			startT := unparsedMeeting.start
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

			endT := unparsedMeeting.end
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

			meeting.Facil = unparsedMeeting.facility
			course.Meetings = append(course.Meetings, meeting)
		}

		unparsedAttributes := unparsed.WMSATTRSRCH
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

		course.ClassFormat = unparsed.WMSCLASSFORMAT
		if unparsed.WMSRQMTEVAL == " " {
			course.ClassReqEval = ""
		} else {
			course.ClassReqEval = unparsed.WMSRQMTEVAL
		}

		course.ExtraInfo = unparsed.WMSEXTRAINFO
		if unparsed.WMSEXTRAINFO2 != " " {
			course.ExtraInfo += "; " + unparsed.WMSEXTRAINFO2
		}

		course.Prereqs = unparsed.WMSPREREQS
		course.DepartmentNotes = unparsed.WMSDEPTNOTES

		course.DescriptionSearch = unparsed.WMSDESCRSRCH
		course.EnrlPref = unparsed.WMSENRLPREF

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
