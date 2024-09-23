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
   --stop-on-error      stop migration on error
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
    `go run jellyfish_migration.go --user-id=924edd2e-b8fc-45ad-b3f4-3032bb6b0a45 --stop-error --dry-run`

- run migration:
    `go run jellyfish_migration.go --user-id=924edd2e-b8fc-45ad-b3f4-3032bb6b0a45`


- finding upload records with blobs

```
[
  { "$match": { "deviceManufacturers": { "$in": ["Tandem", "Insulet"] }, "client.private.blobId": { "$exists": true }}},
  { "$project": { "blobId": "$client.private.blobId", "_userId": 1, "deviceId": 1}},
  { "$group": { "_id": "$_userId", "detail": { "$push": "$$ROOT" }}}
]
```

- finding error types

grep -c "InsOmn.*Checksum" prod_blob_error.log
grep -c "InsOmn.*rawdata" prod_blob_error.log
grep -c "InsOmn.*value-not-exists" prod_blob_error.log
grep -c "InsOmn.*value-out-of-range" prod_blob_error.log

grep -c "tandem" prod_blob_error.log


cat prod_blob_upload.log | grep "InsOmn" | sed -n 's/.*records: \([0-9]*\).*/\1/p' | awk '{sum+=$1} END {print sum}'
cat prod_blob_upload.log | grep "InsOmn" | sed -n 's/.*upload \([0-9]*\).*/\1/p' | awk '{sum+=$1} END {print sum}'

cat prod_blob_upload.log | grep "tandem" | sed -n 's/.*records: \([0-9]*\).*/\1/p' | awk '{sum+=$1} END {print sum}'
cat prod_blob_upload.log | grep "tandem" | sed -n 's/.*upload \([0-9]*\).*/\1/p' | awk '{sum+=$1} END {print sum}'



