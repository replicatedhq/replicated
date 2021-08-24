package kotsclient

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

const embeddedInstallBaseURL = "https://k8s.kurl.sh"

var embeddedInstallOverrideURL = os.Getenv("EMBEDDED_INSTALL_BASE_URL")

// this is not client logic, but sure, let's go with it
func embeddedInstallCommand(appSlug string, c *types.KotsChannel) string {

	kurlBaseURL := embeddedInstallBaseURL
	if embeddedInstallOverrideURL != "" {
		kurlBaseURL = embeddedInstallOverrideURL
	}

	kurlURL := fmt.Sprintf("%s/%s-%s", kurlBaseURL, appSlug, c.ChannelSlug)
	if c.ChannelSlug == "stable" {
		kurlURL = fmt.Sprintf("%s/%s", kurlBaseURL, appSlug)
	}
	return fmt.Sprintf(`    curl -fsSL %s | sudo bash`, kurlURL)

}

func embeddedAirgapInstallCommand(appSlug string, c *types.KotsChannel) string {

	kurlBaseURL := embeddedInstallBaseURL
	if embeddedInstallOverrideURL != "" {
		kurlBaseURL = embeddedInstallOverrideURL
	}

	slug := fmt.Sprintf("%s-%s", appSlug, c.ChannelSlug)
	if c.ChannelSlug == "stable" {
		slug = appSlug
	}
	kurlURL := fmt.Sprintf("%s/bundle/%s.tar.gz", kurlBaseURL, slug)

	return fmt.Sprintf(`    curl -fSL -o %s.tar.gz %s
    # ... scp or sneakernet %s.tar.gz to airgapped machine, then
    tar xvf %s.tar.gz
    sudo bash ./install.sh airgap`, slug, kurlURL, slug, slug)

}

// this is not client logic, but sure, let's go with it
func existingInstallCommand(appSlug string, c *types.KotsChannel) string {

	slug := appSlug
	if c.ChannelSlug != "stable" {
		slug = fmt.Sprintf("%s/%s", appSlug, c.ChannelSlug)
	}

	return fmt.Sprintf(`    curl -fsSL https://kots.io/install | bash
    kubectl kots install %s`, slug)
}

type ListChannelsResponse struct {
	Channels []*types.KotsChannel `json:"channels"`
}

func (c *VendorV3Client) ListChannels(appID string, appSlug string, channelName string) ([]types.Channel, error) {
	var response = ListChannelsResponse{}

	v := url.Values{}
	v.Set("channelName", channelName)
	v.Set("excludeDetail", "true")

	url := fmt.Sprintf("/v3/app/%s/channels?%s", appID, v.Encode())
	err := c.DoJSON("GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list channels")
	}

	channels := make([]types.Channel, 0, 0)
	for _, kotsChannel := range response.Channels {
		channel := types.Channel{
			ID:              kotsChannel.Id,
			Name:            kotsChannel.Name,
			ReleaseLabel:    kotsChannel.CurrentVersion,
			ReleaseSequence: int64(kotsChannel.ReleaseSequence),
			InstallCommands: &types.InstallCommands{
				Existing: existingInstallCommand(appSlug, kotsChannel),
				Embedded: embeddedInstallCommand(appSlug, kotsChannel),
				Airgap:   embeddedAirgapInstallCommand(appSlug, kotsChannel),
			},
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (c *VendorV3Client) CreateChannel(appID, name, description string) (*types.Channel, error) {
	request := types.CreateChannelRequest{
		Name:        name,
		Description: description,
	}

	type createChannelResponse struct {
		Channel types.KotsChannel `json:"channel"`
	}
	var response createChannelResponse

	url := fmt.Sprintf("/v3/app/%s/channel", appID)
	err := c.DoJSON("POST", url, http.StatusCreated, request, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list channels")
	}

	return &types.Channel{
		ID:              response.Channel.Id,
		Name:            response.Channel.Name,
		Description:     response.Channel.Description,
		Slug:            response.Channel.ChannelSlug,
		ReleaseSequence: int64(response.Channel.ReleaseSequence),
		ReleaseLabel:    response.Channel.CurrentVersion,
		IsArchived:      response.Channel.IsArchived,
	}, nil
}

func (c *VendorV3Client) GetChannel(appID string, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error) {
	type getChannelResponse struct {
		Channel types.KotsChannel `json:"channel"`
	}

	response := getChannelResponse{}
	url := fmt.Sprintf("/v3/app/%s/channel/%s", appID, url.QueryEscape(channelID))
	err := c.DoJSON("GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, nil, errors.Wrap(err, "get app channel")
	}

	channelDetail := channels.AppChannel{
		Id:              response.Channel.Id,
		Name:            response.Channel.Name,
		Description:     response.Channel.Description,
		ReleaseLabel:    response.Channel.CurrentVersion,
		ReleaseSequence: int64(response.Channel.ReleaseSequence),
	}
	return &channelDetail, nil, nil
}

func (c *VendorV3Client) ArchiveChannel(appID, channelID string) error {
	url := fmt.Sprintf("/v3/app/%s/channel/%s", appID, url.QueryEscape(channelID))

	err := c.DoJSON("DELETE", url, http.StatusOK, nil, nil)
	if err != nil {
		return errors.Wrap(err, "archive app channel")
	}

	return nil
}
