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

You can use the replicated CLI command (i.e. `$ replicated --help`). 

#### (Pre-requirement) Config the CLI to have access to the Replicated Vendor portal 

To use the CLI you need a user API Token to connect. You can create one by clicking on `New User API Token` in your 
[Account Settings](https://vendor.replicated.com/account-settings). 

Then, you can set the following environment variable to avoid passing it as an argument to each command

```shell
$ export REPLICATED_API_TOKEN=<Your new User API Token>
```

#### Now, let's check an example

We will check all Applications available and the distribution channels:

To list the applications available run `replicated app ls`, i.e.:

```shell
$ replicated app ls  
ID                             NAME     SLUG              SCHEDULER
2FOfwth3fHauBqCvsZ1OaBAr7MU    test     test-rodent       kots
2FOvdw6IR0oewVPVCcmH12tSRoL    nginx    nginx-sheepdog    kots
```

Then, to check the channels run `replicated channel ls --app <Your APP SLUG OR ID>`, i.e.:

```shell
$ replicated channel ls --app 2FOfwth3fHauBqCvsZ1OaBAr7MU
ID                             NAME        RELEASE    VERSION
2FOfwru7Rq1plkqyZFLH6MLR1fk    Stable      1          0.0.1
2FOfwu2zcDbqR24BVPSjjnkVwIe    Beta        1          0.0.1
2FOfwreTFbmkXtf9bukwh8s1ewb    Unstable    1          0.0.1
```

> **Notes:**
> - If you do not export the environment variable above then, you must pass your User API token via the flag `--token <Your new User API Token>`
> - You can also, set via environment variable your app slug or ID (i.e. `export REPLICATED_APP=<Your APP SLUG OR ID>`). So that, the above 
command would be simply `replicated channel ls`.

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
