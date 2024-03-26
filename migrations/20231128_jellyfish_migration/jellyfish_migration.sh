#!/bin/bash

go run jellyfish_migration.go --cap=500000 --nop-percent=1 --query-limit=100 --query-batch=50
