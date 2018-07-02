GOCMD=go
GOGEN=$(GOCMD) generate
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOCOVER=$(GOCMD) tool cover

GO:=$(shell command -v go 2> /dev/null)
APT:=$(shell command -v apt-get 2> /dev/null)

## Reliant on go and $GOPATH being set.
.PHONY: check-go-env
check-go-env:
ifndef GO
	$(error go command is not installed or in PATH)
endif
ifndef GOPATH
	$(error GOPATH is not set)
endif

.PHONY: install-dep
install-dep: check-go-env ## Installs dep
	@mkdir -p $(GOPATH)/bin
	@curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

.PHONY: install-linter
install-linter: check-go-env ## Installs linter
	@$(GOGET) -u github.com/alecthomas/gometalinter
	@gometalinter --install
ifdef APT
	@sudo apt-get install golang-race-detector-runtime || true
endif

.PHONY: install-cover
install-cover: check-go-env ## Installs code coverage tool
	@$(GOGET) -u golang.org/x/tools/cmd/cover

.PHONY: setup
setup: check-go-env install-dep install-linter install-cover ## Sets up the tooling.

## gometalinter config in .gometalinter.json
.PHONY: lint
lint: ## Runs linting.
	@gometalinter ./...

.PHONY: build
build: ## Builds the code.
	$(GOBUILD) ./...

.PHONY: test
test: ## Runs unit tests and tests code coverage.
	@echo 'mode: atomic' > coverage.txt && $(GOTEST) -covermode=atomic -coverprofile=coverage.txt -v -race -timeout=30s ./...

.PHONY: cover
cover: test ## Runs unit tests, code coverage, and runs HTML coverage tool.
	@$(GOCOVER) -html=coverage.txt

.PHONY: clean
clean: ## go clean and clean up of artifacts.
	@$(GOCLEAN) ./... || true
	@rm coverage.txt || true

## Some magic from http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

