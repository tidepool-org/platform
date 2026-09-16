#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

output=${1:-push}
case "$output" in
    print|load)
        exec docker buildx bake --file ci/images.hcl --progress=plain "--${output}"
        ;;
    push) ;;
    *) echo "Usage: $0 [push|load|print]" >&2; exit 1 ;;
esac

read -r -a services <<< "${CI_SERVICES:?}"
read -r -a tags <<< "${CI_TAGS:?}"
: "${CI_IMAGE_PREFIX:?}"

# The classic Docker image store makes Bake push each tag separately. Publish
# only the unique tag here, then copy its manifest for the remaining aliases.
echo "Building and pushing one tag per service"
CI_TAGS=${tags[0]} time -p docker buildx bake --file ci/images.hcl --progress=plain --push

tag_image() {
    local repository="${CI_IMAGE_PREFIX}-$1${CI_IMAGE_SUFFIX:-}"
    local tag
    local -a tag_args=()
    for tag in "${tags[@]:1}"; do
        tag_args+=(--tag "${repository}:${tag}")
    done
    # Preserve a single-platform manifest instead of wrapping it in an index.
    docker buildx imagetools create --prefer-index=false --progress=plain \
        "${tag_args[@]}" "${repository}:${tags[0]}"
}

publish_aliases() {
    local service index failed=0
    local -a pids=()
    for service in "${services[@]}"; do
        tag_image "$service" &
        pids+=("$!")
    done
    # Wait for every service, and fail CI if any alias could not be published.
    for index in "${!pids[@]}"; do
        if ! wait "${pids[$index]}"; then
            echo "Failed to publish aliases for ${services[$index]}" >&2
            failed=1
        fi
    done
    return "$failed"
}

if (( ${#tags[@]} > 1 )); then
    echo "Publishing remaining tags from registry manifests"
    time -p publish_aliases
fi
