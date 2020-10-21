#!/bin/sh -ex

if [ -f artifact_go.sh ]; then
  rm -f artifact_go.sh
fi
wget -q -O artifact_go.sh 'https://raw.githubusercontent.com/mdblp/tools/dblp/artifact/artifact_go.sh'
chmod +x artifact_go.sh
chmod +x build.sh

export ARTIFACT_GO_VERSION='1.11.4'
# Disable binary deployment in artifact_go.sh as it is done by the Make command
export ARTIFACT_DEPLOY=false
export ARTIFACT_BUILD=true
export BUILD_OPENAPI_DOC=false
export SECURITY_SCAN=true

./artifact_go.sh service=data 
