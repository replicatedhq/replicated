# replicated

This repository provides a client and CLI for interacting with the Replicated Vendor API.
The models are generated from the API's swagger spec.

Grab the latest [release](https://github.com/replicatedhq/replicated/releases) and extract it to your path.

```
replicated channel ls --app my-app-slug --token e8d7ce8e3d3278a8b1255237e6310069
```

Set the following env vars to avoid passing them as arguments to each command.
* REPLICATED_APP_SLUG
* REPLICATED_API_TOKEN

## Development
```make build``` installs the binary to ```$GOPATH/bin```

### Tests
REPLICATED_API_ORIGIN may be set for testing an alternative environment.

### Releases
Releases are created locally with [goreleaser](https://github.com/goreleaser/goreleaser).
Tag the commit to release then run goreleaser.
Ensure GITHUB_TOKEN is [set](https://github.com/settings/tokens/new).

```
git tag -a v0.1.0 -m "First release" && git push origin v0.1.0
make release
```
