package kotsclient

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListReleases struct {
	Data   *KotsReleasesData  `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

type KotsReleasesData struct {
	KotsReleases []*KotsRelease `json:"allKotsReleases"`
}

type KotsRelease struct {
	ID           string               `json:"id"`
	Sequence     int64                `json:"sequence"`
	CreatedAt    string               `json:"created"`
	ReleaseNotes string               `json:"releaseNotes"`
	Channels     []*types.KotsChannel `json:"channels"`
}

func (c *VendorV3Client) TestRelease(appID string, sequence int64) (string, error) {
	resp := types.KotsTestReleaseResponse{}

	path := fmt.Sprintf("/v3/app/%s/release/%v/test", appID, sequence)

	err := c.DoJSON("POST", path, http.StatusOK, nil, &resp)
	if err != nil {
		return "", errors.Wrap(err, "failed to test release")
	}

	return "", nil
}

func (c *VendorV3Client) GetRelease(appID string, sequence int64) (*types.AppRelease, error) {
	resp := types.KotsGetReleaseResponse{}

	path := fmt.Sprintf("/v3/app/%s/release/%v", appID, sequence)

	err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get release")
	}

	appRelease := types.AppRelease{
		Config:     resp.Release.Spec,
		CreatedAt:  resp.Release.CreatedAt,
		Editable:   !resp.Release.IsReleaseNotEditable,
		Sequence:   resp.Release.Sequence,
		Charts:     resp.Release.Charts,
		IsHelmOnly: resp.Release.IsHelmOnly,
	}

	return &appRelease, nil
}

func (c *VendorV3Client) CreateRelease(appID string, multiyaml string) (*types.ReleaseInfo, error) {
	gzipData := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(gzipData)
	_, err := io.Copy(gzipWriter, strings.NewReader(multiyaml))
	if err != nil {
		gzipWriter.Close()
		return nil, errors.Wrap(err, "failed to write gzip data")
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close gzip writer")
	}

	request := types.KotsCreateReleaseRequest{
		SpecGzip: gzipData.Bytes(),
	}

	response := types.KotsGetReleaseResponse{}

	url := fmt.Sprintf("/v3/app/%s/release", appID)
	err = c.DoJSON("POST", url, http.StatusCreated, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create release")
	}

	releaseInfo := types.ReleaseInfo{
		AppID:    response.Release.AppID,
		Sequence: response.Release.Sequence,
		Charts:   response.Release.Charts,
	}

	return &releaseInfo, nil
}

func (c *VendorV3Client) UpdateRelease(appID string, sequence int64, multiyaml string) error {
	gzipData := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(gzipData)
	_, err := io.Copy(gzipWriter, strings.NewReader(multiyaml))
	if err != nil {
		gzipWriter.Close()
		return errors.Wrap(err, "failed to write gzip data")
	}

	if err := gzipWriter.Close(); err != nil {
		return errors.Wrap(err, "failed to close gzip writer")
	}

	request := types.KotsUpdateReleaseRequest{
		SpecGzip: gzipData.Bytes(),
	}

	url := fmt.Sprintf("/v3/app/%s/release/%d", appID, sequence)
	err = c.DoJSON("PUT", url, http.StatusOK, request, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create release")
	}

	return nil
}

func (c *VendorV3Client) ListReleases(appID string) ([]types.ReleaseInfo, error) {
	allReleases := []types.ReleaseInfo{}
	done := false
	page := 0
	for !done {
		resp := types.KotsListReleasesResponse{}
		path := fmt.Sprintf("/v3/app/%s/releases?currentPage=%d&pageSize=20", appID, page)
		err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
		if err != nil {
			done = true
			continue
		}
		page += 1
		for _, release := range resp.Releases {
			activeChannels := []types.Channel{}

			for _, kotsReleaseChannel := range release.Channels {
				if kotsReleaseChannel.IsArchived {
					continue
				}
				activeChannel := types.Channel{
					ID:   kotsReleaseChannel.ID,
					Name: kotsReleaseChannel.Name,
				}
				activeChannels = append(activeChannels, activeChannel)
			}

			newReleaseInfo := types.ReleaseInfo{
				ActiveChannels: activeChannels,
				AppID:          release.AppID,
				CreatedAt:      release.CreatedAt,
				Editable:       !release.IsReleaseNotEditable,
				Sequence:       release.Sequence,
			}
			allReleases = append(allReleases, newReleaseInfo)
		}

		if len(resp.Releases) == 0 {
			done = true
			continue
		}
	}

	return allReleases, nil
}

func (c *VendorV3Client) PromoteRelease(appID string, sequence int64, label string, notes string, required bool, channelIDs ...string) error {
	request := types.KotsPromoteReleaseRequest{
		ReleaseNotes: notes,
		VersionLabel: label,
		IsRequired:   required,
		ChannelIDs:   channelIDs,
	}

	path := fmt.Sprintf("/v3/app/%s/release/%v/promote", appID, sequence)
	err := c.DoJSON("POST", path, http.StatusOK, request, nil)
	if err != nil {
		return err
	}

	return nil
}
