SHELL=bash

BUILD=build
BIN_DIR?=.

.PHONY: all
all: test build

.PHONY: build
build:
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build -o $(BUILD)/$(BIN_DIR)/food-recipes main.go

.PHONY: debug
debug:
	HUMAN_LOG=1 go run -race main.go

.PHONY: test
test:
	go test -race -cover ./...

.PHONEY: test build debug
