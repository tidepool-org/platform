## migrate_tokens_expand

This tool will migrate all existing tokens in the database to the longer form that includes additional fields such as isServer, serverId, userId, duration, createdAt, and expiresAt.

This tool *MUST* be executed immediately after upgrading `shoreline` to version `v0.9.1`.

To execute this tool against a local development setup:

1. Prepare the environment and build the executables. For more information, please see the main README.md.

  ```
  source .env
  make build
  ```

1. Run the tool against the Mongo database running on `localhost`.

  ```
  _bin/tools/migrate_tokens_expand/migrate_tokens_expand -a localhost
  ```

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_tokens_expand/migrate_tokens_expand.go","level":"info","line":297,"msg":"Deleted 15 expired tokens and migrated 32 tokens to expanded form","pid":27539,"process":"migrate_tokens_expand","time":"2016-09-23T16:18:41-07:00","version":"1.1.0+62861a2"}
  ```

  Note this log message reports `Deleted 15 expired tokens and migrated 32 tokens to expanded form`.

1. Run the tool again, just to ensure all tokens have been migrated.

  ```
  _bin/tools/migrate_tokens_expand/migrate_tokens_expand -a localhost
  ```

  There should be a final log message along the lines of:

  ```
  {"file":"tools/migrate_tokens_expand/migrate_tokens_expand.go","level":"info","line":297,"msg":"Deleted 0 expired tokens and migrated 0 tokens to expanded form","pid":27541,"process":"migrate_tokens_expand","time":"2016-09-23T16:19:34-07:00","version":"1.1.0+62861a2"}
  ```

  Note that this log message reports `Deleted 0 expired tokens and migrated 0 tokens to expanded form`. This means that all tokens were properly migrated during the previous step. This is the expected final response.
