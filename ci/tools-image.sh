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

# Return 2 only for a missing manifest. Authentication, rate limits, transport
# errors and ambiguous failures must never turn a pull into a tools rebuild.
pull_image() {
    local reference=$1 attempt
    for attempt in 1 2 3; do
        if docker pull "$reference" 2>"$error_file"; then
            return 0
        fi
        cat "$error_file" >&2
        # Classic Docker reports MANIFEST_UNKNOWN; the containerd image store
        # reports a failed resolution of this exact reference as "not found".
        if grep -Eiq 'manifest unknown|MANIFEST_UNKNOWN|no such manifest' "$error_file" ||
            grep -Fq -- "failed to resolve reference \"$reference\": $reference: not found" "$error_file"; then
            return 2
        fi
        if [[ "$attempt" -lt 3 ]]; then
            echo "Tools image pull failed; retrying ($attempt/3): $reference" >&2
            sleep "$((attempt * 5))" || return 1
        fi
    done
    return 1
}

case "${1:-ref}" in
    ref) printf '%s\n' "$image" ;;
    build) time -p build_image ;;
    ensure)
        error_file=$(mktemp)
        trap 'rm -f "$error_file"' EXIT
        echo "Pulling shared tools image: $image"
        if time -p pull_image "$image"; then
            echo "TOOLS_IMAGE_CACHE=hit (no build or push required)"
            exit 0
        else
            status=$?
        fi
        [[ "$status" == 2 ]] || exit "$status"
        echo "TOOLS_IMAGE_CACHE=miss"
        # Inline cache is optional on the first build of this tools image.
        if time -p pull_image "$cache_image"; then
            :
        else
            status=$?
            [[ "$status" == 2 ]] || exit "$status"
            echo "No previous tools image cache available"
        fi
        time -p build_image
        if [[ "${CI_TOOLS_PUBLISH:-false}" == true && -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
            printf '%s' "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin
            time -p docker push "$image"
            docker tag "$image" "$cache_image"
            time -p docker push "$cache_image"
        else
            echo "Using locally built tools image (publishing disabled or credentials unavailable)"
        fi
        ;;
    *) echo "Usage: $0 [ref|build|ensure]" >&2; exit 1 ;;
esac
