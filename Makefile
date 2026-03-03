# Makefile

TaskGoNode := "taskGoNode" 

PROJECT_BASE := $(shell pwd)
PROJECT_NODE := $(PROJECT_BASE)/node

EXE_DIR := $(PROJECT_BASE)/bin
EXE_NODE_PATH := $(EXE_DIR)/node

NODE_TARGET_FILE = $(PROJECT_NODE)/cmd/main.go
NODE_CONFIG_FILE = $(PROJECT_NODE)/conf/config.json

EXE_NODE := $(EXE_NODE_PATH)/$(TaskGoNode)

SH := $(SHELL)

build:
	@echo "build taskgo node"
	@go mod tidy
	@CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -v -o $(EXE_NODE) $(NODE_TARGET_FILE)

run_node_help:build
	@$(EXE_NODE) -h

run_node:build
	@$(EXE_NODE) -c $(NODE_CONFIG_FILE)

clean:
	@echo "clear node"
	@rm -rf $(EXE_NODE_PATH)

