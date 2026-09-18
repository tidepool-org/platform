# syntax=docker/dockerfile:1
ARG GOLANG_VERSION=1.25.7-alpine
ARG MONGO_VERSION=6.0.23
ARG PLUGIN_VISIBILITY=public

### Bases

# platform-base-alpine
FROM alpine:latest AS platform-base-alpine
RUN apk --no-cache update && apk --no-cache upgrade && apk --no-cache add ca-certificates tzdata
RUN adduser --disabled-password tidepool
USER tidepool
WORKDIR /home/tidepool

# platform-base-golang
FROM golang:${GOLANG_VERSION} AS platform-base-golang
RUN apk --no-cache update && apk --no-cache upgrade && apk --no-cache add ca-certificates tzdata git make

# platform-base-delve
FROM platform-base-golang AS platform-base-delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN adduser --disabled-password delve
USER delve
WORKDIR /home/delve

# platform-base-mongo
FROM mongo:${MONGO_VERSION} AS platform-base-mongo
# Set HOME to override default /data/db
ENV HOME="/home/tidepool/"
ENV DEBIAN_FRONTEND="noninteractive"
RUN apt -y update && apt -y install ca-certificates tzdata
RUN adduser --disabled-password --gecos GECOS tidepool
USER tidepool
WORKDIR /home/tidepool

### Inits

# platform-init
FROM platform-base-golang AS platform-init
WORKDIR /build
COPY Makefile go.* ./
COPY plugin/abbott/go.* ./plugin/abbott/
COPY plugin/visibility/ ./plugin/visibility/

# platform-init-public
FROM platform-init AS platform-init-public
COPY plugin/abbott/abbott/plugin/ ./plugin/abbott/abbott/plugin/
RUN make init plugins-visibility

# platform-init-private
FROM platform-init AS platform-init-private
COPY private/plugin/abbott/go.* ./private/plugin/abbott/
COPY private/plugin/abbott/abbott/plugin/ ./private/plugin/abbott/abbott/plugin/
RUN make init plugins-visibility

### Build

# platform-build
FROM platform-init-${PLUGIN_VISIBILITY} AS platform-build
ARG SERVICE DELVE_PORT
COPY . .
RUN BUILD=services/${SERVICE} make build

# CI supplies _bin as this named context. Ordinary Docker builds still compile
# from source through platform-build, including the development/Delve image.
FROM scratch AS platform-binaries
COPY --from=platform-build /build/_bin/ /

### Delve

# platform-delve
FROM platform-base-delve AS platform-delve
ARG SERVICE DELVE_PORT
ENV SERVICE=${SERVICE} DELVE_PORT=${DELVE_PORT}
COPY --from=platform-build --chown=delve:delve /build/_bin/services/${SERVICE}/ .
CMD exec /go/bin/dlv --listen=:${DELVE_PORT} --headless=true --api-version=2 exec ./${SERVICE}

### Services

# platform-server: all HTTP APIs and background workers share one process.
FROM platform-base-alpine AS platform-server
COPY --from=platform-binaries --chown=tidepool:tidepool /services/server/ .
CMD ["./server"]

# platform-migrations
FROM platform-base-alpine AS platform-migrations
COPY --from=platform-binaries --chown=tidepool:tidepool /services/migrations/ .
CMD ["./migrations"]

# platform-tools
FROM platform-base-mongo AS platform-tools
COPY --from=platform-binaries --chown=tidepool:tidepool /services/tools/ .
COPY ./services/tools/ashrc .bashrc
CMD ["./tools"]
