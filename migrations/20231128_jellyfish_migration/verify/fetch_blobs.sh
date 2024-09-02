#!/bin/bash
JSON_FILE=$1
OUTPUT_DIR=$2
LOG_PREFIX=prod_blob_series

check_val() {
    if [[ -z "$1" ]]; then
        echo "missing $2 value"
        exit 2
    fi
}

## PRD
SECRET=$(op item get "PRD Server Secret" --account tidepool.1password.com --fields label=credential --format json | jq -r '.value')
API_ENV=api

## QA
# SECRET=$(op item get "qa3 server secret" --account tidepool.1password.com --fields label=credential --format json | jq -r '.value')
# API_ENV=qa2.development

check_val $JSON_FILE "JSON_FILE"
check_val $OUTPUT_DIR "OUTPUT_DIR"
check_val $SECRET "SECRET"

SESSION_TOKEN="$(curl -s -I -X POST -H "X-Tidepool-Server-Secret: $SECRET" -H "X-Tidepool-Server-Name: devops" "https://${API_ENV}.tidepool.org/auth/serverlogin" | grep 'x-tidepool-session-token' | sed 's/[^:]*: //')"

check_val $SESSION_TOKEN "SESSION_TOKEN"


## enbale downloading o
counter=0

jq -c ' reverse .[]' $JSON_FILE | while read i; do

    # counter=$((counter+1))

    # # if [[ "$counter" == 2 ]]; then
    # #     # reset counter
    # #     counter=0

        DEVICE_ID=$(jq -r '.deviceId' <<<"$i")
        check_val $DEVICE_ID "DEVICE_ID"

        BLOB_ID=$(jq -r '.blobId' <<<"$i")
        check_val $BLOB_ID "BLOB_ID"

        if [[ "$DEVICE_ID" =~ .*"tandem".* ]]; then
            OUTPUT_FILE="$OUTPUT_DIR/$DEVICE_ID/$BLOB_ID"_blob.gz
        else
            OUTPUT_FILE="$OUTPUT_DIR/$DEVICE_ID/$BLOB_ID"_blob.ibf
        fi

        # is already downloaded?
        if grep -wq "$OUTPUT_FILE" "${LOG_PREFIX}_upload.log" || grep -wq "$OUTPUT_FILE" "${LOG_PREFIX}_error.log"; then
            echo "$OUTPUT_FILE already downloaded"
        else
            
            mkdir -p "$OUTPUT_DIR/$DEVICE_ID"

            check_val $OUTPUT_FILE "OUTPUT_FILE"

        
            http_response=$(curl -s -o $OUTPUT_FILE -w "%{response_code}" --request GET \
                --url https://${API_ENV}.tidepool.org/v1/blobs/${BLOB_ID}/content \
                --header 'Accept: */*' \
                --header "X-Tidepool-Session-Token: $SESSION_TOKEN")

            if [ $http_response != "200" ]; then
                echo "$http_response error downloading blob $BLOB_ID for device $DEVICE_ID"
                rm -rf $OUTPUT_FILE
            else
                if [[ "$DEVICE_ID" =~ .*"tandem".* ]]; then
                    echo "status $http_response done downloading tandem blob $OUTPUT_FILE"
                else
                    gzip $OUTPUT_FILE
                    echo "status $http_response done downloading omnipod blob $OUTPUT_FILE"
                fi
            fi
        fi
    # fi
done
