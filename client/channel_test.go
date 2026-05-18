package client

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestCurrentChannelReleaseUsesChannelSequence(t *testing.T) {
	releases := []types.ChannelRelease{
		{ChannelSequence: 1},
		{ChannelSequence: 2},
		{ChannelSequence: 3, IsDemoted: true},
	}

	currentRelease := currentChannelRelease(releases, 2)

	require.NotNil(t, currentRelease)
	require.EqualValues(t, 2, currentRelease.ChannelSequence)
}

func TestCurrentChannelReleaseSkipsDemotedReleases(t *testing.T) {
	releases := []types.ChannelRelease{
		{ChannelSequence: 1},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelRelease(releases, 0)

	require.NotNil(t, currentRelease)
	require.EqualValues(t, 1, currentRelease.ChannelSequence)
}

func TestCurrentChannelReleaseReturnsNilWhenAllReleasesAreDemoted(t *testing.T) {
	releases := []types.ChannelRelease{
		{ChannelSequence: 1, IsDemoted: true},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelRelease(releases, 0)

	require.Nil(t, currentRelease)
}

func TestCurrentChannelReleasePtrsUsesChannelSequence(t *testing.T) {
	releases := []*types.ChannelRelease{
		{ChannelSequence: 1},
		{ChannelSequence: 2},
		{ChannelSequence: 3, IsDemoted: true},
	}

	currentRelease := currentChannelReleasePtrs(releases, 2)

	require.NotNil(t, currentRelease)
	require.EqualValues(t, 2, currentRelease.ChannelSequence)
}

func TestCurrentChannelReleasePtrsSkipsDemotedReleases(t *testing.T) {
	releases := []*types.ChannelRelease{
		{ChannelSequence: 1},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelReleasePtrs(releases, 0)

	require.NotNil(t, currentRelease)
	require.EqualValues(t, 1, currentRelease.ChannelSequence)
}

func TestCurrentChannelReleasePtrsReturnsNilWhenAllReleasesAreDemoted(t *testing.T) {
	releases := []*types.ChannelRelease{
		{ChannelSequence: 1, IsDemoted: true},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelReleasePtrs(releases, 0)

	require.Nil(t, currentRelease)
}
