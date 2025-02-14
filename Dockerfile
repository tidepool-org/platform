ARG GOLANG_VERSION=1.23.4-alpine
ARG MONGO_VERSION=6.0.19
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

### Delve

# platform-delve
FROM platform-base-delve AS platform-delve
ARG SERVICE DELVE_PORT
ENV SERVICE=${SERVICE} DELVE_PORT=${DELVE_PORT}
COPY --from=platform-build --chown=delve:delve /build/_bin/services/${SERVICE}/ .
CMD exec /go/bin/dlv --listen=:${DELVE_PORT} --headless=true --api-version=2 exec ./${SERVICE}

### Services

# platform-auth
FROM platform-base-alpine AS platform-auth
COPY --from=platform-build --chown=tidepool:tidepool /build/_bin/services/auth/ .
CMD ["./auth"]

# platform-blob
FROM platform-base-alpine AS platform-blob
COPY --from=platform-build --chown=tidepool:tidepool /build/_bin/services/blob/ .
CMD ["./blob"]

# platform-data
FROM platform-base-alpine AS platform-data
COPY --from=platform-build --chown=tidepool:tidepool /build/_bin/services/data/ .
CMD ["./data"]

# platform-migrations
FROM platform-base-alpine AS platform-migrations
COPY --from=platform-build --chown=tidepool:tidepool /build/_bin/services/migrations/ .
CMD ["./migrations"]

# platform-prescription
FROM platform-base-alpine AS platform-prescription
COPY --from=platform-build --chown=tidepool:tidepool /build/_bin/services/prescription/ .
CMD ["./prescription"]

# platform-task
FROM platform-base-alpine AS platform-task
COPY --from=platform-build --chown=tidepool:tidepool /build/_bin/services/task/ .
CMD ["./task"]

# platform-tools
FROM platform-base-mongo AS platform-tools
COPY --from=platform-build --chown=tidepool:tidepool /build/_bin/services/tools/ .
COPY ./services/tools/ashrc .bashrc
CMD ["./tools"]
