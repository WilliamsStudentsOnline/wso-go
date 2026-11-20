package jobctl

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedis(t *testing.T) {
	t.Helper()
	err := ConfigureRedisForTest()
	if err != nil && os.Getenv("GITHUB_ACTIONS") == "" {
		t.Skipf("redis not reachable: %v", err)
	}
	require.NoError(t, err)
}

func TestSetJobDisabled(t *testing.T) {
	setupRedis(t)
	ctx := context.Background()

	disabled, err := IsJobDisabled(ctx, "example")
	require.NoError(t, err)
	assert.False(t, disabled)

	require.NoError(t, SetJobDisabled(ctx, "example", true))
	require.NoError(t, SetJobDisabled(ctx, "other", true))

	disabled, err = IsJobDisabled(ctx, "example")
	require.NoError(t, err)
	assert.True(t, disabled)

	names, err := DisabledJobs(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"example": true, "other": true}, names)

	require.NoError(t, SetJobDisabled(ctx, "example", false))
	disabled, err = IsJobDisabled(ctx, "example")
	require.NoError(t, err)
	assert.False(t, disabled)
}
