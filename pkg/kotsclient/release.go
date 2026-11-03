package kotsclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/version"
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

	err := c.DoJSON(context.TODO(), "POST", path, http.StatusOK, nil, &resp)
	if err != nil {
		return "", errors.Wrap(err, "failed to test release")
	}

	return "", nil
}

func (c *VendorV3Client) GetRelease(appID string, sequence int64) (*types.AppRelease, error) {
	resp := types.KotsGetReleaseResponse{}

	path := fmt.Sprintf("/v3/app/%s/release/%v", appID, sequence)

	err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get release")
	}

	appRelease := types.AppRelease{
		Config:               resp.Release.Spec,
		CreatedAt:            resp.Release.CreatedAt,
		Editable:             !resp.Release.IsReleaseNotEditable,
		Sequence:             resp.Release.Sequence,
		Charts:               resp.Release.Charts,
		CompatibilityResults: resp.Release.CompatibilityResults,
		Channels:             resp.Release.Channels,
		IsHelmOnly:           resp.Release.IsHelmOnly,
		AirgapBuilds:         resp.Release.AirgapBuilds,
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
	err = c.DoJSON(context.TODO(), "POST", url, http.StatusCreated, request, &response)
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
	err = c.DoJSON(context.TODO(), "PUT", url, http.StatusOK, request, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create release")
	}

	return nil
}

// ListReleases returns the app's releases, newest first, fetching pageSize releases per request.
// A limit of 0 returns all releases. A pageSize of 0 uses the default of 20, or limit if smaller.
func (c *VendorV3Client) ListReleases(appID string, limit int, pageSize int) ([]types.ReleaseInfo, error) {
	if pageSize == 0 {
		pageSize = 20
		if limit > 0 && limit < pageSize {
			pageSize = limit
		}
	}

	allReleases := []types.ReleaseInfo{}
	page := 0
	for {
		if version.IsDebugMode() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Fetching releases page %d (%d releases so far)\n", page, len(allReleases))
		}
		releases, err := c.ListReleasesPaged(appID, page, pageSize)
		if err != nil {
			return nil, err
		}
		if len(releases) == 0 {
			break
		}
		allReleases = append(allReleases, releases...)
		if limit > 0 && len(allReleases) >= limit {
			allReleases = allReleases[:limit]
			break
		}
		page += 1
	}

	return allReleases, nil
}

// ListReleasesPaged returns one page of the app's releases, newest first. page is 0-indexed.
func (c *VendorV3Client) ListReleasesPaged(appID string, page int, pageSize int) ([]types.ReleaseInfo, error) {
	resp := types.KotsListReleasesResponse{}
	path := fmt.Sprintf("/v3/app/%s/releases?currentPage=%d&pageSize=%d", appID, page, pageSize)
	err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "list releases")
	}

	releases := []types.ReleaseInfo{}
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
		releases = append(releases, newReleaseInfo)
	}

	return releases, nil
}

func (c *VendorV3Client) PromoteRelease(appID string, opts types.PromoteReleaseOptions) (*types.PromoteReleaseResponse, error) {
	request := types.KotsPromoteReleaseRequest{
		ReleaseNotes:          opts.Notes,
		VersionLabel:          opts.Label,
		IsRequired:            opts.Required,
		ChannelIDs:            opts.ChannelIDs,
		NotifyUsers:           opts.NotifyUsers,
		OmitDetailsInResponse: true,
	}

	resp := types.PromoteReleaseResponse{}
	path := fmt.Sprintf("/v3/app/%s/release/%v/promote", appID, opts.Sequence)
	err := c.DoJSON(context.TODO(), "POST", path, http.StatusOK, request, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetAirgapBuildStatus polls the dedicated airgap build status endpoint for a
// single channel-release. The server returns "pending" (synthesized) until the
// airgap-builder writes its first row, real status after that, and 404 if the
// channel-release does not exist.
func (c *VendorV3Client) GetAirgapBuildStatus(appID string, channelID string, channelSequence int64) (*types.AirgapBuildSummary, error) {
	resp := types.AirgapBuildSummary{}
	path := fmt.Sprintf("/v3/app/%s/channel/%s/release/%d/airgap/status", appID, channelID, channelSequence)
	err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get airgap build status")
	}
	return &resp, nil
}
