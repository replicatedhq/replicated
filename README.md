# Replicated Vendor CLI

This repository provides a Go client and CLI for interacting with the Replicated Vendor API.

## CLI


### Mac Install
```
brew install replicatedhq/replicated/cli
```

### Linux Install
```
curl -o install.sh -sSL https://raw.githubusercontent.com/replicatedhq/replicated/master/install.sh
sudo bash ./install.sh
```

### Getting Started
```
replicated channel ls --app my-app-slug --token e8d7ce8e3d3278a8b1255237e6310069
```

Set the following env vars to avoid passing them as arguments to each command.
* REPLICATED_APP - either an app slug or app ID
* REPLICATED_API_TOKEN

Then the above command would be simply
```
replicated channel ls
```

### CI Example
Creating a new release for every tagged build is a common use of the replicated command.

Assume the app's yaml config is checked in at replicated.yaml and you have configured TravisCI or CircleCI with your REPLICATED_APP and REPLICATED_API_TOKEN environment variables.

Then add  a release.sh script to your project something like this:

```bash
#!/bin/bash

# Create a new release from replicated.yaml and promote the Unstable channel to use it.
# Aborts if version tag is empty.

set -e

VERSION=$1
INSTALL_SCRIPT=https://raw.githubusercontent.com/replicatedhq/replicated/master/install.sh
CHANNEL=Unstable

if [ -z "$VERSION" ]; then
echo "No version; skipping replicated release"
  exit
fi

# install replicated
curl -sSL "$INSTALL_SCRIPT" > install.sh
sudo bash ./install.sh

cat replicated.yaml | replicated release create --yaml - --promote Unstable --version "$VERSION"
# Channel ee9d99e87b4a5acc2863f68cb2a0c390 successfully set to release 15
```

Now you can automate tagged releases in TravisCI or CircleCI:

```yaml
# .travis.yml
sudo: required
after_success:
  - ./release.sh "$TRAVIS_TAG"

```

```yaml
# circle.yml
deployment:
  tag:
    tag: /v.*/
    owner: replicatedcom
    commands:
      - ./release.sh "$CIRCLE_TAG"
```

## Client

[GoDoc](https://godoc.org/github.com/replicatedhq/replicated/client)

```golang
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/replicatedhq/replicated/pkg/platformclient"
)

func main() {
	token := os.Getenv("REPLICATED_API_TOKEN")
	appSlugOrID := os.Getenv("REPLICATED_APP")

	api := platformclient.New(token)

	app, err := api.GetApp(appSlugOrID)
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
The models are generated from the API's swagger spec.

### Tests

#### Environment
* ```REPLICATED_API_ORIGIN``` may be set to override the API endpoint
* ```VENDOR_USER_EMAIL``` and ```VENDOR_USER_PASSWORD``` should be set to delete apps created for testing

### Releases
Releases are created on Travis when a tag is pushed. This will also update the docs container.
```
git tag -a v0.1.0 -m "First release" && git push upstream v0.1.0
```

### Regenerating Client Code

When the swagger definitions change, you can regenerate the Client code from the swagger spec with

    make get-spec-prod gen-models

models for the v2 api isn't really working yet, need to find the URL for that OpenAPI spec.

## Usage Recipes

#### Make a new release by editing another
```
replicated release inspect 130 | sed 1,4d > config.yaml
vim config.yaml
cat config.yaml | replicated release create --yaml -
# SEQUENCE: 131
```
