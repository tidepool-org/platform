# Platform

The Tidepool Platform API.

[![Build Status](https://app.travis-ci.com/tidepool-org/platform.svg?branch=master)](https://app.travis-ci.com/tidepool-org/platform)

# Setup

1. Install Go version 1.11.4 or later
1. Install mongodb (if it is not already installed, or run it from Docker)

    The tests assume that mongodb is listening on 127.0.0.1:27017.
    1. Configure mongodb replica sets (required for tests to pass)

        A single node is all that's required. It can be as simple as simple adding `--replSet rs0` when running mongd, or the equivalent config file change.
1. Start mongodb (if it is not already running)
    1. Initiate the replica set

	    Something like: `mongosh rs.initiate()`
1. Clone this repo
1. Change directory to the path you cloned the repo into
1. Source the `env.sh` file
1. Execute `make buildable` to install the various Go tools needed for building and editing the project

For example:

```
brew install go
brew install mongo
brew services start mongodb
git clone https://github.com/tidepool-org/platform.git
cd platform
. ./env.sh
make buildable
```

# Configure Plugin Visibility

By default, all plugins are configured to use the public versions (which typically are nothing more than a bare minimum shell). Public visilibity is fine for developing, building, or using platform without any of the private plugin functionality.

To change plugin visibility to private:

```
make plugins-visibility-private
```

This will update the private plugin submodules and create a Go workspace to use the private plugin during development and builds.

To restore the plugin visibility to public:

```
make plugins-visibility-public
```

NOTE: Do **NOT** commit any `update = none` changes to `.gitmodules` nor the `go.work` or `go.work.sum` files. Furthermore, ensure no private code is committed to the public platform repository.

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

> **Note:** For testing and development, services are generally run on a local Kubernetes cluster through the [development repo](https://github.com/tidepool-org/development#developing-tidepool-services).

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

* To run `gofmt`, `goimports`, and `go vet`:

```
make pre-commit
```

* To clean the project of all build files:

```
make clean
```

# Incremental CI builds

Travis saves both `GOCACHE` (compiled packages and successful test results) and
`GOMODCACHE` (downloaded modules). Go decides which packages and tests to reuse
from their inputs; CI still requests every package on every run. A new branch
can use Travis's default-branch cache until it has a cache of its own. The first
run after changing the Go version or cache configuration may need to populate
the cache.

`make ci-test-go` keeps race detection and coverage enabled. It omits `-count=1`
and `-shuffle`, since either flag disables Go's test-result cache. Ginkgo still
chooses a random seed when a suite actually runs. To force a full shuffled run,
including after changing an external dependency such as MongoDB, use:

```sh
make ci-test-go-fresh
# Or force tests during the entire CI pipeline:
make ci GOTEST_CI_FLAGS='-buildvcs=false -race -cover -count=1 -shuffle=on'
```

`make ci-build` builds static Linux binaries in `_bin` using the current Go cache.
`make ci-docker` runs `ci-build` and supplies `_bin` as Docker's
`platform-binaries` named build context, so image packaging does not compile Go
again. This requires Buildx (included in the CI tools image). `GOARCH` can select a different
Linux architecture.
Ordinary `make build` and Docker builds without this context still build from
source as before.

To inspect cache reuse locally, run these commands twice with the same Go
version and plugin visibility (MongoDB must be running for the full test suite):

```sh
GODEBUG=gocachetest=1 make ci-test-go TIMING_CMD='time -p'
make ci-build GO_BUILD_FLAGS='-buildvcs=false -x' TIMING_CMD='time -p'
```

The second test run should report `(cached)` for unchanged successful packages;
the second build should omit compiler invocations. A fresh checkout at the same
path does not invalidate Go's source compilation cache. Tests that read fixture
files or inspect file metadata can rerun when checkout timestamps change. CI
leaves those checks intact so changed test inputs cannot be hidden by timestamp
normalization. A new commit also requires linking binaries with the new version
metadata, even when compiled packages are reused.

## Shared CI tools image

Travis first runs one `tools` job, followed by the public/private jobs in parallel.
`ci/tools-image.sh ensure` checks for a content-addressed image in the existing
`tidepool/platform-tools` repository. A cache hit only checks the registry manifest;
it does not build, pull, or push the image. On a miss, it builds and publishes the
image and an inline-cache tag for subsequent tool updates. No application source,
private plugins, credentials, or test results are included in this image.

The image includes Go, MongoDB/mongosh, Git, Make, GCC, mockgen, goimports, Docker
CLI, and Buildx. Base images are pinned in `ci/tools.env`; generator versions come
from `go.mod`. The tag includes the architecture and a hash of the Dockerfile,
base-image pins, and Go/tool versions. Update the Go image pin when updating the
Go directive in `go.mod`. Application-only changes reuse the existing image.

Each matrix job pulls that image, mounts its checkout and Travis's Go caches,
and runs the Makefile as the host user's UID. MongoDB runs inside the job's
container with a fresh replica set. The compiler/test cache is namespaced by tools
image, so a MongoDB/toolchain update reruns tests even when Go source is unchanged.
Module downloads remain reusable across tools images. Docker CLI uses the host
Docker socket to package the binaries. The preparation job does not restore or
upload Go caches.

To exercise the same environment locally in a clean disposable checkout:

```sh
bash ci/tools-image.sh build
CI_TOOLS_LOCAL=true CI_CACHE_ROOT=/tmp/platform-ci-cache \
  PLUGINS_VISIBILITY=public bash ci/run.sh
# Repeat with PLUGINS_VISIBILITY=private after authenticating GitHub or fetching
# the private submodule. The same image works for both configurations.
```

`CI_CACHE_ROOT` keeps the Linux caches separate from native developer caches.
`CI_DOCKER_SOCKET` can override `/var/run/docker.sock` for local Docker setups.
`CI_TOOLS_REPOSITORY` selects another registry repository. Image pulls and each
Makefile phase are timed; compare total pipeline elapsed time, including the
preparation stage and its scheduling delay, when evaluating this experiment.

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

## Test

Ensure the `ci-build` and `ci-test` Makefile targets pass using the target Golang version.

If you previously noted any changes or issues of concern, perform any explicit tests necessary.

# Upgrade Dependencies

## Upgrade

```
go get -u <dependency> # e.g. go get -u github.com/onsi/gomega
go mod tidy
```

## Review

Review all pending changes to all dependencies. If any changes could have a negative impact upon this repository, make note of this change to explicitly test afterwards.

## Test

Ensure the `ci-build` and `ci-test` Makefile targets pass using the target Golang version.

If you previously noted any changes or issues of concern, perform any explicit tests necessary.

## Prometheus Metrics

See source files for further details about and usage of each metric.

### Summary Store

* `tidepool_summary_queue_lag` - (histogram) - the current summary queue lag, in minutes
* `tidepool_summary_queue_length` - (gauge) - the current summary queue length, in number of summaries

### C2C

#### Abbott

* `tidepool_abbott_api_request_count` - (counter) - Abbott API request count, sorted by method, path, and status
* `tidepool_abbott_api_request_duration_seconds` - (histogram) - Abbott API duration of each request, in seconds, sorted by method, path, and status

#### Dexcom

* `tidepool_dexcom_api_request_count` - (counter) - Dexcom API request count, sorted by method, path, and status
* `tidepool_dexcom_api_request_duration_seconds` - (histogram) - Dexcom API duration of each request, in seconds, sorted by method, path, and status
* `tidepool_dexcom_api_request_time_seconds` - (histogram) - Dexcom API duration of each request, as reported in the "request-time" response header from Dexcom, in seconds, sorted by method, path, and status

#### Oura

* `tidepool_oura_api_request_count` - (counter) - Oura API request count, sorted by method, path, and status
* `tidepool_oura_api_request_duration_seconds` - (histogram) - Oura API duration of each request, in seconds, sorted by method, path, and status

### Task

#### Queue

* `tidepool_task_workers_total` - (gauge) - configured number of task queue workers, sorted by queue (per config)
* `tidepool_task_workers_available` - (gauge) - number of available task queue workers, sorted by queue
* `tidepool_task_runner_not_found_total` - (counter) - total number of task runs with no registered runner for the task type, sorted by type (ideally zero)
* `tidepool_task_run_duration_seconds` - (histogram) - duration of task runs in seconds, sorted by type
* `tidepool_task_runner_timeout_exceeded_total` - (counter) - total number of task runs that exceeded the runner timeout, sorted by type and disposition ("blocked", "recovered") (ideally zero)
* `tidepool_task_run_panic_total` - (counter) - total number of task runs that panicked, sorted by type (ideally 0)

#### Store

* `tidepool_task_type_state_total` - (counter) - total number of tasks run, sorted by type and state
* `tidepool_task_type_lost_completion_total` - (counter) - total number of task completions dropped because the claim-token compare-and-swap missed, sorted by type (ideally low-ish)
* `tidepool_task_type_revision_mismatch_total` - (counter) - total number of task revisions that do not match the task revision in the database, sorted by type (ideally zero)
