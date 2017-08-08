## migrate_sessions_expand

This tool will migrate all existing sessions in the database to the longer form that includes additional fields such as isServer, serverId, userId, duration, createdAt, and expiresAt.

This tool *MUST* be executed immediately after upgrading `shoreline` to version `v0.9.1`.

To execute this tool against a local development setup:

1. Prepare the environment and build the executables. For more information, please see the main README.md.

  ```
  . ./env.sh
  make build
  ```

1. Run the tool against the Mongo database running on `localhost`.

  ```
  _bin/tools/migrate_sessions_expand/migrate_sessions_expand -a localhost
  ```

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_sessions_expand/migrate_sessions_expand.go","level":"info","line":297,"msg":"Deleted 15 expired sessions and migrated 32 sessions to expanded form","pid":27539,"process":"migrate_sessions_expand","time":"2016-09-23T16:18:41-07:00","version":"1.1.0+62861a2"}
  ```

  Note this log message reports `Deleted 15 expired sessions and migrated 32 sessions to expanded form`.

1. Run the tool again, just to ensure all sessions have been migrated.

  ```
  _bin/tools/migrate_sessions_expand/migrate_sessions_expand -a localhost
  ```

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_sessions_expand/migrate_sessions_expand.go","level":"info","line":297,"msg":"Deleted 0 expired sessions and migrated 0 sessions to expanded form","pid":27541,"process":"migrate_sessions_expand","time":"2016-09-23T16:19:34-07:00","version":"1.1.0+62861a2"}
  ```

  Note that this log message reports `Deleted 0 expired sessions and migrated 0 sessions to expanded form`. This means that all sessions were properly migrated during the previous step. This is the expected final response.
