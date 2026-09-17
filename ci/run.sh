#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
checkout=$(pwd -P)
image=$(bash ci/tools-image.sh ref)
visibility=${PLUGINS_VISIBILITY:-public}
case "$visibility" in public|private) ;; *) echo "Invalid plugin visibility: $visibility" >&2; exit 1 ;; esac

# A local build can be tested without publishing it to the registry.
if [[ "${CI_TOOLS_LOCAL:-false}" != true ]]; then
    # Only the public, non-PR Travis job publishes a missing tools image. Keep
    # this out of the matrix env so Travis reuses the existing Go-cache keys.
    publish=false
    if [[ "$visibility" == public && "${TRAVIS_PULL_REQUEST:-}" == false ]]; then
        publish=true
    fi
    CI_TOOLS_PUBLISH="${CI_TOOLS_PUBLISH:-$publish}" bash ci/tools-image.sh ensure
fi
cache_root=${CI_CACHE_ROOT:-$HOME}
# Go cannot detect changes in external test inputs such as the MongoDB version.
# Give each tools image its own test/build cache, while reusing module downloads.
build_cache="$cache_root/.cache/go-build/${image##*:}"
mkdir -p "$build_cache" "$cache_root/gopath/pkg/mod"
socket=${CI_DOCKER_SOCKET:-/var/run/docker.sock}
if [[ $(uname -s) == Darwin ]]; then
    socket_group=0
else
    socket_group=$(stat -c '%g' "$socket")
fi
echo "Running ${visibility} CI in $image"
time -p docker run --rm --init \
    --user "$(id -u):$(id -g)" --group-add "$socket_group" \
    --volume "$checkout:/home/travis/build/tidepool-org/platform" \
    --volume "$build_cache:/home/travis/.cache/go-build" \
    --volume "$cache_root/gopath/pkg:/home/travis/gopath/pkg" \
    --volume "$socket:/var/run/docker.sock" \
    --workdir /home/travis/build/tidepool-org/platform \
    --env HOME=/tmp/platform-ci-home \
    --env GOCACHE=/home/travis/.cache/go-build \
    --env GOMODCACHE=/home/travis/gopath/pkg/mod \
    --env "PLUGINS_VISIBILITY=$visibility" \
    --env GITHUB_TOKEN --env DOCKER_USERNAME --env DOCKER_PASSWORD \
    --env TRAVIS_COMMIT --env TRAVIS_BRANCH --env GOTEST_CI_FLAGS \
    "$image" bash ci/container.sh "$@"
