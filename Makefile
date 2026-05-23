#!/usr/bin/make -f

export VERSION := $(shell echo $(shell git describe --always --match "v*") | sed 's/^v//')
export COMMIT := $(shell git log -1 --format='%H')
export COMETBFT_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')

BIN_DIR ?= $(GOPATH)/bin
BUILD_DIR ?= $(CURDIR)/build
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)
HTTPS_GIT := https://github.com/agex-labs/orynx.git
DOCKER := $(shell which docker)
DOCKER_COMPOSE := $(shell which docker-compose)
HOMEDIR ?= $(CURDIR)
GENESIS ?= $(HOMEDIR)/config/genesis.json
GENESIS_TMP ?= $(HOMEDIR)/config/genesis_tmp.json
APP_TOML ?= $(HOMEDIR)/config/app.toml
CONFIG_TOML ?= $(HOMEDIR)/config/config.toml
COVER_FILE ?= cover.out
BENCHMARK_ITERS ?= 10
USE_CORE_MARKETS ?= true
USE_RAYDIUM_MARKETS ?= false
USE_UNISWAPV3_BASE_MARKETS ?= false
USE_COINGECKO_MARKETS ?= false
USE_COINMARKETCAP_MARKETS ?= false
USE_OSMOSIS_MARKETS ?= false
USE_POLYMARKET_MARKETS ?= false
SCRIPT_DIR := $(CURDIR)/scripts
DEV_COMPOSE ?= $(CURDIR)/contrib/compose/docker-compose-dev.yml

LEVANT_VAR_FILE:=$(shell mktemp -d)/levant.yaml
NOMAD_FILE_ORYNX:=contrib/nomad/orynx.nomad

TAG := $(shell git describe --tags --always --dirty)

export HOMEDIR := $(HOMEDIR)
export APP_TOML := $(APP_TOML)
export GENESIS := $(GENESIS)
export GENESIS_TMP := $(GENESIS_TMP)
export USE_CORE_MARKETS ?= $(USE_CORE_MARKETS)
export USE_RAYDIUM_MARKETS ?= $(USE_RAYDIUM_MARKETS)
export USE_UNISWAPV3_BASE_MARKETS ?= $(USE_UNISWAPV3_BASE_MARKETS)
export USE_COINGECKO_MARKETS ?= $(USE_COINGECKO_MARKETS)
export USE_COINMARKETCAP_MARKETS ?= $(USE_COINMARKETCAP_MARKETS)
export USE_OSMOSIS_MARKETS ?= $(USE_OSMOSIS_MARKETS)
export USE_POLYMARKET_MARKETS ?= $(USE_POLYMARKET_MARKETS)
export SCRIPT_DIR := $(SCRIPT_DIR)

BUILD_TAGS := -X github.com/agex-labs/orynx/cmd/build.Build=$(TAG)

###############################################################################
###                               build                                     ###
###############################################################################

build: tidy
	go build -ldflags="$(BUILD_TAGS)" \
	 -o ./build/ ./...

run-oracle-client: build
	@./build/client --host localhost --port 8080

start-all-dev:
	@echo "Starting development oracle side-car, blockchain, grafana, and prometheus dashboard..."
	@$(DOCKER_COMPOSE) -f $(DEV_COMPOSE) --profile all up -d --build

stop-all-dev:
	@echo "Stopping development network..."
	@$(DOCKER_COMPOSE) -f $(DEV_COMPOSE) --profile all down

start-sidecar-dev:
	@echo "Starting development oracle side-car, grafana, and prometheus dashboard..."
	@$(DOCKER_COMPOSE) -f $(DEV_COMPOSE) --profile sidecar up -d --build

stop-sidecar-dev:
	@echo "Stopping development oracle..."
	@$(DOCKER_COMPOSE) -f $(DEV_COMPOSE) --profile sidecar down

install: tidy
	@go install -ldflags="$(BUILD_TAGS)" -mod=readonly ./cmd/orynx

.PHONY: build install run-oracle-client start-all-dev stop-all-dev

###############################################################################
##                                  Docker                                   ##
###############################################################################

docker-build:
	@echo "Building E2E Docker image..."
	@DOCKER_BUILDKIT=1 $(DOCKER) build -t agex-labs/orynx-e2e -f contrib/images/orynx.e2e.Dockerfile .
	@DOCKER_BUILDKIT=1 $(DOCKER) build -t agex-labs/orynx-e2e-oracle -f contrib/images/orynx.sidecar.dev.Dockerfile .

.PHONY: docker-build


###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: tidy proto-format proto-gen proto-pulsar-gen format

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-pulsar-gen:
	@echo "Generating Dep-Inj Protobuf files"
	@$(protoImage) sh ./scripts/protocgen-pulsar.sh

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	@$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update

.PHONY: proto-all proto-gen proto-pulsar-gen proto-format proto-lint proto-check-breaking proto-update-deps


###############################################################################
###                              Formatting                                 ###
###############################################################################

tidy:
	@go mod tidy

.PHONY: tidy

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --out-format=tab

lint-fix:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --out-format=tab --issues-exit-code=0

lint-markdown:
	@echo "--> Running markdown linter"
	@markdownlint **/*.md

govulncheck:
	@echo "--> Running govulncheck"
	@go run golang.org/x/vuln/cmd/govulncheck -test ./...

.PHONY: lint lint-fix lint-markdown govulncheck

###############################################################################
###                                Mocks                                    ###
###############################################################################

mocks: gen-mocks format

gen-mocks:
	@echo "--> generating mocks"
	@go install github.com/vektra/mockery/v2
	@go generate ./...

###############################################################################
###                                Formatting                               ###
###############################################################################

format:
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "*/mocks/*" -not -name '*.pb.go' -not -name '*.pulsar.go' -not -name '*.gw.go' | xargs go run mvdan.cc/gofumpt -w .
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "*/mocks/*" -not -name '*.pb.go' -not -name '*.pulsar.go' -not -name '*.gw.go' | xargs go run github.com/client9/misspell/cmd/misspell -w
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "/*mocks/*" -not -name '*.pb.go' -not -name '*.pulsar.go' -not -name '*.gw.go' | xargs go run golang.org/x/tools/cmd/goimports -w -local github.com/agex-labs/orynx

.PHONY: format

###############################################################################
###                                dev-deploy                               ###
###############################################################################

deploy-dev:
	@touch ${LEVANT_VAR_FILE}
	@yq e -i '.sidecar_image |= "${SIDECAR_IMAGE}"' ${LEVANT_VAR_FILE}
	@yq e -i '.chain_image |= "${CHAIN_IMAGE}"' ${LEVANT_VAR_FILE}
	@levant deploy -force -force-count -var-file=${LEVANT_VAR_FILE} ${NOMAD_FILE_ORYNX}

.PHONY: deploy-dev

###############################################################################
##                                  Docs                                     ##
###############################################################################

docs:
	@if which mintlify > /dev/null 2>&1; then \
		echo "Starting Mintlify dev server..."; \
		cd docs && mintlify dev; \
	else \
		echo "Mintlify not found. Install at https://mintlify.com/docs/quickstart"; \
	fi
.PHONY: docs
