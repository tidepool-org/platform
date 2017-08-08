## tapi

This tool will allow easy command-line access to many features of the Tidepool API. It currently has limited features, but additional features will be added over time.

**NB:** If you would like to use this tool on the Tidepool platform running on your machine (i.e., localhost), please note the following additional instructions:

- make sure you're running the dataservice (for instructions see the repo [README](../../README.md), and don't forget to `. ./env.sh`)
- from the root of your platform repo directory, use the full path wherever you see `tapi` in the shell commands below (e.g., `$ _bin/tools/tapi/tapi login`)

## Which Server?

This tool must be configured with the appropriate HTTP endpoint in order to communicate with the Tidepool API. You can specify either the endpoint directly, or the Tidepool environment name and the endpoint will be determine automatically. It is recommended that you use specify the Tidepool environment name rather than the endpoint.

Note: The endpoint takes precedence over the environment name.

### Environment

If you choose to specify the Tidepool environment name when invoking the tool it will automatically convert from all known environments (ie. `prd`, `int`, `stg`, `dev`, `local`) to the appropriate endpoint. For example:

```
$ tapi login --env=prd
```

#### TIDEPOOL_ENV

Rather than specifying the `--env` argument on every command, you may define the `TIDEPOOL_ENV` environment variable. All subsequent invocations of the tool will use the environment specified in the `TIDEPOOL_ENV` environment variable. For example:

```
$ export TIDEPOOL_ENV=prd
$ tapi login
```

Note: The command line argument takes precedence over the environment variable.

### Endpoint

Alternatively, you may choose the specify the HTTP endpoint directly. For example:

```
tapi login --endpoint=https://api.tidepool.org
```

#### TIDEPOOL_ENDPOINT

Rather than specifying the `--endpoint` argument on every command, you may define the `TIDEPOOL_ENDPOINT` environment variable. All subsequent invocations of the tool will use the endpoint specified in the `TIDEPOOL_ENDPOINT` environment variable. For example:

```
$ export TIDEPOOL_ENDPOINT=https://api.tidepool.org
$ tapi login
```

Note: The command line argument takes precedence over the environment variable.

## Who Are You?

You must authenticate with the Tidepool API in order to use most features of the tool. Once authenticated, the Tidepool session will be stored for future invocations of the tool. (For the curious, the Tidepool session is persisted in `~/.tidepool/session`.)

### User

To authenticate as a user:

```
$ tapi login
Email: your-email@domain.com
Password: ********
Logged in.
```

Alternatively:

```
$ tapi login --email your-email@domain.com
Password: ********
Logged in.
```

### Server

To use the tool on a Tidepool server instance, authenticate with the server login:

```
$ tapi server-login
```

Note: The `TIDEPOOL_USER_CLIENT_SERVERTOKENSECRET` environment variable must be set as appropriate.

### Additional

To logout and destroy the Tidepool session:

```
$ tapi logout
Logged out.
```

To verify the Tidepool session with the Tidepool API:

```
$ tapi whoami
{"isserver":false,"userid":"1234567890"}
```

## Which User?

Many of the Tidepool APIs operate on a particular user, specified via id. Therefore, many of the tool commands include a `--user-id` argument.

If you are authenticated as a user, then this argument is optional. If not specified, the user id associated with the current Tidepool session will be used.

If you are authenticated on a Tidepool server instance, then this argument is required. A Tidepool server session is not associated with any specific user.

## Features

To keep the following examples as succinct as possible, it is assumed that the tool has been previously authenticated with a user account and the `--user-id` argument is not specified.

### Dataset

This tool can manage datasets.

#### List

To list all datasets for a user:

```
$ tapi dataset list
(... output ...)
```

##### Filter

You may provide limited filtering of the datasets.

###### Deleted

Datasets that have been previously deleted are not normally included in the list of datasets. If you wish to include deleted datasets in the list, then specify the `--deleted` argument. For example:

```
$ tapi dataset list --deleted
(... output ...)
```

##### Pagination

By default, the Tidepool API returns datasets in pages with a default and maximum pages size of 100. To effectively list all datasets you will need to specify the `--page` and `--size` arguments to interate through all of the available datasets.  For example:

```
$ tapi dataset list --size 50 --page 0
(... first 50 datasets ...)
$ tapi dataset list --size 50 --page 1
(... second 50 datasets ...)
```

The `--size` argument can be from 1 to 100. The `--page` argument can be zero or more.

The last page will either contain less than `size` number of datasets or will be empty.

#### Delete

To delete a specific dataset, determine its dataset id, also known as `uploadId`, via the list command above, and then invoke the delete dataset command. For example:

```
$ tapi dataset delete --dataset-id ff2346e5623914b1234565661f093459
Dataset deleted.
```

**NB:** Older upload IDs (from ingestion through the legacy "jellyfish" ingestion service) begin with `upid_` and contain only 12 characters in the hash.

## Help

For general help with the tool:

```
$ tapi --help
```

For more detailed help with a specific command:

```
$ tapi <command> --help
```

Replace \<command\> with one of the top-level commands.

For more detailed help with a specific sub-command:

```
$ tapi <command> <sub-command> --help
```

Replace \<command\> with one of the top-level commands and \<sub-command\> with a sub-command.

For example:

```
$ tapi dataset list --help
NAME:
   tapi dataset list - list datasets

USAGE:
   tapi dataset list [command options] [arguments...]

OPTIONS:
   --user-id USERID     USERID of the user to list datasets
   --deleted            include deleted datasets in the list
   --page PAGE          pagination PAGE (default: 0)
   --size SIZE          pagination SIZE (default: 0)
   --endpoint ENDPOINT  Tidepool API ENDPOINT (eg. 'https://api.tidepool.org') [$TIDEPOOL_ENDPOINT]
   --env ENVIRONMENT    Tidepool ENVIRONMENT (ie. 'prd', 'int', 'stg', 'dev', 'local') [$TIDEPOOL_ENV]
   --proxy URL          proxy URL [$HTTP_PROXY]
   --pretty, -p         pretty print JSON
   --verbose, -v        include info output
```
