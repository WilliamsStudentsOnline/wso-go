package lib

import (
	"encoding/json"
	"fmt"
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

type UnparsedJSON []map[string]interface{}

const FallSemesterID = 1201
const WinterSemesterID = 1202
const SpringSemesterID = 1203

func ParseCatalog(catalog string) []Course {
	textBytes := []byte(catalog)

	unparsedCourses := UnparsedJSON{}
	err := json.Unmarshal(textBytes, &unparsedCourses)
	if err != nil {
		fmt.Println(err)
	}

	courses := []Course{}

	for _, unparsed := range unparsedCourses {
		if unparsed["OFFERED"] != "Y" || unparsed["WMS_FACIL_DESCR1"] == "Cancelled" {
			continue
		}
		course := Course{}

		course.Year, err = strconv.Atoi(unparsed["WMS_ACAD_YEAR"].(string))

		semID, err := strconv.Atoi(unparsed["STRM"].(string))
		if err != nil {
			fmt.Println(err)
		}
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

		course.CourseID, err = strconv.Atoi(unparsed["CRSE_ID"].(string))
		course.Department = unparsed["SUBJECT"].(string)
		course.Number, err = strconv.Atoi(unparsed["CATALOG_NBR"].(string))

		// Tutorial sections start with 'T'
		course.Section = unparsed["CLASS_SECTION"].(string)
		course.PeoplesoftNumber, err = strconv.Atoi(unparsed["CLASS_NBR"].(string))

		// Options for CONSENT are 'N', ' ', 'D
		course.Consent = unparsed["CONSENT"].(string)

		// Options for GRADING_BASIS are OPT,GRD,OPX,OPP, ,WPP,PF4,NON,PNP,XEG,PF5
		course.GradingBasis = unparsed["GRADING_BASIS"].(string)
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

		ssrComponent := unparsed["SSR_COMPONENT"].(string)
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

		course.TitleLong = fmt.Sprintf("%v", unparsed["COURSE_TITLE_LONG"])
		course.TitleShort = unparsed["DESCR"].(string)

		course.Instructors = []Instructor{}

		for i := 1; i <= 6; i++ {
			instructor := Instructor{}

			fn := fmt.Sprintf("%v", unparsed["WMS_FIRST_NAME"+strconv.Itoa(i)])
			mn := fmt.Sprintf("%v", unparsed["WMS_MID_NAME"+strconv.Itoa(i)])
			ln := fmt.Sprintf("%v", unparsed["WMS_LAST_NAME"+strconv.Itoa(i)])

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
			days := fmt.Sprintf("%v", unparsed["WMS_STND_MTG_PAT"+strconv.Itoa(i)])
			if days == " " {
				continue
			} else if days == "TBA" {
				break
			}

			meeting.Days = days

			const twentyFourHour = "15:04"
			const twelveHour = "3:04pm"

			startT := fmt.Sprintf("%v", unparsed["WMS_START_TIME"+strconv.Itoa(i)])
			if startT == " " {
				meeting.Start = ""
				meeting.Start12 = ""
			} else {
				startTime, err := time.Parse(twentyFourHour, startT)
				if err != nil {
					fmt.Println(err)
				}

				meeting.Start = startTime.Format(twentyFourHour)
				meeting.Start12 = startTime.Format(twelveHour)
			}

			endT := fmt.Sprintf("%v", unparsed["WMS_END_TIME"+strconv.Itoa(i)])
			if endT == " " {
				meeting.End = ""
				meeting.End12 = ""
			} else {
				endTime, err := time.Parse(twentyFourHour, endT)
				if err != nil {
					fmt.Println(err)
				}

				meeting.End = endTime.Format(twentyFourHour)
				meeting.End12 = endTime.Format(twelveHour)
			}

			meeting.Facil = fmt.Sprintf("%v", unparsed["WMS_FACIL_DESCR"+strconv.Itoa(i)])
			course.Meetings = append(course.Meetings, meeting)
		}

		unparsedAttributes := unparsed["WMS_ATTR_SRCH"].(string)
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

		course.ClassFormat = unparsed["WMS_CLASS_FORMAT"].(string)
		if unparsed["WMS_RQMT_EVAL"].(string) == " " {
			course.ClassReqEval = ""
		} else {
			course.ClassReqEval = unparsed["WMS_RQMT_EVAL"].(string)
		}

		course.ExtraInfo = unparsed["WMS_EXTRA_INFO"].(string)
		if unparsed["WMS_EXTRA_INFO2"] != " " {
			course.ExtraInfo += "; " + unparsed["WMS_EXTRA_INFO2"].(string)
		}

		course.Prereqs = unparsed["WMS_PREREQS"].(string)
		course.DepartmentNotes = unparsed["WMS_DEPT_NOTES"].(string)

		course.DescriptionSearch = unparsed["WMS_DESCR_SRCH"].(string)
		course.EnrlPref = unparsed["WMS_ENRL_PREF"].(string)

		courses = append(courses, course)
	}

	return courses
}
