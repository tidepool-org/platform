#!/bin/bash

INPUT_FILE=$1

echo "input_file: $INPUT_FILE"
echo $INPUT_FILE | cut -d'.' -f1 | read OUTPUT

echo "output_file: ${OUTPUT}_summary.log"

cat $INPUT_FILE | jq '.error.detail' | sort | uniq -c | sort -nr >"${OUTPUT}_summary.log"
