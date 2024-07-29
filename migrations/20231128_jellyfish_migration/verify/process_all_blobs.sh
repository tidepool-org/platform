#!/bin/bash
BLOBS_DIR=~/Documents/tmp/blob_files/tandemCIQ100035490810069
USER_ID=6a452338-5064-4795-81ca-84957bad2280
USER_EMAIL=$1
USER_PW=$2

SERVER_TOKEN="$(curl -s -I -X POST -H "X-Tidepool-Server-Secret: $SERVER_SECRET" -H "X-Tidepool-Server-Name: devops" "https://${API_ENV}.tidepool.org/auth/serverlogin" | grep 'x-tidepool-session-token' | sed 's/[^:]*: //')"

for filename in $BLOBS_DIR/**/*blob.gz; do
    echo "$filename"
    source ./upload_blob.sh "$filename" "$USER_EMAIL" "$USER_PW"
    source ./cleanup_user_data.sh "$USER_ID" "$SERVER_TOKEN"
done
