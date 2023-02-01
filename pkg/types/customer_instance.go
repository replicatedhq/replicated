package types

import (
	"fmt"
	"github.com/replicatedhq/replicated/pkg/util"
	"math"
	"time"
)

type CustomerInstanceRaw struct {
	ID string `json:"id"`

	//app info
	AppStatus      string     `json:"appStatus"`
	AppVersion     string     `json:"versionLabel"`
	VersionsBehind int        `json:"numberVersionsBehind"`
	LastCheckIn    *util.Time `json:"lastCheckinAt"`

	// install info
	K8sDistribution     string     `json:"k8sDistribution"`
	K8sVersion          string     `json:"k8sVersion"`
	KOTSVersion         string     `json:"kotsVersion"`
	FirstCheckinAt      *util.Time `json:"firstCheckinAt"`
	CloudProvider       string     `json:"cloudProvider"`
	CloudProviderRegion string     `json:"cloudProviderRegion"`

	// insights inputs
	CurrentVersionReleasedAt *util.Time `json:"currentVersionReleasedAt"`
	LatestVersionReleasedAt  *util.Time `json:"latestVersionReleasedAt"`
	FirstReadyAt             *util.Time `json:"firstReadyAt"`
}

type CustomerInstanceInsights struct {
	LicenseCreatedAt time.Time `json:"licenseCreatedAt"`

	AbsoluteVersionAge    time.Duration  `json:"absoluteVersionAge"`
	RelativeVersionAge    time.Duration  `json:"relativeVersionAge"`
	InstanceTimeToInstall *time.Duration `json:"instanceTimeToInstall"`
	LicenseTimeToInstall  *time.Duration `json:"licenseTimeToInstall"`
	Uptime                float64        `json:"uptime"`
	UpgradeSuccessRate    float64        `json:"upgradeSuccessRate"`
	UpgradesCompleted     int            `json:"upgradesCompleted"`
}

type CustomerInstance struct {
	Instance CustomerInstanceRaw      `json:"instance"`
	Insights CustomerInstanceInsights `json:"insights"`
}

func (c *CustomerInstanceRaw) WithInsights(licenseCreated time.Time) CustomerInstance {
	instance := CustomerInstance{
		Instance: *c,
		Insights: CustomerInstanceInsights{
			LicenseCreatedAt:   licenseCreated,
			AbsoluteVersionAge: time.Now().Sub(c.CurrentVersionReleasedAt.Time),
			RelativeVersionAge: c.LatestVersionReleasedAt.Time.Sub(c.CurrentVersionReleasedAt.Time),
		},
	}

	if c.FirstReadyAt != nil {
		licenseTTI := c.FirstReadyAt.Time.Sub(licenseCreated)
		instance.Insights.LicenseTimeToInstall = &licenseTTI
		if c.FirstCheckinAt != nil {
			instanceTTI := c.FirstReadyAt.Time.Sub(c.FirstCheckinAt.Time)
			instance.Insights.InstanceTimeToInstall = &instanceTTI
		}
	}

	return instance
}

func (c *CustomerInstance) ShortID() string {
	return fmt.Sprintf("%.7s", c.Instance.ID)
}

func (c *CustomerInstance) AgeAbsolute() string {
	return c.FormatDuration(&c.Insights.RelativeVersionAge)
}
func (c *CustomerInstance) AgeRelative() string {
	return c.FormatDuration(&c.Insights.RelativeVersionAge)
}

func (c *CustomerInstance) TTILicense() string {
	return c.FormatDuration(c.Insights.LicenseTimeToInstall)
}

func (c *CustomerInstance) TTIInstance() string {
	return c.FormatDuration(c.Insights.InstanceTimeToInstall)
}

func (c *CustomerInstance) FormatDuration(d *time.Duration) string {
	if d == nil {
		return "--"
	}

	duration := *d
	if duration == 0 {
		return "0s"
	}
	if duration < 1*time.Hour {
		minutes := math.Trunc(duration.Minutes())
		seconds := duration.Seconds() - minutes*60
		return fmt.Sprintf("%.0fm%.0fs", minutes, seconds)
	}
	if duration < 24*time.Hour {
		hours := math.Trunc(duration.Hours())
		minutes := duration.Minutes() - hours*60
		return fmt.Sprintf("%.0fh%.0fm", hours, minutes)
	}
	if duration < 7*24*time.Hour {
		days := math.Trunc(duration.Hours() / 24)
		hours := duration.Hours() - (days * 24)
		return fmt.Sprintf("%.0fd%.0fh", days, hours)
	}

	return fmt.Sprintf("%.1fd", duration.Hours()/24)
}

/*


[{"id":"2JEmiplxi5hu5p2nNE1o8sFq3fd",
"firstCheckinAt":"2022-12-21T19:43:54.277Z",
"lastCheckinAt":"2022-12-23T17:15:00.722Z",
"firstReadyAt":"2022-12-21T19:43:59.455Z",
"licenseCreatedAt":"2022-12-21T19:32:37Z",
"appStatus":"ready",
"versionLabel":"1.0.2",
"numberVersionsBehind":3,
"currentVersionReleasedAt":"2022-12-21T17:03:31Z",
"latestVersionReleasedAt":"2022-12-23T23:00:52Z",
"isKurl":false,
"k8sVersion":"v1.23.12-gke.1600",
"k8sDistribution":"gke",
"k8sVersionsBehind":0,
"cloudProvider":"gcp",
"cloudProviderRegion":"us-central1",
"kotsInstanceId":"dxuhskgrddicjhtvvlfyxcrxxbhbvjvh",
"kotsVersion":"v1.91.3",
"isAirgap":false,
"_embedded":{"app":{"id":"2C0OQyxiiA4pc3blK0cKs5VQS3N",
"name":"wordpress-enterprise",
"slug":"wordpress-enterprise",
"createdAt":"2022-07-16T01:47:04Z",
"customHostnames":{},
"resourceUrls":{"_self":{"href":"/v4/app/2C0OQyxiiA4pc3blK0cKs5VQS3N",
"method":"GET",
"documentation":"https://help.replicated.com/api/vendor-api/#get-v4-app-appid"},
"channels":{"list":{"href":"/v4/channels?appSelector=wordpress-enterprise",
"method":"GET",
"documentation":"https://help.replicated.com/api/vendor-api/#get-v4-apps-appid-channels"}}}},
"channel":{"id":"2C0OR5HUAAge0xo2ggEwGwOHSG4",
"name":"Stable",
"slug":"stable",
"createdAt":"2022-07-16T01:47:03Z",
"isDefault":true,
"_embedded":{"app":{"id":"2C0OQyxiiA4pc3blK0cKs5VQS3N",
"name":"wordpress-enterprise",
"slug":"wordpress-enterprise",
"createdAt":"2022-07-16T01:47:04Z",
"customHostnames":{},
"resourceUrls":{"_self":{"href":"/v4/app/2C0OQyxiiA4pc3blK0cKs5VQS3N",
"method":"GET",
"documentation":"https://help.replicated.com/api/vendor-api/#get-v4-app-appid"},
"channels":{"list":{"href":"/v4/channels?appSelector=wordpress-enterprise",
"method":"GET",
"documentation":"https://help.replicated.com/api/vendor-api/#get-v4-apps-appid-channels"}}}}},
"_resources":{"_self":{"href":"/v4/channel/2C0OR5HUAAge0xo2ggEwGwOHSG4",
"method":"GET",
"documentation":"https://help.replicated.com/api/vendor-api/#get-v4-channel-channelid"},
"releases":{"list":{"href":"/v4/releases?channelSelector=stable",
"method":"GET",
"documentation":"https://help.replicated.com/api/vendor-api/#get-v4-channels-channelid-releases"}}}}},
"_resources":{"_self":{"href":"/v4/instance/2JEmiplxi5hu5p2nNE1o8sFq3fd",
"method":"GET",
"documentation":"https://replicated-vendor-api.readme.io/v4/reference/getinstance"},
"events":{"href":"/v4/instance/2JEmiplxi5hu5p2nNE1o8sFq3fd/events",
"method":"GET",
"documentation":"https://replicated-vendor-api.readme.io/v4/reference/listinstanceevents"}}},
{"id":"2JKZifK6EvX5vSL4b5vUVSgiqjC",
"firstCheckinAt":"2022-12-23T20:55:45.551Z",
"lastCheckinAt":"2022-12-24T16:54:00.944Z",
"firstReadyAt":"2022-12-23T20:55:49.166Z","licenseCreatedAt":"2022-12-21T19:32:37Z","appStatus":"ready","versionLabel":"1.2.0","numberVersionsBehind":0,"currentVersionReleasedAt":"2022-12-23T23:00:52Z","latestVersionReleasedAt":"2022-12-23T23:00:52Z","isKurl":false,"k8sVersion":"v1.23.12-gke.1600","k8sDistribution":"gke","k8sVersionsBehind":0,"cloudProvider":"gcp","cloudProviderRegion":"us-central1","kotsInstanceId":"sywujpjnnoipmonrlrmxkdlshootdpsi","kotsVersion":"v1.91.3","isAirgap":false,"_embedded":{"app":{"id":"2C0OQyxiiA4pc3blK0cKs5VQS3N","name":"wordpress-enterprise","slug":"wordpress-enterprise","createdAt":"2022-07-16T01:47:04Z","customHostnames":{},"resourceUrls":{"_self":{"href":"/v4/app/2C0OQyxiiA4pc3blK0cKs5VQS3N","method":"GET","documentation":"https://help.replicated.com/api/vendor-api/#get-v4-app-appid"},"channels":{"list":{"href":"/v4/channels?appSelector=wordpress-enterprise","method":"GET","documentation":"https://help.replicated.com/api/vendor-api/#get-v4-apps-appid-channels"}}}},"channel":{"id":"2C0OR5HUAAge0xo2ggEwGwOHSG4","name":"Stable","slug":"stable","createdAt":"2022-07-16T01:47:03Z","isDefault":true,"_embedded":{"app":{"id":"2C0OQyxiiA4pc3blK0cKs5VQS3N","name":"wordpress-enterprise","slug":"wordpress-enterprise","createdAt":"2022-07-16T01:47:04Z","customHostnames":{},"resourceUrls":{"_self":{"href":"/v4/app/2C0OQyxiiA4pc3blK0cKs5VQS3N","method":"GET","documentation":"https://help.replicated.com/api/vendor-api/#get-v4-app-appid"},"channels":{"list":{"href":"/v4/channels?appSelector=wordpress-enterprise","method":"GET","documentation":"https://help.replicated.com/api/vendor-api/#get-v4-apps-appid-channels"}}}}},"_resources":{"_self":{"href":"/v4/channel/2C0OR5HUAAge0xo2ggEwGwOHSG4","method":"GET","documentation":"https://help.replicated.com/api/vendor-api/#get-v4-channel-channelid"},"releases":{"list":{"href":"/v4/releases?channelSelector=stable","method":"GET","documentation":"https://help.replicated.com/api/vendor-api/#get-v4-channels-channelid-releases"}}}}},"_resources":{"_self":{"href":"/v4/instance/2JKZifK6EvX5vSL4b5vUVSgiqjC","method":"GET","documentation":"https://replicated-vendor-api.readme.io/v4/reference/getinstance"},"events":{"href":"/v4/instance/2JKZifK6EvX5vSL4b5vUVSgiqjC/events","method":"GET","documentation":"https://replicated-vendor-api.readme.io/v4/reference/listinstanceevents"}}}]


*/
