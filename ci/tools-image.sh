#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source ci/tools.env

case "${CI_ARCH:-$(uname -m)}" in
    amd64|x86_64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) echo "Unsupported CI architecture" >&2; exit 1 ;;
esac
go_version=$(awk '$1 == "go" { print $2 }' go.mod)
mockgen_version=$(awk '$1 == "go.uber.org/mock" { print $2 }' go.mod)
goimports_version=$(awk '$1 == "golang.org/x/tools" { print $2 }' go.mod)
[[ "$GO_IMAGE" == "golang:${go_version}-"* ]] || {
    echo "Update ci/tools.env to match Go ${go_version}" >&2; exit 1;
}
[[ -n "$mockgen_version" && -n "$goimports_version" ]]
key=$({ cat ci/Dockerfile ci/tools.env; printf '%s\n' "$go_version" "$mockgen_version" "$goimports_version"; } | shasum -a 256 | awk '{ print $1 }')
repository=${CI_TOOLS_REPOSITORY:-tidepool/platform-tools}
image=${repository}:ci-tools-${arch}-${key}
cache_image=${repository}:ci-tools-cache-${arch}

build_image() {
    docker build --platform="linux/${arch}" --progress=plain \
        --build-arg GO_IMAGE="$GO_IMAGE" --build-arg MONGO_IMAGE="$MONGO_IMAGE" \
        --build-arg DOCKER_IMAGE="$DOCKER_IMAGE" \
        --build-arg MOCKGEN_VERSION="$mockgen_version" \
        --build-arg GOIMPORTS_VERSION="$goimports_version" \
        --build-arg BUILDKIT_INLINE_CACHE=1 --cache-from "$cache_image" \
        --tag "$image" --file ci/Dockerfile ci
}

case "${1:-ref}" in
    ref) printf '%s\n' "$image" ;;
    build) time -p build_image ;;
    ensure)
        error_file=$(mktemp)
        trap 'rm -f "$error_file"' EXIT
        echo "Checking shared tools image: $image"
        if docker manifest inspect "$image" >/dev/null 2>"$error_file"; then
            echo "TOOLS_IMAGE_CACHE=hit (no build or push required)"
            exit 0
        fi
        # A registry outage or authentication error must not trigger a rebuild.
        if ! grep -Eq 'no such manifest|manifest unknown|not found' "$error_file"; then
            cat "$error_file" >&2
            exit 1
        fi
        echo "TOOLS_IMAGE_CACHE=miss"
        : "${DOCKER_USERNAME:?Required to publish the shared tools image}"
        : "${DOCKER_PASSWORD:?Required to publish the shared tools image}"
        printf '%s' "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin
        # Inline cache is optional on the first build of this tools image.
        docker pull "$cache_image" || echo "No previous tools image cache available"
        time -p build_image
        time -p docker push "$image"
        docker tag "$image" "$cache_image"
        time -p docker push "$cache_image"
        ;;
    *) echo "Usage: $0 [ref|build|ensure]" >&2; exit 1 ;;
esac
