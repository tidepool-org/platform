# syntax=docker/dockerfile:1.7-labs

ARG GOLANG_VERSION=1.23.4-alpine
ARG MONGO_VERSION=6.0.19

# tidepool-golang
FROM golang:${GOLANG_VERSION} AS tidepool-golang
RUN apk --no-cache update && apk --no-cache upgrade && apk --no-cache add ca-certificates tzdata git make

# tidepool-alpine
FROM alpine:latest AS tidepool-alpine
RUN apk --no-cache update && apk --no-cache upgrade && apk --no-cache add ca-certificates tzdata

# delve
FROM tidepool-golang AS tidepool-delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# init
FROM tidepool-golang as init
WORKDIR /build
COPY --exclude=*.go . .
RUN make init

# build
FROM init as build
ARG SERVICE DELVE_PORT
WORKDIR /build
COPY . .
RUN BUILD=services/${SERVICE} make plugins-visibility build

# delve
FROM tidepool-alpine AS delve
ARG SERVICE DELVE_PORT
ENV SERVICE=${SERVICE} DELVE_PORT=${DELVE_PORT}
RUN adduser --disabled-password delve
USER delve
WORKDIR /delve
COPY --from=tidepool-delve --chown=delve:delve /go/bin/dlv .
COPY --from=build --chown=delve:delve /build/_bin/services/${SERVICE}/ .
CMD ./dlv --listen=:${DELVE_PORT} --headless=true --api-version=2 exec ./${SERVICE}

# auth
FROM tidepool-alpine AS production-auth
RUN adduser --disabled-password tidepool
USER tidepool
WORKDIR /home/tidepool
COPY --from=build --chown=tidepool:tidepool /build/_bin/services/auth/ .
CMD ./auth

# blob
FROM tidepool-alpine AS production-blob
RUN adduser --disabled-password tidepool
USER tidepool
WORKDIR /home/tidepool
COPY --from=build --chown=tidepool:tidepool /build/_bin/services/blob/ .
CMD ./blob

# data
FROM tidepool-alpine AS production-data
RUN adduser --disabled-password tidepool
USER tidepool
WORKDIR /home/tidepool
COPY --from=build --chown=tidepool:tidepool /build/_bin/services/data/ .
CMD ./data

# migrations
FROM tidepool-alpine AS production-migrations
RUN adduser --disabled-password tidepool
USER tidepool
WORKDIR /home/tidepool
COPY --from=build --chown=tidepool:tidepool /build/_bin/services/migrations/ .
CMD ./migrations

# prescription
FROM tidepool-alpine AS production-prescription
RUN adduser --disabled-password tidepool
USER tidepool
WORKDIR /home/tidepool
COPY --from=build --chown=tidepool:tidepool /build/_bin/services/prescription/ .
CMD ./prescription

# task
FROM tidepool-alpine AS production-task
RUN adduser --disabled-password tidepool
USER tidepool
WORKDIR /home/tidepool
COPY --from=build --chown=tidepool:tidepool /build/_bin/services/task/ .
CMD ./task

# tools
FROM mongo:${MONGO_VERSION} AS production-tools
# Set HOME to override default /data/db
ENV HOME="/home/tidepool/"
ENV DEBIAN_FRONTEND="noninteractive"
RUN apt -y update && apt -y install ca-certificates tzdata
RUN adduser --disabled-password --gecos GECOS tidepool
USER tidepool
WORKDIR /home/tidepool
COPY ./services/tools/ashrc /home/tidepool/.bashrc
COPY --from=build --chown=tidepool:tidepool /build/_bin/services/tools/ .
CMD ./tools
