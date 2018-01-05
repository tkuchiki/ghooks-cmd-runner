# ghooks-cmd-runner
Receives Github webhooks and runs commands

## Installation

Download from https://github.com/tkuchiki/ghooks-cmd-runner/releases

## Usage

`-c, --config` is required.

```shell
$ ./ghooks-cmd-runner --help
usage: ghooks-cmd-runner --config=CONFIG [<flags>]

Receives Github webhooks and runs commands

Flags:
      --help              Show context-sensitive help (also try --help-long and --help-man).
  -c, --config=CONFIG     config file location
  -p, --port=18889        listen port
      --host="127.0.0.1"  listen host
  -l, --logfile=LOGFILE   log file location
      --pidfile=PIDFILE   pid file location
      --raw-payload       raw payload
      --version           Show application version.
```

### Config

```
# port = 18889 (default: 18889)
# host = "0.0.0.0 (default: 127.0.0.1)"
# secret = "your webhook secret"
# logfile = "path to logfile (default: stdout)"
# pidfile = "path to pidfile"

[[hook]]
event = "push"
command = "/path/to/script"
branch = "feature/*"
# exclude_actions = ["deleted", "created"]

[[hook]]
event = "pull_request"
command = "/path/to/script"
# call Status API (See: https://developer.github.com/v3/repos/statuses/#create-a-status)
access_token = "your access token"
include_actions = [ "opened", "reopened" ]
# exclude_actions = [ "closed", "unlabeled" ]
```

If you don't want to put secret in Config, alternatively you can store secret in .env file as follows:

```shell
SECRET_TOKEN=your webhook secret
```

### Script

```shell
# read from stdin
cat - | base64 -d | jq .
```

```shell
# change the target_url
echo http://example.com > ${SUCCESS_TARGET_FILE}
echo http://example.com > ${FAILURE_TARGET_FILE}
```

## Example

```shell
$ ./ghooks-cmd-runner -c /path/to/config --pidfile /path/to/pid -l /path/to/logfile
ghooks server start 127.0.0.1:18889
```

## TODO

- Unix Domain Socket
- Signal handling
