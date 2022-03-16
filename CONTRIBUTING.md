# Contributing

All you need is a computer with Go installed, and a Replicatd account to test with.

### Updating Goreleaser / Testing Homebrew updates

#### How releases work

In general, push a SemVer tag to the repo on the `main` branch and a GitHub action will trigger, invoking [goreleaser](https://goreleaser.com/) to create the necessary artifacts including go binaries, docker images, and homebrew tap updates. Binaries are availabled on the [releases page](https://github.com/replicatedhq/replicated/releases).

Goreleaser is configured to push updates to our [Hombebrew Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap) at https://github.com/replicatedhq/homebrew-replicated.

These can be installed with

```
# preferred method, more clear
brew install replicatedhq/replicated/cli 

# later
brew upgrade replicatedhq/replicated/cli 
```
or

```
# alternative method. less typing but also less clear, esp on upgrades
brew tap replicatedhq/replicated
brew install cli 

# later
brew upgrade cli
```
#### Testing goreleaser locally

You can test the artifacts goreleaser will create locally!

1. Install [goreleaser](https://goreleaser.com/install/) using `brew install goreleaser`
2. Run `goreleaser release --snapshot --rm-dist`.
3. Check the artifacts created under `dist/`
   1. Check `cli.rb` and make sure it has the expected content.
   2. Check the `replicated_SNAPSHOT-*.tar.gz` files and by using `tar -tvf` making sure they contain the `replicated` binary.
   3. Optionally, follow the instructions below to test your created `cli.rb` with `brew install` or `brew upgrade` by copying it into your tap cache.

#### Testing homebrew tap changes locally

You can edit the tap file directly in your local homebrew cache. E.g. mine is checked out at /Users/dex/homebrew/Library/Taps/replicatedhq/homebrew-replicated/. By editing those files and running e.g. `brew upgrade replicatedhq/replicated/cli`, you can verify the changes.

If you want to allow others to test your homebrew-only changes (that is, just changes to the ruby file) without you needing to create a full release via goreleaser, you can also edit the file in the github repo. Testers may need to `rm -r` their locally cached checkout of the tap (e.g. `rm -r /Users/dex/homebrew/Library/Taps/replicatedhq/homebrew-replicated/`) before running `brew install` or `brew upgrade`.

### Design

Replicated apps can be "platform" or "ship". Avoid deep-in-the-callstack checks for app type. There's a common "Client" class that should handle the switch on appType, and call the appropriate implementation. We would like to avoid having this switch get lower in the call stack.

Sometimes, the two different app types require different parametes (promote release is an example, one takes "required" and one doesn't). Don't normalize these to the lowest common denominator. The goal of this CLI is to provide all functionality, with minimal internal knowledge to manage Replicated apps. The app schemas will continue to be a little different, this CLI should mask these differences while still providing access to all features of both appTypes.

