package jobctl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenUnitsGolden(t *testing.T) {
	r := registry.New()
	require.NoError(t, r.Register(registry.Job{
		Name:        "library-hours-update",
		Description: "refresh library hours JSON",
		Binary:      registry.InstallDir + "/job-library-hours-update_linux",
		Params: []registry.Param{
			{Name: "out", Type: registry.ParamString, Default: "/home/wso/wso/static/library.json"},
		},
		Schedule:      &registry.Schedule{OnCalendar: "*-*-* 01:00:00"},
		RuntimeMaxSec: 600,
		MemoryMax:     "256M",
	}))
	require.NoError(t, r.Register(registry.Job{
		Name:   "manual-only",
		Binary: registry.InstallDir + "/job-manual-only_linux",
		Params: []registry.Param{
			{Name: "file", Type: registry.ParamFile, Required: true},
		},
	}))

	out := t.TempDir()
	require.NoError(t, GenUnits(r, out))

	assertGolden(t, out, "wso-job-library-hours-update.service")
	assertGolden(t, out, "wso-job-library-hours-update.timer")

	_, err := os.Stat(filepath.Join(out, "wso-job-manual-only.service"))
	assert.True(t, os.IsNotExist(err), "manual jobs with required params should not get a static unit")
}

func TestGenUnitsPrune(t *testing.T) {
	out := t.TempDir()
	stale := filepath.Join(out, "wso-job-gone.service")
	require.NoError(t, os.WriteFile(stale, []byte("stale"), 0644))
	keepOther := filepath.Join(out, "unrelated.service")
	require.NoError(t, os.WriteFile(keepOther, []byte("keep"), 0644))

	r := registry.New()
	require.NoError(t, GenUnits(r, out))

	_, err := os.Stat(stale)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(keepOther)
	assert.NoError(t, err)
}

func assertGolden(t *testing.T, outDir, name string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(outDir, name))
	require.NoError(t, err)

	goldenPath := filepath.Join("testdata", name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll("testdata", 0755))
		require.NoError(t, os.WriteFile(goldenPath, got, 0644))
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "missing golden %s; run with UPDATE_GOLDEN=1", goldenPath)
	assert.Equal(t, string(want), string(got))
}
