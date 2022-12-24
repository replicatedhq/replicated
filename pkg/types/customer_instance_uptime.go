package types

import (
	"github.com/replicatedhq/replicated/pkg/util"
)

type InstanceUptimeInterval struct {
	StartTime *util.Time              `json:"startTime"`
	EndTime   *util.Time              `json:"endTime"`
	Interval  string                  `json:"interval"`
	Statuses  *InstanceUptimeStatuses `json:"statuses"`
}

type InstanceUptimeStatuses struct {
	Missing     float64 `json:"missing"`
	Unavailable float64 `json:"unavailable"`
	Degraded    float64 `json:"degraded"`
	Updating    float64 `json:"updating"`
	Ready       float64 `json:"ready"`
	Unknown     float64 `json:"unknown"`
}

type CustomerInstanceUptime struct {
	Histogram []InstanceUptimeInterval `json:"histogram"`
}

/*

{
  "histogram": [
    {
      "startTime": "2022-12-10T08:00:00Z",
      "endTime": "2022-12-10T16:00:00Z",
      "interval": "8h",
      "statuses": {
        "missing": 0,
        "unavailable": 0,
        "degraded": 0,
        "updating": 0,
        "ready": 0,
        "unknown": 100
      },
      "_embedded": null,
      "_resources": {
        "_self": {
          "href": "/v4/instance/2JK93avStGaNpwTfiHHUkc2iZ9p/appstatus?endTime=2022-12-10T16%3A00%3A00Z&interval=8h&startTime=2022-12-10T08%3A00%3A00Z",
          "method": "GET",
          "documentation": "https://replicated-vendor-api.readme.io/v4/reference/getinstanceappstatus"
        }
      }
    },
    {
      "startTime": "2022-12-10T16:00:00Z",
      "endTime": "2022-12-11T00:00:00Z",
      "interval": "8h",
      "statuses": {
        "missing": 0,
        "unavailable": 0,
        "degraded": 0,
        "updating": 0,
        "ready": 0,
        "unknown": 100
      },
      "_embedded": null,
      "_resources": {
        "_self": {
          "href": "/v4/instance/2JK93avStGaNpwTfiHHUkc2iZ9p/appstatus?endTime=2022-12-11T00%3A00%3A00Z&interval=8h&startTime=2022-12-10T16%3A00%3A00Z",
          "method": "GET",
          "documentation": "https://replicated-vendor-api.readme.io/v4/reference/getinstanceappstatus"
        }
      }
    },
*/
