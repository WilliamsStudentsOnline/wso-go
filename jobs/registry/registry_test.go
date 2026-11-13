package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleJob() Job {
	return Job{
		Name:   "example",
		Binary: InstallDir + "/job-example_linux",
		Params: []Param{
			{Name: "file", Type: ParamFile, Required: true, Description: "input path"},
			{Name: "console", Type: ParamBool, Description: "log to console"},
			{Name: "mode", Type: ParamEnum, EnumValues: []string{"a", "b"}, Default: "a"},
			{Name: "out", Type: ParamString, Default: "/tmp/out.json"},
		},
		Schedule: &Schedule{OnCalendar: "*-*-* 01:00:00"},
	}
}

func TestValidate(t *testing.T) {
	j := sampleJob()
	require.NoError(t, j.Validate())

	j.Name = ""
	assert.Error(t, j.Validate())

	j = sampleJob()
	j.Binary = ""
	assert.Error(t, j.Validate())

	j = sampleJob()
	j.Schedule = &Schedule{}
	assert.Error(t, j.Validate())

	j = sampleJob()
	j.Params = append(j.Params, Param{Name: "file", Type: ParamString})
	assert.Error(t, j.Validate())

	j = sampleJob()
	j.Params = []Param{{Name: "mode", Type: ParamEnum}}
	assert.Error(t, j.Validate())
}

func TestBuildArgs(t *testing.T) {
	j := sampleJob()

	args, err := j.BuildArgs(map[string]string{
		"file":    "/tmp/in.tsv",
		"console": "true",
		"mode":    "b",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"-file", "/tmp/in.tsv",
		"-console",
		"-mode", "b",
		"-out", "/tmp/out.json",
	}, args)

	_, err = j.BuildArgs(map[string]string{})
	assert.Error(t, err)

	_, err = j.BuildArgs(map[string]string{"file": "/x", "nope": "1"})
	assert.Error(t, err)

	_, err = j.BuildArgs(map[string]string{"file": "/x", "mode": "z"})
	assert.Error(t, err)

	_, err = j.BuildArgs(map[string]string{"file": "/x", "console": "maybe"})
	assert.Error(t, err)

	args, err = j.BuildArgs(map[string]string{"file": "/x", "console": "false"})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"-file", "/x",
		"-console=false",
		"-mode", "a",
		"-out", "/tmp/out.json",
	}, args)
}

func TestManualOnlyDormsUpdateRequiresLocal(t *testing.T) {
	j, ok := Default().Get("dorms-update")
	require.True(t, ok)
	assert.Nil(t, j.Schedule)

	local, ok := j.ParamByName("local")
	require.True(t, ok)
	assert.True(t, local.Required)
	assert.Empty(t, local.Default)

	_, err := j.BuildArgs(map[string]string{})
	assert.Error(t, err)

	args, err := j.BuildArgs(map[string]string{"local": "true"})
	require.NoError(t, err)
	assert.Contains(t, args, "-local")

	args, err = j.BuildArgs(map[string]string{"local": "false"})
	require.NoError(t, err)
	assert.Contains(t, args, "-local=false")
}

func TestFroshPhotosUsesDir(t *testing.T) {
	j, ok := Default().Get("frosh-photos")
	require.True(t, ok)
	dir, ok := j.ParamByName("dir")
	require.True(t, ok)
	assert.True(t, dir.Required)
	_, ok = j.ParamByName("in")
	assert.False(t, ok)
}

func TestUserHomeInfoFileIsPathString(t *testing.T) {
	j, ok := Default().Get("user-home-info")
	require.True(t, ok)
	file, ok := j.ParamByName("file")
	require.True(t, ok)
	assert.Equal(t, ParamString, file.Type)
	assert.True(t, file.Required)
}

func TestRegistry(t *testing.T) {
	r := New()
	require.NoError(t, r.Register(sampleJob()))
	assert.Error(t, r.Register(sampleJob()))

	j, ok := r.Get("example")
	assert.True(t, ok)
	assert.Equal(t, "example", j.Name)
	assert.Equal(t, "wso-job-example", j.UnitName())
	assert.Equal(t, "wso-job-example.service", j.ServiceFile())
	assert.Equal(t, "wso-job-example.timer", j.TimerFile())
	assert.Equal(t, "wso", j.EffectiveUser())

	manual := sampleJob()
	manual.Name = "manual"
	manual.Schedule = nil
	require.NoError(t, r.Register(manual))

	assert.Len(t, r.All(), 2)
	assert.Len(t, r.Scheduled(), 1)
	assert.Equal(t, "example", r.Scheduled()[0].Name)
}

func TestValidateUnitDeps(t *testing.T) {
	j := sampleJob()
	j.Requires = []string{"mysql.service"}
	j.After = []string{"network-online.target"}
	require.NoError(t, j.Validate())

	j = sampleJob()
	j.Requires = []string{"../evil.service"}
	assert.Error(t, j.Validate())

	j = sampleJob()
	j.Wants = []string{"not-a-unit"}
	assert.Error(t, j.Validate())
}

func TestDefaultHasPilotJobs(t *testing.T) {
	_, ok := Default().Get("user-home-info")
	assert.True(t, ok)
	j, ok := Default().Get("library-hours-update")
	assert.True(t, ok)
	assert.NotNil(t, j.Schedule)
}

func TestDefaultPortedJobs(t *testing.T) {
	_, ok := Default().Get("update-all-users-from-ldap")
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(Default().Scheduled()), 10)
	assert.GreaterOrEqual(t, len(Default().All()), 13)

	for _, j := range Default().Scheduled() {
		_, err := j.BuildArgs(nil)
		require.NoError(t, err, "scheduled job %s", j.Name)
	}
}
