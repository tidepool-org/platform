#!/usr/bin/env bash
set -euo pipefail
umask 077
mkdir -p "$HOME" /tmp/platform-mongodb
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    printf 'machine github.com\n  login %s\n' "$GITHUB_TOKEN" > "$HOME/.netrc"
fi
export TIMING_CMD='time -p'
go version
go env GOCACHE GOMODCACHE GOOS GOARCH
git --version
docker buildx version
mongosh --version

echo "Starting isolated MongoDB replica set"
mongod --fork --dbpath /tmp/platform-mongodb --bind_ip 127.0.0.1 \
    --replSet rs0 --logpath /tmp/platform-mongodb/mongod.log
mongosh --quiet --eval '
    rs.initiate();
    for (let attempt = 0; attempt < 60; attempt++) {
        if (db.hello().isWritablePrimary) { quit(0); }
        sleep(1000);
    }
    throw new Error("MongoDB replica set did not become ready");
'
time -p make "plugins-visibility-${PLUGINS_VISIBILITY}"
if [[ $# == 0 ]]; then set -- ci; fi
time -p make "$@"
