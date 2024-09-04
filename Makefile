APP 		?= $(shell basename `go list -m`)
VERSION		?= $(shell cat $(PWD)/.version 2> /dev/null || echo SNAPSHOT)
BUILT_AT 	?= $(shell date +'%Y-%m-%dT%T')

BUILD_DIR=.build
TARGET="$(BUILD_DIR)/${APP}-${OS}-${ARCH}"

.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

env: ## Print environment variables
	@echo APP=$(APP)
	@echo VERSION=$(VERSION)
	@echo BUILT_AT=$(BUILT_AT)

test: ## Run short tests
	go test -v ./... -short

test-it: ## Run all tests (including integration)
	go test -v ./...

build_arch:
	@echo Building for $(OS)/$(ARCH)
	env GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 \
    		go build \
    		-ldflags="-s -w -X main.version=$(VERSION) -X main.builtAt=$(BUILT_AT)" \
    		-o "$(BUILD_DIR)/$(APP)-$(OS)-$(ARCH)" ./cmd/loadinfo

build: ## Build application
	$(MAKE) OS=linux ARCH=amd64 build_arch
