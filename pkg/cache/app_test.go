package cache

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAppByID(t *testing.T) {
	c := Cache{Apps: []types.App{
		{ID: "production-id", Slug: "chartsmith", Scheduler: "kots"},
		{ID: "local-id", Slug: "chartsmith", Scheduler: "kots"},
	}}

	app, err := c.GetAppByID("local-id")
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, "local-id", app.ID)
}

func TestGetAppByIDDoesNotResolveSlugFromOriginAgnosticCache(t *testing.T) {
	c := Cache{Apps: []types.App{
		{ID: "production-id", Slug: "chartsmith", Scheduler: "kots"},
		{ID: "local-id", Slug: "chartsmith", Scheduler: "kots"},
	}}

	app, err := c.GetAppByID("chartsmith")
	require.NoError(t, err)
	assert.Nil(t, app, "slugs must be resolved against the active API origin")
}

func TestGetAppRetainsSlugLookupForOfflineCallers(t *testing.T) {
	c := Cache{Apps: []types.App{
		{ID: "app-id", Slug: "chartsmith", Scheduler: "kots"},
	}}

	app, err := c.GetApp("chartsmith")
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, "app-id", app.ID)
}
