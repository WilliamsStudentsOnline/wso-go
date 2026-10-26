package registry

func init() {
	mustRegister(Job{
		Name:        "user-home-info",
		Description: "ingest registrar WSO TSV (home address, SU box, entry)",
		Binary:      InstallDir + "/job-user-home-info_linux",
		Params: []Param{
			{Name: "file", Type: ParamFile, Required: true, Description: "path to registrar TSV"},
			{Name: "console", Type: ParamBool, Description: "also log to console"},
			{Name: "config", Type: ParamString, Default: "/opt/wso-go/config.yaml"},
		},
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
}

func mustRegister(j Job) {
	if err := defaultRegistry.Register(j); err != nil {
		panic(err)
	}
}
