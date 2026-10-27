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
	assert.Equal(t, []string{"-file", "/x", "-mode", "a", "-out", "/tmp/out.json"}, args)
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

