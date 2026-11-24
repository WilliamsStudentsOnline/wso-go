package jobctl

import (
	"context"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedis(t *testing.T) {
	t.Helper()
	utils.RequireRedis(t, ConfigureRedisForTest())
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
