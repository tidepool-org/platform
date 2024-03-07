#!/bin/bash

INPUT_FILE=$1
OUTPUT_FILE=$2
EXCLUDE_TXT=$3
TMP_FILE=tmp.json

echo "input_file: $INPUT_FILE"
echo "output_file: $OUTPUT_FILE"
echo "exclude error code: $EXCLUDE_TXT"

if [[ -z "$EXCLUDE_TXT" ]]; then  
    jq -c "map(.error)|unique_by(.code)|.[]" $TMP_FILE >$OUTPUT_FILE
else
    jq -c "map(.error)|unique_by(.code)|.[]|select(.code!=null)|select(.code|contains(\"$EXCLUDE_TXT\")|not)" $TMP_FILE >$OUTPUT_FILE
fi

rm $TMP_FILE
