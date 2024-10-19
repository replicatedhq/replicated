# Contributing

All you need is a computer with Go installed, and a Replicated account to test with.

### Design

Replicated apps can be "platform" or "kots". Avoid deep-in-the-callstack checks for app type. There's a common "Client" class that should handle the switch on appType, and call the appropriate implementation. We would like to avoid having this switch get lower in the call stack.

Sometimes, the two different app types require different parametes (promote release is an example, one takes "required" and one doesn't). Don't normalize these to the lowest common denominator. The goal of this CLI is to provide all functionality, with minimal internal knowledge to manage Replicated apps. The app schemas will continue to be a little different, this CLI should mask these differences while still providing access to all features of both appTypes.

## Releasing

In order to release a new version of this CLI, you will need to have `dagger`.

There's a `make` target that will work if you have a couple of specific environment variables set:

```bash
export GITHUB_TOKEN=<your github token>
export OP_SERVICE_ACCOUNT_PRODUCTION=<your onepassword service account production>
```

Then you can run `make release version=x.y.z` to create a new release.
You can also replace `x.y.z` with `major`, `minor`, or `patch` to bump the version accordingly.

So, to ship a patch release, you would run `make release version=patch`.
