# Makefile

TaskGoNode := "taskGoNode" 
TaskGoAdmin := "taskGoAdmin"

PROJECT_BASE := $(shell pwd)
PROJECT_NODE := "$(PROJECT_BASE)/node"
PROJECT_ADMIN := "$(PROJECT_BASE)/admin"

EXE_DIR := $(PROJECT_BASE)/bin
EXE_NODE_PATH := $(EXE_DIR)/node
EXE_ADMIN_PATH := $(EXE_DIR)/admin

NODE_TARGET_FILE = $(PROJECT_NODE)/cmd/main.go
NODE_CONFIG_FILE = $(PROJECT_NODE)/conf/config.json

ADMIN_TARGET_FILE = $(PROJECT_ADMIN)/cmd/main.go
ADMIN_CONFIG_FILE = $(PROJECT_ADMIN)/conf/config.json

EXE_NODE := $(EXE_NODE_PATH)/$(TaskGoNode)
EXE_ADMIN := $(EXE_ADMIN_PATH)/$(TaskGoAdmin)

build:
	@echo "build taskgo node"
	@go mod tidy
	@CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -v -o $(EXE_NODE) $(NODE_TARGET_FILE)
	@echo "build taskgo admin"
	@CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -v -o $(EXE_ADMIN) $(ADMIN_TARGET_FILE)

run_node_help:build
	@$(EXE_NODE) -h

run_node:build
	@$(EXE_NODE) -c $(NODE_CONFIG_FILE)

run_admin_help:build
	@$(EXE_ADMIN) -h

run_admin:build
	@$(EXE_ADMIN) -c $(ADMIN_CONFIG_FILE)

clean:
	@echo "clear node"
	@rm -rf $(EXE_NODE_PATH)
	@echo "clear admin"
	@rm -rf $(EXE_ADMIN_PATH)

