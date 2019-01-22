# Platform

The Tidepool Platform API.

[![Build Status](https://travis-ci.com/tidepool-org/platform.png)](https://travis-ci.com/tidepool-org/platform)
[![Code Climate](https://codeclimate.com/github/tidepool-org/platform/badges/gpa.svg)](https://codeclimate.com/github/tidepool-org/platform)
[![Issue Count](https://codeclimate.com/github/tidepool-org/platform/badges/issue_count.svg)](https://codeclimate.com/github/tidepool-org/platform)

# Setup

1. Install Go version 1.10 or later.
1. Install `dep` dependency management tool. See https://golang.github.io/dep/docs/installation.html.
1. Create a brand new Go directory.
1. Set the `GOPATH` environment variable to the newly created Go directory.
1. Add `$GOPATH/bin` to the `PATH` environment variable.
1. Execute `go get github.com/tidepool-org/platform` to pull down the project. You may ignore a "no buildable Go source files" warning.
1. Change directory to `$GOPATH/src/github.com/tidepool-org/platform`.
1. Source the `env.sh` file.
1. Execute `make buildable` to install the various Go tools needed for building and editing the project.

For example:

```
brew install go
brew install dep
brew install mongo
brew services start mongodb
mkdir ~/go
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH
go get github.com/tidepool-org/platform
cd $GOPATH/src/github.com/tidepool-org/platform
. ./env.sh
make buildable
```

For reuse, you may want to include the following lines in your shell config (e.g. `.bashrc`) or use a tool like [direnv](http://direnv.net/ 'direnv'):

```
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH
```

# Execute

1. Setup the environment, as above.
1. Build the project.
1. Execute a service.

In addition to the setup above, for example:

```
make build
_bin/services/data/data
```

Use `Ctrl-C` to stop the executable. It may take up to 60 seconds to stop.

# Makefile

* To setup your Go environment for building and editing the project:

```
make buildable
```

* To build the executables:

```
make build
```

All executables are built to the `_bin` directory in a hierarchy that matches the locations of executable source files.

The environment variable `BUILD` indicates which executables to build. If not specified, then all executables are built. For example, to build just the executables found in the `services` directory:

```
BUILD=services make build
```

* To run all of the tests manually:

```
make test
```

The environment variable `TEST` indicates which package hierarchy to test. If not specified, then all packages are tested. For example,

```
TEST=user make test
```

* To run all of the tests automatically after any changes are made, in a separate terminal window:

```
make watch
```

The environment variable `WATCH` indicates which package hierarchy to test. If not specified, then all packages are tested. For example,

```
WATCH=user make watch
```

* To run `gofmt`, `goimports`, `go tool vet`, and `golint`:

```
make pre-commit
```

* To clean the project of all build files:

```
make clean
```

# Sublime Text

If you use the Sublime Text editor with the GoSublime plugin, open the `platform.sublime-project` project to ensure the `GOPATH` and `PATH` environment variables are set correctly within Sublime Text. In addition, the recommended user settings are:

```
{
  "autocomplete_builtins": true,
  "autocomplete_closures": true,
  "autoinst": false,
  "fmt_cmd": [
    "goimports"
  ],
  "fmt_enabled": true,
  "fmt_tab_width": 4,
  "use_named_imports": true
}
```

# Upgrade Golang Version

## Prepare

**Before** you update this repository to use a newer version of Golang, please perform these checks:

- Review the release notes for **all** Golang versions, major and minor, from the current Golang version to the target Golang version. The entire Golang release history can be found at https://golang.org/doc/devel/release.html.
  - For major revisions, if any change described in the release notes could have a negative impact upon this repository, follow up and review any associated issues and the updated code. Make note of this change in order to explicitly test after upgrading.
  - For minor revisions, review all issues included in the associated GitHub milestone issue tracker. These can be found in the minor revision release notes. If any issue could have a negative impact upon this repository, review the updated code. Make note of this issue in order to explicitly test after upgrading.
- Install `gimme`(https://github.com/travis-ci/gimme) via `brew`. Execute `gimme -k`. Ensure that the target Golang version is listed. The `gimme` tool is used by Travis CI to manage Golang versions. If the version is not listed, then the Travis CI build will not succeed.
- Browse to https://hub.docker.com/_/golang and ensure the target Golang version in an Alpine Linux image is available. For example, if the target version is `1.11.4`, then ensure that the `1.11.4-alpine` image tag is available. If the image tag is not avaiable, then the Travis CI build will not succeed.

## Upgrade

Ensure you are using the target Golang version locally.

Change the version in `.travis.yml` and all `Dockerfile.*` files.

Add an entry in `CHANGELOG.md` and commit.

## Test

Ensure the `ci-build` and `ci-test` Makefile targets pass using the target Golang version.

If you previously noted any changes or issues of concern, perform any explicit tests necessary.

# Upgrade Dependencies

## Upgrade

Install `dep` via `brew`. Execute `dep ensure -update`.

## Review

Review all pending changes to all dependencies. If any changes could have a negative impact upon this repository, make note of this change to explicitly test afterwards.

## Test

Ensure the `ci-build` and `ci-test` Makefile targets pass using the target Golang version.

If you previously noted any changes or issues of concern, perform any explicit tests necessary.
