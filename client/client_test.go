// These are integration tests that generate garbage in the Vendor API.
package client

import "os"

var apiKey = os.Getenv("VENDOR_API_KEY")
var apiOrigin = os.Getenv("VENDOR_API_ORIGIN")
var appID = os.Getenv("VENDOR_APP_ID")
