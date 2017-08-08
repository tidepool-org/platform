## migrate_pmid_to_uid

This tool will migrate all Seagull metadata using the _id field and add a userId field containing the associated user id.

This tool *MUST* be executed immediately before and immediately after upgrading `seagull` to version `v0.3.1`.

To execute this tool against a local development setup:

1. Prepare the environment and build the executables. For more information, please see the main README.md.

  ```
  . ./env.sh
  make build
  ```

1. Run the tool against the Mongo database running on `localhost`.

  ```
  _bin/tools/migrate_pmid_to_uid/migrate_pmid_to_uid -a localhost
  ```

  There may be a number of warnings if the local development setup has partially created accounts and/or old data.

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_pmid_to_uid/migrate_pmid_to_uid.go","level":"info","line":455,"msg":"Migrated 156 metadata for 156 meta","pid":23991,"process":"migrate_pmid_to_uid","time":"2016-09-06T13:45:58-07:00","version":"0.1.3+88d5568"}
  ```

  Note this log message reports `Migrated 156 metadata for 156 meta`. This means that the metadata for 156 users was properly migrated.

1. Run the tool again, just to ensure all data has been migrated.

  ```
  _bin/tools/migrate_pmid_to_uid/migrate_pmid_to_uid -a localhost
  ```

  Again, there may be a number of warnings if the local development setup has partially created accounts and/or old data.

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_pmid_to_uid/migrate_pmid_to_uid.go","level":"info","line":455,"msg":"Migrated 0 metadata for 0 meta","pid":23991,"process":"migrate_pmid_to_uid","time":"2016-09-06T13:45:58-07:00","version":"0.1.3+88d5568"}
  ```

  Note that this log message reports `Migrated 0 metadata for 0 meta`. This means that all metadata was properly migrated during the previous step. This is the expected final response.
