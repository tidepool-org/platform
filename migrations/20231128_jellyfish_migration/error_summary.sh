#!/bin/bash

INPUT_FILE=$1
FIND_VAL=$2

echo "input_file: $INPUT_FILE"
parts=($(echo $INPUT_FILE | cut -d '.' -f1))

if [[ -z "$FIND_VAL" ]]; then
    OUTPUT_FILE="${parts[0]}_summary.log"
    echo "output_file: $OUTPUT_FILE"
    cat $INPUT_FILE | jq '.error.detail // .error.errors[]?.detail' | sort | uniq -c | sort -nr >$OUTPUT_FILE
else
    OUTPUT_FILE="${parts[0]}_detail.log"
    echo "output_file: $OUTPUT_FILE"
    cat $INPUT_FILE | jq -c "select(.error.detail == \"$FIND_VAL\") // select(.error.errors[]?.detail == \"$FIND_VAL\") \
                | if .error.code != null then \
                { id:._id, error: .error.code, detail: .error.detail, source: .error.source } else \
                { id:._id, source: .error.errors[].source } end \
                " >$OUTPUT_FILE
fi
