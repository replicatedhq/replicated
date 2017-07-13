# replicated

This repository provides a client and CLI for interacting with the Replicated Vendor API.
The models are generated from the API's swagger spec.

## CLI

Grab the latest [release](https://github.com/replicatedhq/replicated/releases) and extract it to your path.

```
replicated channel ls --app my-app-slug --token e8d7ce8e3d3278a8b1255237e6310069
```

Set the following env vars to avoid passing them as arguments to each command.
* REPLICATED_APP_SLUG
* REPLICATED_API_TOKEN

## Client

(GoDoc)[https://godoc.org/github.com/replicatedhq/replicated/client]

```golang
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/replicatedhq/replicated/client"
)

func main() {
	token := os.Getenv("REPLICATED_API_TOKEN")
	appSlug := os.Getenv("REPLICATED_APP_SLUG")

	api := client.New(token)

	app, err := api.GetAppBySlug(appSlug)
	if err != nil {
		log.Fatal(err)
	}

	channels, err := api.ListChannels(app.Id)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range channels {
		fmt.Printf("channel %s is on release %d\n", c.Name, c.ReleaseSequence)
	}
}
```

## Development
```make build``` installs the binary to ```$GOPATH/bin```

### Tests
REPLICATED_API_ORIGIN may be set for testing an alternative environment.

Since apps can only be deleted in a login session, set these to cleanup garbage from the tests.
VENDOR_USER_EMAIL should be set to delete app
VENDOR_USER_PASSWORD should be set to delete app

### Releases
Releases are created locally with [goreleaser](https://github.com/goreleaser/goreleaser).
Tag the commit to release then run goreleaser.
Ensure GITHUB_TOKEN is [set](https://github.com/settings/tokens/new).

```
git tag -a v0.1.0 -m "First release" && git push origin v0.1.0
make release
```
