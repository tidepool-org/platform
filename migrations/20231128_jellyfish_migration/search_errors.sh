#!/bin/bash

INPUT_FILE=$1
OUTPUT_FILE=$2
EXCLUDE_TXT=$3
TMP_FILE=tmp.json

echo "input_file: $INPUT_FILE"
echo "output_file: $OUTPUT_FILE"
echo "exclude error code: $EXCLUDE_TXT"

jq -cnr '(reduce inputs as $line ([]; . + [$line]))' $INPUT_FILE >$TMP_FILE

if [[ -z "$EXCLUDE_TXT" ]]; then
    jq -c "map(.)|unique_by(.error.detail)|.[]|{"id":._id,"detail":.error.detail,"code":.error.code, "source":.error.source}" $TMP_FILE >$OUTPUT_FILE
else
    jq -c "map(.)|unique_by(.error.detail)|.[]|select(.error.detail!=null)|select(.error.detail|contains(\"$EXCLUDE_TXT\")|not)|.[]|{"id":._id,"detail":.error.detail,"code":.error.code, "source":.error.source}" $TMP_FILE >$OUTPUT_FILE
fi
