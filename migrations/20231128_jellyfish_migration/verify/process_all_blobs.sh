#!/bin/bash

USER_ID=6a452338-5064-4795-81ca-84957bad2280
USER_EMAIL=jamie+platform_upload@tidepool.org
USER_PW=$1
LOG_PREFIX=qa_blob

DEVICE_TYPE=$2
if [[ -z "$DEVICE_TYPE" ]]; then
    DEVICE_TYPE=tandem
    USER_ID=04afa5f8-9cb4-4824-9d76-67fa8740da2b
    USER_EMAIL=jamie+jellyfish_upload@tidepool.org
fi

BLOBS_DIR=~/Documents/tmp/prd_blobs/${DEVICE_TYPE}
SECRET=$(op item get "qa3 server secret" --account tidepool.1password.com --fields label=credential --format json | jq -r '.value')

for filename in $BLOBS_DIR*/*blob*.gz; do

    if grep -wq "$filename" "${LOG_PREFIX}_upload.log"; then
        echo "$filename already uploaded so cleaning up"
        file_path=$(echo  $filename  | rev | cut -d"/" -f2-  | rev)
        rm -rf "$file_path"
        echo "$file_path removed"
    elif grep -wq "$filename" "${LOG_PREFIX}_error.log"; then
        echo "$filename already failed to upload"
    else
        SERVER_TOKEN="$(curl -s -I -X POST -H "X-Tidepool-Server-Secret: $SECRET" -H "X-Tidepool-Server-Name: devops" "https://${API_ENV}.tidepool.org/auth/serverlogin" | grep 'x-tidepool-session-token' | sed 's/[^:]*: //')"
        source ./upload_blob.sh "$filename" "$USER_EMAIL" "$USER_PW" "$LOG_PREFIX"
        source ./cleanup_user_data.sh "$USER_ID" "$SERVER_TOKEN"
    fi
done
