# Platform

The Tidepool Platform API.

[![Build Status](https://travis-ci.org/tidepool-org/platform.png)](https://travis-ci.org/tidepool-org/platform)
[![Code Climate](https://codeclimate.com/github/tidepool-org/platform/badges/gpa.svg)](https://codeclimate.com/github/tidepool-org/platform)
[![Issue Count](https://codeclimate.com/github/tidepool-org/platform/badges/issue_count.svg)](https://codeclimate.com/github/tidepool-org/platform)

# Setup

1. Install Go version 1.6 or later.
1. Create a brand new Go directory.
1. Set the `GOPATH` environment variable to the newly created Go directory.
1. Add `$GOPATH/bin` to the `PATH` environment variable.
1. Execute `go get github.com/tidepool-org/platform` to pull down the project. You may ignore a "no buildable Go source files" warning.
1. Change directory to `$GOPATH/src/github.com/tidepool-org/platform`.
1. Source the `.env` file.
1. Execute `make editable` to install the various Go tools needed for building and editing the project.
1. If you are going to use the `test`, `ci-test`, or `watch` Makefile targets or the `ginkgo` executable, add `$GOPATH/src/github.com/tidepool-org/platform/Godeps/_workspace/bin` to the `PATH` environment variable. (Consider using the `direnv` tool via `homebrew` to automatically update the `GOPATH` and `PATH` environment variables when in certain directories).

For example:

```
brew install go
mkdir ~/go
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH
go get github.com/tidepool-org/platform
cd $GOPATH/src/github.com/tidepool-org/platform
source .env
make editable
```

For reuse, you may want to include the following lines in your shell config (e.g., .bashrc) or use a tool like [direnv](http://direnv.net/ 'direnv'):

```
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH
```

# Execute

1. Setup the environment, as above.
1. Build the project.
1. Execute the `dataservices` executable.

In addition to the setup above, for example:

```
make build
_bin/dataservices
```

Use `Ctrl-C` to stop the `dataservices` executable. It may take up to 60 seconds to stop.


# Makefile

* To setup your GO environment for building and editing the project:

```
make editable
```

* To build the executables:

```
make build
```

All executables are built in the `_bin` directory.

* To run all of the tests manually:

```
make test
```

The environment variable `TEST` indicates which test to execute. If not specified, then all tests are executed. For example,

```
TEST=user make test
```

* To run all of the tests automatically after any changes are made, in a separate terminal window:

```
make watch
```

The environment variable `WATCH` indicates which test to execute. If not specified, then all tests are executed. For example,

```
WATCH=user make watch
```

* To run `go fmt`, `goimports`, and `golint` all at once:

```
make pre-commit
```

* To clean the project of all build files:

```
make clean
```

* To add the required git hooks:

```
make git-hooks
```

# Sublime Text

If you use the Sublime Text editor with the GoSublime plugin, open the `platform.sublime-project` project to ensure the `GOPATH` and `PATH` environment variables are set correctly within Sublime Text. In addition, the recommended user settings are:

```
{
  "autocomplete_builtins": true,
  "autocomplete_closures": true,
  "autoinst": false,
  "fmt_cmd": [
    "goimports"
  ],
  "fmt_enabled": true,
  "fmt_tab_width": 4,
  "use_named_imports": true
}
```
