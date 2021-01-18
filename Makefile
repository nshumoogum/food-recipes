SHELL=bash

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

.PHONY: all
all: test build

.PHONY: build
build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/food-recipes main.go

.PHONY: debug
debug:
	HUMAN_LOG=1 go run -race main.go

.PHONY: test
test:
	go test -race -cover ./...

.PHONEY: test build debug
