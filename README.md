# Platform

The Tidepool Platform API.

[![Build Status](https://app.travis-ci.com/tidepool-org/platform.svg?branch=master)](https://app.travis-ci.com/tidepool-org/platform)

# Setup

1. Install the Go version specified in `go.mod`
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
1. Execute the combined server.

In addition to the setup above, for example:

```
make build BUILD=services/server
_bin/services/server/server
```

Use `Ctrl-C` to stop the executable. It may take up to 60 seconds to stop.

## Combined server

Auth, blob, data, prescription, and task run in one process, with one HTTP
listener and their existing background workers. CI publishes the runtime as
`tidepool/platform-server` (or `tidepool/platform-server-private`). The migration
and administrative executables and their `platform-migrations` and
`platform-tools` images remain separate.

Merge the configuration previously supplied to the five service deployments.
Component settings keep their existing scopes, such as
`TIDEPOOL_AUTH_SERVICE_SECRET` and `TIDEPOOL_BLOB_SERVICE_UNSTRUCTURED_BLOBS_STORE_*`.
Client identities also remain `auth`, `blob`, `data`, `prescription`, and `task`.
Configure the shared listener with `TIDEPOOL_SERVER_ADDRESS` and
`TIDEPOOL_SERVER_TLS`; old component-specific server addresses no longer open
listeners. For example, for HTTP behind a TLS-terminating proxy:

```sh
export TIDEPOOL_SERVER_ADDRESS=:9220
export TIDEPOOL_SERVER_TLS=false
export TIDEPOOL_AUTH_CLIENT_ADDRESS=http://127.0.0.1:9220
export TIDEPOOL_DATA_CLIENT_ADDRESS=http://127.0.0.1:9220
export TIDEPOOL_DATA_SOURCE_CLIENT_ADDRESS=http://127.0.0.1:9220
export TIDEPOOL_TASK_CLIENT_ADDRESS=http://127.0.0.1:9220
```

Update any more-specific client address overrides as well. Internal clients
currently use loopback HTTP and retain their existing service secrets. Keep
`TIDEPOOL_AUTH_CLIENT_EXTERNAL_ADDRESS` pointed at the existing external
authentication API; it must not point at this server. MongoDB, Kafka, object
storage, and APIs supplied by other repositories remain dependencies.

Copy each old deployment's `KAFKA_CONSUMER_GROUP` into its new setting:
`TIDEPOOL_AUTH_SERVICE_EVENTS_CONSUMER_GROUP`,
`TIDEPOOL_BLOB_SERVICE_EVENTS_CONSUMER_GROUP`, and
`TIDEPOOL_DATA_SERVICE_EVENTS_CONSUMER_GROUP`. All three are required and must
be distinct, so every component receives user events and retains its existing
Kafka offsets. The common broker and topic settings still apply to all three.
Set `CLOUD_EVENTS_SOURCE` to the combined deployment's event source identity.

Point the gateway and other callers at the combined listener instead of the
five service addresses. API methods and paths are preserved. `GET /status`
reports the running components after all have initialized; individual status
responses are at `/status/auth`, `/status/blob`, `/status/data`,
`/status/prescription`, and `/status/task`. `GET /v1/metrics` exposes the shared
Prometheus registry. Component authentication and authorization still apply to
their routes. Unexpected duplicate routes prevent startup.

Deploying this image requires replacing the five old service deployments and
updating their gateway, client, health-check, and monitoring configuration.
Scaling the combined deployment scales all five components and their workers
together. The repository does not apply these deployment changes automatically.

The combined server's tests exercise every registered route over HTTP, check
parameter and body preservation and authentication, and verify startup failure
handling and shutdown. The checked-in route inventory makes removed routes
visible in review; private plugin routes are included dynamically. Run them with:

```sh
go test -race ./service/combined
```

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
again. This requires Buildx (included in the CI tools image). `GOARCH` can select
a different Linux architecture.
Ordinary `make build` and Docker builds without this context still build from
source as before.

On Travis, `ci-docker` logs in once and uses `ci/images.hcl` to package and
publish all services in parallel with Buildx Bake. Each image is pushed once;
additional tags copy the published manifest without checking image layers again.
Images retain the branch/commit/timestamp, branch/commit, and branch/latest tags;
`master` also updates `latest`. Private images retain the `-private` suffix.
`CI_DOCKER_IMAGE_PREFIX` can select a local test registry instead of
`tidepool/platform`. Set `CI_DOCKER_OUTPUT=print` to inspect the resolved targets,
or `CI_DOCKER_OUTPUT=load` to build without publishing.
For local validation without registry credentials, set `DOCKER_LOGIN_CMD=true`.

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

Travis starts the public/private jobs in parallel, without a preparation stage.
Each job runs `ci/tools-image.sh ensure` to pull the image identified by the tools
configuration from `tidepool/platform-tools`. An existing image needs no build,
login, or push. If its manifest is missing, the job builds it locally using the
previous tools image's inline cache when available. Other pull failures get up
to three attempts and fail the job rather than trigger a rebuild.

The public, non-PR Travis job also publishes a newly built image and its inline
cache when Docker credentials are available. Other jobs use their local build
without publishing. Both jobs may build on the first run after a tools change;
neither depends on the other finishing. No application source, private plugins,
credentials, or test results are included in this image.

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
Docker socket to package the binaries.

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
`CI_TOOLS_REPOSITORY` selects another registry repository. `CI_TOOLS_PUBLISH=true`
opts into publishing a newly built tools image when credentials are available;
`false` disables publication. Image pulls and each Makefile phase are timed;
compare total pipeline elapsed time when evaluating changes.
Measure a warm run only after both matrix jobs have completed successfully and
uploaded their caches. The first run after a tools-image change repopulates the
compiler/test cache; a registry image hit alone does not mean the Go cache is warm.

Run `python3 -B -m unittest discover -s ci/tests -v` to check tools-image fallback,
retry, and publication behavior without contacting a registry.

# Upgrade Golang Version

## Prepare

**Before** you update this repository to use a newer version of Golang, please perform these checks:

- Review the release notes for **all** Golang versions, major and minor, from the current Golang version to the target Golang version. The entire Golang release history can be found at https://golang.org/doc/devel/release.html.
  - For major revisions, if any change described in the release notes could have a negative impact upon this repository, follow up and review any associated issues and the updated code. Make note of this change in order to explicitly test after upgrading.
  - For minor revisions, review all issues included in the associated GitHub milestone issue tracker. These can be found in the minor revision release notes. If any issue could have a negative impact upon this repository, review the updated code. Make note of this issue in order to explicitly test after upgrading.
- Check https://hub.docker.com/_/golang for the target version's Bookworm image
  (shared CI tools) and Alpine image (ordinary source-based Docker builds).
  Record the Bookworm image's multi-platform digest for `ci/tools.env`.

## Upgrade

Ensure you are using the target Golang version locally.

Update the Go directive in `go.mod`, the Go image version and digest in
`ci/tools.env`, and the Go image version in `Dockerfile`. Check any other
`Dockerfile.*` files and plugin modules for matching version requirements.
Travis obtains Go from the shared tools image; it does not install Go on the host.

## Test

Ensure the `ci-build` and `ci-test` Makefile targets pass using the target Golang version.

Build the updated tools image with `bash ci/tools-image.sh build`, then exercise
both plugin configurations with `CI_TOOLS_LOCAL=true bash ci/run.sh` in a clean
checkout. Use `CI_CACHE_ROOT` to keep container caches separate from native
caches. The changed tools-image key creates a fresh compiler/test cache.

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
