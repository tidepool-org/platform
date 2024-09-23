#!/bin/bash
BLOB_FILE=$1
USER_EMAIL=$2
USER_PW=$3
OUTPUT_FILE_PREFIX=$4
UPLOADER_DIR=~/Documents/src/tidepool/uploader

SCRIPT=$(realpath "$0")
BASE_DIR=$(dirname "$SCRIPT")

check_val() {
    if [[ -z "$1" ]]; then
        echo "missing required '$2' value"
        exit 2
    fi
}

cd $UPLOADER_DIR
# config file exists in the uploader repo
source ./config/qa3.sh

check_val $BLOB_FILE "BLOB_FILE"
check_val $USER_EMAIL "USER_EMAIL"
check_val $USER_PW "USER_PW"
check_val $API_URL "API_URL"
check_val $BASE_DIR "BASE_DIR"
check_val $UPLOADER_DIR "UPLOADER_DIR"

start=$(date +%s)
SUCCESS=false
output='not yet run'

UPLOAD_LOG=_upload.log
ERROR_LOG=_error.log

if [[ -n "$OUTPUT_FILE_PREFIX" ]]; then
    UPLOAD_LOG="${BASE_DIR}/${OUTPUT_FILE_PREFIX}${UPLOAD_LOG}"
    ERROR_LOG="${BASE_DIR}/${OUTPUT_FILE_PREFIX}${ERROR_LOG}"
fi

if [[ "$BLOB_FILE" =~ .*"tandem".* ]]; then
    output=$(node -r @babel/register lib/drivers/tandem/cli/loader.js loader.js -f $BLOB_FILE -u $USER_EMAIL -p $USER_PW)
    echo "$output" | grep -q 'upload.toPlatform: all good' && SUCCESS=true
else
    output=$(node -r @babel/register lib/drivers/insulet/cli/ibf_loader.js ibf_loader.js -f $BLOB_FILE -u $USER_EMAIL -p $USER_PW)
    echo "$output" | grep -q 'upload.toPlatform: all good' && SUCCESS=true
fi

cd $BASE_DIR

end=$(date +%s)

if [ "$SUCCESS" = true ]; then
    echo 'upload all good'
    records=$(echo "$output" | grep -A100000 'attempting to upload' | grep -B100000 'device data records')
    runtime=$((end - start))
    echo "{"blob":"$BLOB_FILE", "account":"$USER_EMAIL", "time": "$runtime", "records": "$records" }" >>"$UPLOAD_LOG"
    echo "$records"
else
    echo 'upload failed!'
    error_details=$(echo "$output" | grep -A100000 'add data to dataset failed' | grep -B100000 'Offending record for error 0')

    if [[ -z "$error_details" ]]; then
        error_details=$(echo "$output" | grep -A100000 'add data to dataset failed' | grep -B100000 'upload.toPlatform: failed')
    fi

    if [[ -z "$error_details" ]]; then
        error_details=$(echo "$output")
    fi

    error_details=$(echo $error_details | tr -d '\n\t\r')

    echo "{"blob":"$BLOB_FILE", "details": {"$error_details"}" >>"$ERROR_LOG"
fi
