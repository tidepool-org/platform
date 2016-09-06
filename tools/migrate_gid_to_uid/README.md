## migrate_gid_to_uid

This tool will migrate all device data with a _groupId field and add a _userId field containing the associated user id. It should be executed immediately after upgrading to any version of the `platform` *after* `v0.1.0`.

To execute this tool against a local development setup:

1. Prepare the environment and build the executables. For more information, please see the main README.md.

  ```
  source .env
  make build
  ```

1. Run the tool against the Mongo database running on `localhost`.

  ```
  _bin/tools/migrate_gid_to_uid/migrate_gid_to_uid -a localhost
  ```

  There may be a number of warnings if the local development setup has partially created accounts and/or very old device data.

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_gid_to_uid/migrate_gid_to_uid.go","level":"info","line":455,"msg":"Migrated 13538 device data for 4 groups","pid":23991,"process":"migrate_gid_to_uid","time":"2016-09-06T13:45:58-07:00","version":"0.1.3+88d5568"}
  ```

  Note this log message reports `Migrated 13538 device data for 4 groups`. This means that the device data for 4 users was properly migrated.

1. Run the tool again, just to ensure all data has been migrated.

  ```
  _bin/tools/migrate_gid_to_uid/migrate_gid_to_uid -a localhost
  ```

  Again, there may be a number of warnings if the local development setup has partially created accounts and/or very old data.

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_gid_to_uid/migrate_gid_to_uid.go","level":"info","line":455,"msg":"Migrated 0 device data for 0 groups","pid":23991,"process":"migrate_gid_to_uid","time":"2016-09-06T13:45:58-07:00","version":"0.1.3+88d5568"}
  ```

  Note that this log message reports `Migrated 0 device data for 0 groups`. This means that all device data was properly migrated during the previous step. This is the expected final response.
