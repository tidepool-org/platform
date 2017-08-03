## permission_gid

This tool will encode a user id to permission group id or decode a permission group id to user id.

To execute this tool against a local development setup:

1. Prepare the environment and build the executables. For more information, please see the main README.md.

  ```
  . ./env.local.sh
  make build
  ```

1. To encode a user id to permission group id:

  ```
  _bin/tools/permission_gid/permission_gid -a localhost --secret <SECRET> --encode <USER_ID>
  ```

1. To decode a permission group id to user id:

  ```
  _bin/tools/permission_gid/permission_gid -a localhost --secret <SECRET> --decode <PERMISSION_GROUP_ID>
  ```
