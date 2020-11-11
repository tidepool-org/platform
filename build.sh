#!/bin/sh -e

make ci-generate ci-build ci-test ci-deploy
GO111MODULE="on" make ci-soups
echo "build done"
