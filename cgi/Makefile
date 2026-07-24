APP_NAME := cgi
BUILD_VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short=8 --verify HEAD)

include ../common.mk

## run: run the application
.PHONY: run
run: build
	@${BUILD_DIR}/${APP_NAME}
