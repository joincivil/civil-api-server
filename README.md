# Civil API Server

This repository contains proprietary code for a GraphQL / REST API server used to return Civil related data.  Mainly used to serve up data from the Civil Events Crawler / Processor pipeline, but can be used for anything we need.

## Install Requirements

This project is using `make` to run setup, builds, tests, etc.

Ensure that your `$GOPATH` and `$GOROOT` are setup properly in your shell configuration and that this repo is cloned into the appropriate place in the `$GOPATH`. i.e. `$GOPATH/src/github.com/joincivil/civil-api-server/`

To setup the necessary requirements:

```
make setup
```

## Lint

Check all the packages for linting errors using a variety of linters via `golangci-lint`.  Check the `Makefile` for the up to date list of linters.

```
make lint
```

## Build


```
make build
```

## Testing

Runs the tests and checks code coverage across the project. Produces a `coverage.txt` file for use later.

```
make test
```

## Code Coverage Tool

Run `make test` and launches the HTML code coverage tool.

```
make cover
```
