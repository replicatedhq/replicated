package client

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestCurrentChannelReleaseSkipsDemotedReleases(t *testing.T) {
	releases := []types.ChannelRelease{
		{ChannelSequence: 1},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelRelease(releases)

	require.NotNil(t, currentRelease)
	require.EqualValues(t, 1, currentRelease.ChannelSequence)
}

func TestCurrentChannelReleaseReturnsNilWhenAllReleasesAreDemoted(t *testing.T) {
	releases := []types.ChannelRelease{
		{ChannelSequence: 1, IsDemoted: true},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelRelease(releases)

	require.Nil(t, currentRelease)
}

func TestCurrentChannelReleasePtrsSkipsDemotedReleases(t *testing.T) {
	releases := []*types.ChannelRelease{
		{ChannelSequence: 1},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelReleasePtrs(releases)

	require.NotNil(t, currentRelease)
	require.EqualValues(t, 1, currentRelease.ChannelSequence)
}

func TestCurrentChannelReleasePtrsReturnsNilWhenAllReleasesAreDemoted(t *testing.T) {
	releases := []*types.ChannelRelease{
		{ChannelSequence: 1, IsDemoted: true},
		{ChannelSequence: 2, IsDemoted: true},
	}

	currentRelease := currentChannelReleasePtrs(releases)

	require.Nil(t, currentRelease)
}
