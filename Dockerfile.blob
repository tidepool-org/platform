# Development
FROM golang:1.11.4-alpine AS development

RUN ["apk", "add", "--no-cache", "git", "make", "tzdata"]

RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]

WORKDIR /go/src/github.com/tidepool-org/platform

COPY . .

ENV SERVICE=services/blob

RUN ["make", "service-build"]

CMD ["make", "service-start"]

# Release
FROM alpine:latest AS release

RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk add --no-cache ca-certificates tzdata && \
    adduser -D tidepool

WORKDIR /home/tidepool

USER tidepool

COPY --from=development --chown=tidepool /go/src/github.com/tidepool-org/platform/_bin/services/blob/ .

CMD ["./blob"]
