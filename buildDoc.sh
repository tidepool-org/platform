#!/bin/sh -eu
# Generate OpenAPI documentation
GOPATH=${GOPATH:-~/go}
BASEDIR=$(dirname "$0")
echo "Using GOPATH: ${GOPATH}"
export GO111MODULE="on"

if [ ! -x "$GOPATH/bin/swag" ]; then
  echo "Getting swag..."
  go get -u github.com/swaggo/swag/cmd/swag
fi

$GOPATH/bin/swag --version
$GOPATH/bin/swag init --generalInfo $BASEDIR/services/data/data.go --output docs/api/v1/data
