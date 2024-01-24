# Running jellyfish migration tool

## login to server with mongo access

## clone platform repo

## set `uri` for migration too
- go to `platform/cd migrations/20231128_jellyfish_migration/`
- create file `uri`
- add single line to file with mongo connection string `mongodb+srv://<secretdetail>/?retryWrites=true&w=majority`

## run tool
- help:
```
GLOBAL OPTIONS:
   --dry-run, -n        dry run only; do not migrate
   --stop-error         stop migration on error
   --audit              run audit
   --cap value          max number of records migrate (default: 0)
   --nop-percent value  how much of the oplog is NOP (default: 50)
   --uri value          mongo connection URI [./uri]
   --datum-id value     id of last datum updated [./lastUpdatedId]
   --user-id value      id of single user to migrate
   --query-limit value  max number of items to return (default: 50000)
   --query-batch value  max number of items in each query batch (default: 10000)
   --help, -h           show help

```
- test migration for a user:
    `go run jellyfish_migration.go --stop-error --dry-run --user-id=924edd2e-b8fc-45ad-b3f4-3032bb6b0a45`

- run migration:
    `go run jellyfish_migration.go --user-id=924edd2e-b8fc-45ad-b3f4-3032bb6b0a45`

