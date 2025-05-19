# Generate
- `go run *.go generate --help`

## Minimal
- `go run *.go generate --outputDir ./load/`

## All
- `go run *.go generate --count 15 --timeout 30 --actions failure:30000,sleep:1000,createWork --errors --futureSecs 120 --failureDuration 3600 --outputDir ./load/`

# Run 
- `go run *.go run --help`

## Minimal
- `go run *.go run --urlBase http://localhost:9220 --filePath ./load/items[100]-duration[0s]-offset[0s]-result-errs[false]-sys-failure[false].json`

## All
- `go run *.go run --urlBase http://localhost:9220 --filePath ./load/items[100]-duration[0s]-offset[0s]-result-errs[false]-sys-failure[false].json --duplicates --serialize --outputDir ./load/run/`

