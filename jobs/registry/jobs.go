package registry

const defaultConfig = "/opt/wso-go/static/wso-backend-config.yaml"

var needsMariaDB = []string{"mariadb.service"}

func init() {
	mustRegister(Job{
		Name:        "user-home-info",
		Description: "ingest registrar WSO TSV (home address, SU box, entry)",
		Binary:      InstallDir + "/job-user-home-info_linux",
		Params: []Param{
			{Name: "file", Type: ParamFile, Required: true, Description: "path to registrar TSV"},
			{Name: "console", Type: ParamBool, Description: "also log to console"},
			{Name: "config", Type: ParamString, Default: defaultConfig},
		},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "frosh-photos",
		Description: "import frosh profile photos from a directory of {unix}.jpg",
		Binary:      InstallDir + "/job-frosh-photos_linux",
		Params: []Param{
			{Name: "in", Type: ParamString, Required: true, Description: "directory of unix.jpg photos"},
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "console", Type: ParamBool},
		},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "dorms-update",
		Description: "reload dorm / room data from embedded fixtures",
		Binary:      InstallDir + "/job-dorms-update_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "local", Type: ParamBool, Default: "true"},
			{Name: "console", Type: ParamBool},
		},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})

	mustRegister(Job{
		Name:        "update-all-factrak-survey-deficits",
		Description: "recompute Factrak survey deficits",
		Binary:      InstallDir + "/job-update-all-factrak-survey-deficits_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "console", Type: ParamBool, Default: "true"},
		},
		Schedule: &Schedule{OnCalendar: "*-02,09-01 00:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "update-all-users-from-ldap",
		Description: "nightly LDAP user sync",
		Binary:      InstallDir + "/job-update-all-users-from-ldap_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "console", Type: ParamBool, Default: "true"},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 01:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "catalog-update",
		Description: "refresh courses.json and Factrak-enriched catalog",
		Binary:      InstallDir + "/job-catalog-update_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "file", Type: ParamString, Default: "/home/wso/wso/static/courses.json"},
			{Name: "factrak-file", Type: ParamString, Default: "/home/wso/wso/static-auth/courses.json"},
			{Name: "console", Type: ParamBool},
			{Name: "draft", Type: ParamBool},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 03:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "user-csv-data",
		Description: "import OIT student CSV (pronouns, off-cycle)",
		Binary:      InstallDir + "/job-user-csv-data_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "file", Type: ParamString, Default: "/var/sftp/oit/wso_student_data.csv"},
			{Name: "console", Type: ParamBool, Default: "true"},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 04:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "dining-update",
		Description: "fetch dining menus into dated JSON files",
		Binary:      InstallDir + "/job-dining-update_linux",
		Params: []Param{
			{Name: "outdir", Type: ParamString, Default: "/home/wso/wso/static/dining"},
			{Name: "vi", Type: ParamString, Default: "/opt/wso-go/vendor_info.yaml"},
			{Name: "today", Type: ParamString},
			{Name: "force", Type: ParamBool},
			{Name: "date", Type: ParamString},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 01:00:00"},
	})
	mustRegister(Job{
		Name:        "library-hours-update",
		Description: "refresh library hours JSON",
		Binary:      InstallDir + "/job-library-hours-update_linux",
		Params: []Param{
			{Name: "out", Type: ParamString, Default: "/home/wso/wso/static/library.json"},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 01:00:00"},
	})
	mustRegister(Job{
		Name:        "schedule-notifs",
		Description: "send a named scheduled notification (e.g. salmon)",
		Binary:      InstallDir + "/job-schedule-notifs_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "notif", Type: ParamString, Default: "salmon"},
			{Name: "console", Type: ParamBool, Default: "true"},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 09:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "update-on-campus-semesters",
		Description: "increment on-campus semester counters (Feb & Sep)",
		Binary:      InstallDir + "/job-update-on-campus-semesters_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "console", Type: ParamBool, Default: "true"},
		},
		Schedule: &Schedule{OnCalendar: "*-02,09-01 00:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "update-profs-areas-of-study",
		Description: "refresh professor areas of study from reviews",
		Binary:      InstallDir + "/job-update_profs_areas_of_study_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "console", Type: ParamBool, Default: "true"},
		},
		Schedule: &Schedule{OnCalendar: "*-04,08,11-01 00:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "ephmatch-reset",
		Description: "yearly EphMatch reset (Sep 1)",
		Binary:      InstallDir + "/job-ephmatch-reset_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "console", Type: ParamBool, Default: "true"},
		},
		Schedule: &Schedule{OnCalendar: "*-09-01 00:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "ephmatch-update-dates",
		Description: "create EphMatch dating windows (Jan 1)",
		Binary:      InstallDir + "/job-ephmatch_update_dates_linux",
		Params: []Param{
			{Name: "file", Type: ParamString, Default: defaultConfig},
		},
		Schedule: &Schedule{OnCalendar: "*-01-01 00:00:00"},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
}

func mustRegister(j Job) {
	if err := defaultRegistry.Register(j); err != nil {
		panic(err)
	}
}
