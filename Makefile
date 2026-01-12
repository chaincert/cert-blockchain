#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)

# Application and binary names
APP_NAME = certd
DAEMON_NAME = certd

# Build tags
build_tags = netgo
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(APP_NAME) \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(DAEMON_NAME) \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

all: install lint test

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/$(DAEMON_NAME)

build:
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/$(DAEMON_NAME) ./cmd/$(DAEMON_NAME)

build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

go.sum: go.mod
	@echo "Ensuring dependencies have not been modified..."
	go mod verify
	go mod tidy

clean:
	rm -rf $(BUILDDIR)/*

###############################################################################
###                                 Testing                                 ###
###############################################################################

test:
	go test -mod=readonly -race ./...

test-unit:
	go test -mod=readonly ./x/...

test-integration:
	go test -mod=readonly -tags=integration ./tests/...

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run --out-format=tab

lint-fix:
	golangci-lint run --fix --out-format=tab --issues-exit-code=0

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs goimports -w -local github.com/chaincertify/certd

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-gen:
	@echo "Generating Protobuf files"
	@sh ./scripts/protocgen.sh

proto-lint:
	@buf lint --error-format=json

###############################################################################
###                                 Docker                                  ###
###############################################################################

docker-build:
	$(DOCKER) build -t cert-blockchain:latest .

docker-run:
	$(DOCKER) run -it --rm -p 26656:26656 -p 26657:26657 -p 1317:1317 -p 8545:8545 cert-blockchain:latest

###############################################################################
###                                  Init                                   ###
###############################################################################

init: build
	DAEMON_BIN=$(BUILDDIR)/$(DAEMON_NAME) ./scripts/init.sh

start:
	$(DAEMON_NAME) start --json-rpc.enable --json-rpc.api eth,txpool,personal,net,debug,web3

.PHONY: all install build build-linux clean test test-unit test-integration lint lint-fix format proto-gen proto-lint docker-build docker-run init start

