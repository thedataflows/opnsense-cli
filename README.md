# OPNSense CLI

[OPNSense](https://opnsense.org/) client using [exposed API](https://docs.opnsense.org/development/api.html)

## Setup

TODO

## Run It 🏃

`go run main.go`

## Usage

- `opnsense-cli`

    ```properties
    OPNSense command line interface

    All flags values can be provided via env vars starting with OSCLI_*
    To pass a command (e.g. 'command1') flag, use OSCLI_COMMAND1_FLAGNAME=somevalue

    Usage:
      opnsense-cli [flags]
      opnsense-cli [command]

    Available Commands:
      completion  Generate the autocompletion script for the specified shell
      help        Help about any command
      raw         Call OPNSense Raw API command
      version     Display version and exit

    Flags:
          --config strings           Config file(s) or directories. When just dirs, file 'main' with extensions 'json, toml, yaml, yml, properties, props, prop, hcl, tfvars, dotenv, env, ini' is looked up.   Can be specified multiple times (default [.,C:\Users\cri\AppData\Roaming\main])
      -h, --help                     help for opnsense-cli
          --log-format string        Set log format to one of: 'console, json' (default "console")
          --log-level string         Set log level to one of: 'trace, debug, info, warn, error, fatal, panic, disabled' (default "info")
          --opnsense-key string      OPNSense Key. See https://docs.opnsense.org/development/api.html#introduction
          --opnsense-secret string   OPNSense Secret
          --opnsense-url string      OPNSense URL (default "https://opnsense.local")
          --opnsense-url-insecure    OPNSense URL is Insecure

    Use "opnsense-cli [command] --help" for more information about a command.
    ```

- `opnsense-cli raw`

    ```properties
    Call OPNSense Raw API command

    Usage:
      opnsense-cli raw [flags]
      opnsense-cli raw [command]

    Aliases:
      raw, r

    Module: captiveportal, Controller: access
      captiveportal/access/logoff                  Method: GET, Arguments: [$zoneid=0]
      captiveportal/access/logon                   Method: POST, Arguments: [$zoneid=0]
      captiveportal/access/status                  Method: POST, Arguments: [$zoneid=0]

    Module: captiveportal, Controller: service
      captiveportal/service/delTemplate            Method: POST, Arguments: [$uuid]
      captiveportal/service/getTemplate            Method: GET, Arguments: [$fileid=null]
      captiveportal/service/reconfigure            Method: POST
      captiveportal/service/saveTemplate           Method: POST
      captiveportal/service/searchTemplates        Method: GET

      ...........
    ```

## Configure It ☑️

- See [sample/myconfig.yaml](./sample/myconfig.yaml) for config file
- All parameters can be set via flags or env as well: `OSCLI_<subcommand>_<flag>`, example: `OSCLI_OPNSENSE_SECRET=1122334455`

## Test It 🧪

Test for coverage and race conditions

`make coverage`

## Lint It 👕

`make pre-commit run`

## Roadmap

- [ ] ?

## Development

### Build

- Preferably: `goreleaser build --clean --single-target` or
- `make build` or
- `scripts/local-build.sh` (deprecated)
