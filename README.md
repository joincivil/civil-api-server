# Civil Events Processor

This repository contains proprietary code to process captured Civil contract events and persists the processed data down for use via APIs. It is written in `golang`. 

## Install Requirements

This project is using `make` to run setup, builds, tests, etc.  

Ensure that your `$GOPATH` and `$GOROOT` are setup properly in your shell configuration and that this repo is cloned into the appropriate place in the `$GOPATH`. i.e. `$GOPATH/src/github.com/joincivil/civil-events-processor/`

To setup the necessary requirements:

```
make setup
```

### Dependencies

Relies on `dep`[https://golang.github.io/dep/](https://golang.github.io/dep/) for dependency management, updating the `/vendor/` directory in the project.  

When adding and removing imports, make sure to run `dep ensure`.  Any adding or removing will require committing the updates on `Gopkg.lock` and `/vendor/` to the repository.


## Lint

Check all the packages for linting errors using a variety of linters via `gometalinter`.  Check the `Makefile` for the up to date list of linters.

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
