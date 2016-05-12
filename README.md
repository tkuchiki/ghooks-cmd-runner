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

[[hook]]
event = "pull_request"
command = "/path/to/script"
```

### Script

```shell
# output github webhook payload
echo ${GITHUB_WEBHOOK_PAYLOAD} | base64 -d | jq .
```

## Examples

```shell
$ ./ghooks-cmd-runner -c /path/to/config --pidfile /path/to/pid -l /path/to/logfile
ghooks server start 127.0.0.1:18889
```
