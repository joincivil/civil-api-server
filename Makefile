GOCMD=go
GOGEN=$(GOCMD) generate
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOCOVER=$(GOCMD) tool cover
ABIGEN=abigen

GO:=$(shell command -v go 2> /dev/null)
APT:=$(shell command -v apt-get 2> /dev/null)

ABI_DIR=abi

## List of expected dirs for generated code
GENERATED_DIR=pkg/generated
GENERATED_CONTRACT_DIR=$(GENERATED_DIR)/contract
GENERATED_EVENT_LIST_DIR=$(GENERATED_DIR)/eventlist

EVENTLIST_GEN_MAIN=cmd/eventlistgen/main.go

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

.PHONY: install-abigen
install-abigen: check-go-env ## Installs the Ethereum abigen tool
	@$(GOGET) -u github.com/ethereum/go-ethereum/cmd/abigen

.PHONY: setup
setup: check-go-env install-dep install-linter install-cover install-abigen ## Sets up the tooling.

.PHONY: generate-event-lists
generate-event-lists: ## Runs eventlistgen to generate event list validators.
	@mkdir -p $(GENERATED_EVENT_LIST_DIR)
	@$(GORUN) $(EVENTLIST_GEN_MAIN) eventlist > ./$(GENERATED_EVENT_LIST_DIR)/eventlist.go

.PHONY: generate-contracts
generate-contracts: ## Builds the contract wrapper code from the ABIs in /abi.
ifneq ("$(wildcard $(ABI_DIR)/*.abi)", "")
	@mkdir -p $(GENERATED_CONTRACT_DIR)
	@$(ABIGEN) -abi ./$(ABI_DIR)/CivilTCR.abi -bin ./$(ABI_DIR)/CivilTCR.bin -type CivilTCRContract -out ./$(GENERATED_CONTRACT_DIR)/CivilTCRContract.go -pkg contract
	@$(ABIGEN) -abi ./$(ABI_DIR)/Newsroom.abi -bin ./$(ABI_DIR)/Newsroom.bin -type NewsroomContract -out ./$(GENERATED_CONTRACT_DIR)/NewsroomContract.go -pkg contract
	@$(ABIGEN) -abi ./$(ABI_DIR)/PLCRVoting.abi -bin ./$(ABI_DIR)/PLCRVoting.bin -type PLCRVotingContract -out ./$(GENERATED_CONTRACT_DIR)/PLCRVotingContract.go -pkg contract
	@$(ABIGEN) -abi ./$(ABI_DIR)/Parameterizer.abi -bin ./$(ABI_DIR)/Parameterizer.bin -type ParameterizerContract -out ./$(GENERATED_CONTRACT_DIR)/ParameterizerContract.go -pkg contract
	@$(ABIGEN) -abi ./$(ABI_DIR)/Government.abi -bin ./$(ABI_DIR)/Government.bin -type GovernmentContract -out ./$(GENERATED_CONTRACT_DIR)/GovernmentContract.go -pkg contract
	@$(ABIGEN) -abi ./$(ABI_DIR)/EIP20.abi -bin ./$(ABI_DIR)/EIP20.bin -type EIP20Contract -out ./$(GENERATED_CONTRACT_DIR)/EIP20.go -pkg contract
else
	$(error No abi files found; copy them to /abi after generation)
endif

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

