#!/bin/bash

INPUT_FILE=$1
OUTPUT_FILE=$2
EXCLUDE_TXT=$3
TMP_FILE=tmp.json

echo "input_file: $INPUT_FILE"
echo "output_file: $OUTPUT_FILE"
echo "exclusion: $EXCLUDE_TXT"

# move to json array
jq -cnr '(reduce inputs as $line ([]; . + [$line]))' $INPUT_FILE >$TMP_FILE

# iterate json and filter out known error
jq -c ".[] | select(.error, ._id | .detail | contains($EXCLUDE_TXT) | not ) | {detail: .error.detail, id: ._id}" $TMP_FILE >$OUTPUT_FILE

rm $TMP_FILE
