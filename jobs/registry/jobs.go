package registry

const defaultConfig = "/opt/wso-go/static/wso-backend-config.yaml"

var needsMariaDB = []string{"mariadb.service"}

func init() {
	mustRegister(Job{
		Name:        "user-home-info",
		Description: "ingest registrar WSO TSV (home address, SU box, entry)",
		Binary:      InstallDir + "/job-user-home-info_linux",
		Params: []Param{
			{Name: "file", Type: ParamString, Required: true, Description: "path to registrar TSV"},
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
			{Name: "dir", Type: ParamString, Required: true, Description: "directory of unix.jpg photos"},
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
			{Name: "dir", Type: ParamString, Required: true, Description: "directory containing dorms.csv and rooms/"},
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
		// Depends on OnCampusSemesters, so run after update-on-campus-semesters (00:00 the same days)
		Schedule: &Schedule{OnCalendar: "*-03,10-01 01:00:00"},
		Requires: needsMariaDB,
		After:    append([]string{"wso-job-update-on-campus-semesters.service"}, needsMariaDB...),
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
			{Name: "file", Type: ParamString, Default: "/opt/wso-go/static/courses.json"},
			{Name: "factrak-file", Type: ParamString, Default: "/opt/wso-go/static-auth/courses.json"},
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
			{Name: "file", Type: ParamString, Required: true, Description: "path to OIT student CSV"},
			{Name: "console", Type: ParamBool},
		},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "dining-update",
		Description: "fetch dining menus into dated JSON files",
		Binary:      InstallDir + "/job-dining-update_linux",
		Params: []Param{
			{Name: "outdir", Type: ParamString, Default: "/opt/wso-go/static/menus"},
			{Name: "vi", Type: ParamString},
			{Name: "today", Type: ParamString, Default: "/opt/wso-go/static/dining.json"},
			{Name: "force", Type: ParamBool},
			{Name: "date", Type: ParamString},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* *:00:00"},
	})
	mustRegister(Job{
		Name:        "daily-messages-update",
		Description: "normalize today's daily messages into dated JSON files",
		Binary:      InstallDir + "/job-daily-messages-update_linux",
		Params: []Param{
			{Name: "outdir", Type: ParamString, Default: "/opt/wso-go/static/mobile/daily_messages"},
			{Name: "today", Type: ParamString, Default: "/opt/wso-go/static/mobile/daily_messages.json"},
			{Name: "url", Type: ParamString},
		},
		Schedule: &Schedule{OnCalendar: "*:0/15"},
	})
	mustRegister(Job{
		Name:        "library-hours-update",
		Description: "refresh library hours JSON",
		Binary:      InstallDir + "/job-library-hours-update_linux",
		Params: []Param{
			{Name: "out", Type: ParamString, Default: "/opt/wso-go/static/library.json"},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 01:00:00"},
	})
	mustRegister(Job{
		Name:        "williams-record-update",
		Description: "refresh Williams Record articles JSON",
		Binary:      InstallDir + "/job-williams-record-update_linux",
		Params: []Param{
			{Name: "out", Type: ParamString, Default: "/opt/wso-go/static/mobile/williams_record.json"},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* *:15:00"},
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
		Schedule: &Schedule{OnCalendar: "*-03,10-01 00:00:00"},
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
	mustRegister(Job{
		Name:        "backfill_board",
		Description: "copy legacy bulletins/discussions/rides into board_* (idempotent)",
		Binary:      InstallDir + "/job-backfill_board_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "console", Type: ParamBool, Default: "true"},
			{
				Name:        "disable-migration-check",
				Type:        ParamBool,
				Default:     "false",
				Description: "allow run when DB migrations are behind (escape hatch only)",
			},
		},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "backfill_factrak",
		Description: "link Factrak surveys to catalog courses and sections (dry run unless apply)",
		Binary:      InstallDir + "/job-backfill_factrak_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "apply", Type: ParamBool, Required: true, Description: "write links (true) or only report what would change (false)"},
			{Name: "skip-linked", Type: ParamBool, Default: "true", Description: "skip surveys that already have a catalog course"},
			{Name: "full-results", Type: ParamBool, Description: "list every survey in the report, not only unmatched and ambiguous ones"},
			{Name: "report", Type: ParamString, Default: "/var/log/wso-jobs/backfill_factrak-report.json", Description: "where to write the JSON report"},
			{Name: "link-sources-only", Type: ParamBool, Description: "only recompute link sources (after a catalog load), no matching; writes only with apply"},
			{Name: "console", Type: ParamBool},
			{Name: "disable-migration-check", Type: ParamBool, Default: "false", Description: "allow run when DB migrations are behind (escape hatch only)"},
		},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
	mustRegister(Job{
		Name:        "backfill_areas",
		Description: "add areas of study for programs that have none, e.g. East Asian Languages and Cultures (dry run unless apply)",
		Binary:      InstallDir + "/job-backfill_areas_linux",
		Params: []Param{
			{Name: "config", Type: ParamString, Default: defaultConfig},
			{Name: "apply", Type: ParamBool, Required: true, Description: "add the areas (true) or only log what would be added (false)"},
			{Name: "console", Type: ParamBool},
			{Name: "disable-migration-check", Type: ParamBool, Default: "false", Description: "allow run when DB migrations are behind (escape hatch only)"},
		},
		Requires: needsMariaDB,
		After:    needsMariaDB,
	})
}

func mustRegister(j Job) {
	if err := defaultRegistry.Register(j); err != nil {
		panic(err)
	}
}
