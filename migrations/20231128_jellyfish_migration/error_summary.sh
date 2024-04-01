#!/bin/bash

INPUT_FILE=$1

echo "input_file: $INPUT_FILE"
parts=($(echo $INPUT_FILE | cut -d '.' -f1))

OUTPUT_FILE="${parts[0]}_summary.log"
echo "output_file: $OUTPUT_FILE"

cat $INPUT_FILE | jq '.error.detail' | sort | uniq -c | sort -nr >$OUTPUT_FILE
