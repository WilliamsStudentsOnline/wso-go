package catalog_update_test

import (
	"encoding/json"
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	testify "github.com/stretchr/testify/assert"
)

const (
	fallSemesterID   = 1201
	winterSemesterID = 1202
	springSemesterID = 1203
)

func TestParse(t *testing.T) {

	assertParse := func(catalog string, expected string, t *testing.T) {
		assert := testify.New(t)

		var rawCourses []RawCourse
		err := json.Unmarshal([]byte(catalog), &rawCourses)
		assert.NoError(err)

		courses, err := ParseCatalog(rawCourses, fallSemesterID, winterSemesterID, springSemesterID)
		assert.NoError(err)
		b, err := json.Marshal(courses)
		assert.NoError(err)

		assert.JSONEq(expected, string(b))

	}

	t.Run("Parses an available course properly", func(t *testing.T) {
		catalog := `[
			{
			  "WMS_ACAD_YEAR": "2020",
			  "OFFERED": "Y",
			  "STRM": "1201",
			  "CRSE_ID": "020209",
			  "EFFDT": "01-SEP-19",
			  "SUBJECT": "AFR",
			  "CATALOG_NBR": "105",
			  "WMS_CRSE_LETTER": " ",
			  "CLASS_SECTION": "01",
			  "CLASS_NBR": "1089",
			  "CONSENT": "N",
			  "GRADING_BASIS": "OPT",
			  "SSR_COMPONENT": "LEC",
			  "DESCR": "African Art Survey",
			  "UNITS_MINIMUM": "1",
			  "COURSE_TITLE_LONG": "Materials, Meanings, and Messages in the Arts of Africa",
			  "WMS_FIRST_NAME1": "Michelle",
			  "WMS_MID_NAME1": "M.",
			  "WMS_LAST_NAME1": "Apotsos",
			  "URL_1": "https://art.williams.edu/profile/ma11/",
			  "WMS_FIRST_NAME2": " ",
			  "WMS_MID_NAME2": " ",
			  "WMS_LAST_NAME2": " ",
			  "URL_2": " ",
			  "WMS_FIRST_NAME3": " ",
			  "WMS_MID_NAME3": " ",
			  "WMS_LAST_NAME3": " ",
			  "URL_3": " ",
			  "WMS_FIRST_NAME4": " ",
			  "WMS_MID_NAME4": " ",
			  "WMS_LAST_NAME4": " ",
			  "URL_4": " ",
			  "WMS_FIRST_NAME5": " ",
			  "WMS_MID_NAME5": " ",
			  "WMS_LAST_NAME5": " ",
			  "URL_5": " ",
			  "WMS_FIRST_NAME6": " ",
			  "WMS_MID_NAME6": " ",
			  "WMS_LAST_NAME6": " ",
			  "URL_6": " ",
			  "WMS_STND_MTG_PAT1": "MW",
			  "WMS_START_TIME1": "11:00",
			  "WMS_END_TIME1": "12:15",
			  "WMS_FACIL_DESCR1": " ",
			  "WMS_STND_MTG_PAT2": " ",
			  "WMS_START_TIME2": " ",
			  "WMS_END_TIME2": " ",
			  "WMS_FACIL_DESCR2": " ",
			  "WMS_STND_MTG_PAT3": " ",
			  "WMS_START_TIME3": " ",
			  "WMS_END_TIME3": " ",
			  "WMS_FACIL_DESCR3": " ",
			  "WMS_ATTR_SRCH": "DIV_D2,DPE_DPE",
			  "WMS_CLASS_FORMAT": "lecture",
			  "WMS_RQMT_EVAL": "three 2-page response papers, class journal on WCMA objects, finals",
			  "WMS_EXTRA_INFO": " ",
			  "WMS_EXTRA_INFO2": " ",
			  "WMS_INSTR_OTH": " ",
			  "WMS_PREREQS": "none",
			  "WMS_ENRL_PREF": "Art History and African Studies majors",
			  "WMS_DEPT_NOTES": " ",
			  "WMS_MATL_FEE": " ",
			  "WMS_EXP_ENRL": "40",
			  "WMS_ENRL_LIMIT": "40",
			  "WMS_NC": " ",
			  "CAMPUS": "WMS",
			  "WMS_DESCR140": " ",
			  "WMS_SHORT_DESCR": null,
			  "WMS_DISTRIB_NT1": null,
			  "WMS_DISTRIB_NT2": "Distribution Notes.",
			  "WMS_DISTRIB_NT3": null,
			  "WMS_DESCR_SRCH": "Lorem Ipsum.",
			  "WMS_DISTRIB_NOTES": null
			}
		  ]`

		expected := `[{"year":2020,"semester":"FALL","courseID":"020209","department":"AFR","number":105,"section":"01","peoplesoftNumber":1089,"consent":"N","gradingBasis":"OPT","gradingBasisDesc":"Pass/Fail Available, Fifth Course Available","classType":"Lecture","titleLong":"Materials, Meanings, And Messages In The Arts Of Africa","titleShort":"African Art Survey","instructors":[{"url":"","name":"Michelle M. Apotsos"}],"meetings":[{"days":"MW","start":"11:00","end":"12:15","facility":""}],"courseAttributes":{"div1":false,"div2":true,"div3":false,"dpe":true,"qfr":false,"wac":false,"passFail":true,"fifthCourse":true},"classFormat":"Lecture","classReqEval":"Three 2-page response papers, class journal on WCMA objects, finals","extraInfo":"","prereqs":"None","departmentNotes":"","descriptionSearch":"Lorem Ipsum.","enrolmentPreferences":"Art History and African Studies majors"}]`
		assertParse(catalog, expected, t)
	})

	t.Run("Ignores an unoffered course", func(t *testing.T) {
		catalog := `[
			{
			  "WMS_ACAD_YEAR": "2020",
			  "OFFERED": "N",
			  "STRM": "1193",
			  "CRSE_ID": "019563",
			  "EFFDT": "01-SEP-19",
			  "SUBJECT": "AFR",
			  "CATALOG_NBR": "113",
			  "WMS_CRSE_LETTER": " ",
			  "CLASS_SECTION": " ",
			  "CLASS_NBR": "3820",
			  "CONSENT": "N",
			  "GRADING_BASIS": "OPT",
			  "SSR_COMPONENT": "LEC",
			  "DESCR": "Musics of Africa",
			  "UNITS_MINIMUM": "1",
			  "COURSE_TitleLong": "Musics of Africa",
			  "WMS_FIRST_NAME1": "Corinna",
			  "WMS_MID_NAME1": "S.",
			  "WMS_LAST_NAME1": "Campbell",
			  "URL_1": "https://music.williams.edu/profile/csc3",
			  "WMS_FIRST_NAME2": " ",
			  "WMS_MID_NAME2": " ",
			  "WMS_LAST_NAME2": " ",
			  "URL_2": " ",
			  "WMS_FIRST_NAME3": " ",
			  "WMS_MID_NAME3": " ",
			  "WMS_LAST_NAME3": " ",
			  "URL_3": " ",
			  "WMS_FIRST_NAME4": " ",
			  "WMS_MID_NAME4": " ",
			  "WMS_LAST_NAME4": " ",
			  "URL_4": " ",
			  "WMS_FIRST_NAME5": " ",
			  "WMS_MID_NAME5": " ",
			  "WMS_LAST_NAME5": " ",
			  "URL_5": " ",
			  "WMS_FIRST_NAME6": " ",
			  "WMS_MID_NAME6": " ",
			  "WMS_LAST_NAME6": " ",
			  "URL_6": " ",
			  "WMS_STND_MTG_PAT1": " ",
			  "WMS_START_TIME1": " ",
			  "WMS_END_TIME1": " ",
			  "WMS_FACIL_DESCR1": " ",
			  "WMS_STND_MTG_PAT2": " ",
			  "WMS_START_TIME2": " ",
			  "WMS_END_TIME2": " ",
			  "WMS_FACIL_DESCR2": " ",
			  "WMS_STND_MTG_PAT3": " ",
			  "WMS_START_TIME3": " ",
			  "WMS_END_TIME3": " ",
			  "WMS_FACIL_DESCR3": " ",
			  "WMS_ATTR_SRCH": "DIV_D1,GBST_GBSTAFR,MUS_MUSW",
			  "WMS_CLASS_FORMAT": "lecture/discussion",
			  "WMS_RQMT_EVAL": "grade based on a listening journal, bi-weekly short assignments",
			  "WMS_EXTRA_INFO": " ",
			  "WMS_EXTRA_INFO2": " ",
			  "WMS_INSTR_OTH": " ",
			  "WMS_PREREQS": "no prerequisites: prior musical background is not essential for this class",
			  "WMS_ENRL_PREF": "current or prospective Music majors and Africana Studies concentrators",
			  "WMS_DEPT_NOTES": " ",
			  "WMS_MATL_FEE": " ",
			  "WMS_EXP_ENRL": "12",
			  "WMS_ENRL_LIMIT": "20",
			  "WMS_NC": " ",
			  "CAMPUS": " ",
			  "WMS_DESCR140": " ",
			  "WMS_SHORT_DESCR": null,
			  "WMS_DISTRIB_NT1": null,
			  "WMS_DISTRIB_NT2": null,
			  "WMS_DISTRIB_NT3": null,
			  "WMS_DESCR_SRCH": "Lorem Ipsum dolor",
			  "WMS_DISTRIB_NOTES": null
			}
		  ]`
		expected := `[]`

		assertParse(catalog, expected, t)
	})

	t.Run("Ignores a cancelled course", func(t *testing.T) {
		catalog := `[
			{
			  "WMS_ACAD_YEAR": "2020",
			  "OFFERED": "Y",
			  "STRM": "1201",
			  "CRSE_ID": "020209",
			  "EFFDT": "01-SEP-19",
			  "SUBJECT": "AFR",
			  "CATALOG_NBR": "105",
			  "WMS_CRSE_LETTER": " ",
			  "CLASS_SECTION": "01",
			  "CLASS_NBR": "1089",
			  "CONSENT": "N",
			  "GRADING_BASIS": "OPT",
			  "SSR_COMPONENT": "LEC",
			  "DESCR": "African Art Survey",
			  "UNITS_MINIMUM": "1",
			  "COURSE_TITLE_LONG": "Materials, Meanings, and Messages in the Arts of Africa",
			  "WMS_FIRST_NAME1": "Michelle",
			  "WMS_MID_NAME1": "M.",
			  "WMS_LAST_NAME1": "Apotsos",
			  "URL_1": "https://art.williams.edu/profile/ma11/",
			  "WMS_FIRST_NAME2": " ",
			  "WMS_MID_NAME2": " ",
			  "WMS_LAST_NAME2": " ",
			  "URL_2": " ",
			  "WMS_FIRST_NAME3": " ",
			  "WMS_MID_NAME3": " ",
			  "WMS_LAST_NAME3": " ",
			  "URL_3": " ",
			  "WMS_FIRST_NAME4": " ",
			  "WMS_MID_NAME4": " ",
			  "WMS_LAST_NAME4": " ",
			  "URL_4": " ",
			  "WMS_FIRST_NAME5": " ",
			  "WMS_MID_NAME5": " ",
			  "WMS_LAST_NAME5": " ",
			  "URL_5": " ",
			  "WMS_FIRST_NAME6": " ",
			  "WMS_MID_NAME6": " ",
			  "WMS_LAST_NAME6": " ",
			  "URL_6": " ",
			  "WMS_STND_MTG_PAT1": " ",
			  "WMS_START_TIME1": " ",
			  "WMS_END_TIME1": " ",
			  "WMS_FACIL_DESCR1": "Cancelled",
			  "WMS_STND_MTG_PAT2": " ",
			  "WMS_START_TIME2": " ",
			  "WMS_END_TIME2": " ",
			  "WMS_FACIL_DESCR2": " ",
			  "WMS_STND_MTG_PAT3": " ",
			  "WMS_START_TIME3": " ",
			  "WMS_END_TIME3": " ",
			  "WMS_FACIL_DESCR3": " ",
			  "WMS_ATTR_SRCH": "DIV_D2,DPE_DPE",
			  "WMS_CLASS_FORMAT": "lecture",
			  "WMS_RQMT_EVAL": "three 2-page response papers, class journal on WCMA objects, finals",
			  "WMS_EXTRA_INFO": " ",
			  "WMS_EXTRA_INFO2": " ",
			  "WMS_INSTR_OTH": " ",
			  "WMS_PREREQS": "none",
			  "WMS_ENRL_PREF": "Art History and African Studies majors",
			  "WMS_DEPT_NOTES": " ",
			  "WMS_MATL_FEE": " ",
			  "WMS_EXP_ENRL": "40",
			  "WMS_ENRL_LIMIT": "40",
			  "WMS_NC": " ",
			  "CAMPUS": "WMS",
			  "WMS_DESCR140": " ",
			  "WMS_SHORT_DESCR": null,
			  "WMS_DISTRIB_NT1": null,
			  "WMS_DISTRIB_NT2": "Distrib Notes2",
			  "WMS_DISTRIB_NT3": null,
			  "WMS_DESCR_SRCH": "Lorem Ipsum dolor amit",
			  "WMS_DISTRIB_NOTES": null
			}
		  ]`
		expected := `[]`

		assertParse(catalog, expected, t)
	})

	// t.Run("Fetches data from online catalog", func(t *testing.T) {
	// 	updateCatalog()
	// })
}
