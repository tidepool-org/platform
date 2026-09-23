# syntax=docker/dockerfile:1
ARG GO_IMAGE
ARG MONGO_IMAGE
ARG DOCKER_IMAGE

FROM ${GO_IMAGE} AS go-tools
ARG MOCKGEN_VERSION
ARG GOIMPORTS_VERSION
WORKDIR /tools-build
# Build both tools in one module so mockgen uses the repository's newer x/tools,
# which supports this Go version, rather than its older standalone dependency.
RUN go mod init github.com/tidepool-org/platform/ci-tools && \
    go get go.uber.org/mock/mockgen@${MOCKGEN_VERSION} golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION} && \
    GOBIN=/tools go install go.uber.org/mock/mockgen golang.org/x/tools/cmd/goimports

FROM ${DOCKER_IMAGE} AS docker-cli

# Sharing this base with platform-tools also avoids downloading MongoDB twice.
# This digest includes mongod 6.0.23 and mongosh 2.5.1.
FROM ${MONGO_IMAGE}
RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential git make time && \
    rm -rf /var/lib/apt/lists/*
COPY --from=go-tools /usr/local/go/ /usr/local/go/
COPY --from=go-tools /tools/ /usr/local/bin/
COPY --from=docker-cli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=docker-cli /usr/local/libexec/docker/cli-plugins/docker-buildx /usr/local/libexec/docker/cli-plugins/docker-buildx
ENV PATH="/usr/local/go/bin:${PATH}" \
    GOTOOLCHAIN=local \
    GOPATH=/home/travis/gopath
# The job supplies its UID, checkout, cache mounts, and ephemeral HOME.
ENTRYPOINT []
CMD ["bash"]
