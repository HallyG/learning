PWD := $(shell pwd)
ROOT_PWD := $(shell git rev-parse --show-toplevel 2>/dev/null || pwd)
BUILD_DIR := ${PWD}/build

BUILD_VERSION := $(shell date '+%Y%m%d-%H%M')
BUILD_SHA := $(shell git rev-parse --verify HEAD)
BUILD_SHA_SHORT := $(shell git rev-parse --short=8 --verify HEAD)

APP_NAME=wal

GO_CMD ?= go
GO_BUILD_TAGS ?= 
GO_BUILD_LDFLAGS ?= -s -w -buildid= 

GO_MODULE_DIR := $(shell go list -m -f '{{.Dir}}')
GO_PKGS := $(shell go list -f '{{.Dir}}' ./... ) 
EXCLUDE_PKGS := $(shell go list -f '{{.Dir}}' ./... | grep -E '//migrations$$')

GO_TEST_PKGS := $(filter-out $(EXCLUDE_PKGS),$(GO_PKGS))
GO_COVERAGE_PKGS := $(filter-out $(EXCLUDE_PKGS),$(GO_PKGS))
GO_COVERAGE_FILE := $(BUILD_DIR)/cover.out
GO_COVERAGE_TEXT_FILE := $(BUILD_DIR)/cover.txt
GO_COVERAGE_HTML_FILE := $(BUILD_DIR)/cover.html

GOLANGCI_CMD := go tool golangci-lint
GOLANGCI_ARGS ?= --fix --concurrency=4

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | sort | column -t -s ':' |  sed -e 's/^/ /'

## clean: remove build artifacts
.PHONY: clean
clean:
	@echo "Cleaning binaries and test cache..."
	@if [ -n "${APP_NAME}" ]; then \
		rm -f ${BUILD_DIR}/${APP_NAME}; \
	fi
	@rm -f ${GO_COVERAGE_FILE} ${GO_COVERAGE_TEXT_FILE} ${GO_COVERAGE_HTML_FILE}
	@$(GO_CMD) clean

## test: run tests
.PHONY: test
test:
	@$(GO_CMD) test ${GO_BUILD_TAGS} -timeout 10s -race $(if $(VERBOSE),-v) ${GO_TEST_PKGS}

## test/cover: run tests with coverage
.PHONY: test/cover
test/cover:
	@mkdir -p ${BUILD_DIR}
	@rm -f ${GO_COVERAGE_FILE} ${GO_COVERAGE_TEXT_FILE} ${GO_COVERAGE_HTML_FILE}
	@$(GO_CMD) test ${GO_BUILD_TAGS} -timeout 10s -race -coverprofile=${GO_COVERAGE_FILE} ${GO_COVERAGE_PKGS}
	@$(GO_CMD) tool cover -func ${GO_COVERAGE_FILE} -o ${GO_COVERAGE_TEXT_FILE}
	@$(GO_CMD) tool cover -html ${GO_COVERAGE_FILE} -o ${GO_COVERAGE_HTML_FILE}

## lint: run golangci-lint
.PHONY: lint
lint:
	@$(GO_CMD) vet ${GO_PKGS}
	@${GOLANGCI_CMD} run ${GOLANGCI_ARGS} ${GO_PKGS}

## audit: format, vet, and lint Go code
.PHONY: audit
audit: clean lint
	@$(GO_CMD) mod tidy
	@$(GO_CMD) mod verify
	@$(GO_CMD) fmt ${GO_PKGS}

## build: build the application
.PHONY: build
build:
	$(GO_CMD) build ${GO_BUILD_TAGS} \
		-o ${BUILD_DIR}/${APP_NAME} \
		-trimpath -mod=readonly \
		-ldflags="${GO_BUILD_LDFLAGS} -X main.BuildSHA=${BUILD_SHA} -X main.BuildVersion=${BUILD_VERSION}" .