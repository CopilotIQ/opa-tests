# Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
#
# Created by M. Massenzio, 2022-06-23

GOOS ?= $(shell uname -s | tr "[:upper:]" "[:lower:]")
GOARCH ?= amd64

version := v0.1.0
release := $(version)-g$(shell git rev-parse --short HEAD)
prog := opatest
bin := out/bin/$(prog)-$(version)_$(GOOS)-$(GOARCH)

# Source files & Test files definitions
#
# Edit only the packages list, when adding new functionality,
# the rest is deduced automatically.
#
pkgs := ./testing ./testing/internals
all_go := $(shell for d in $(pkgs); do find $$d -name "*.go"; done)
test_srcs := $(shell for d in $(pkgs); do find $$d -name "*_test.go"; done)
srcs := $(filter-out $(test_srcs),$(all_go))

##@ Development

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: version
version: ## Emits the current software version
	@echo $(version)

.PHONY: version
release: ## Generates a release tag, based on the software version and commit SHA
	@echo $(release)

$(bin): cmd/main.go $(srcs)

build: $(bin)  ## Build the opatest binary in the out/bin directory.
	@mkdir -p $(shell dirname $(bin))
	GOOS=$(GOOS); GOARCH=$(GOARCH); go build \
		-ldflags "-X main.Release=$(release) -X main.ProgName=$(prog)" \
		-o $(bin) cmd/main.go

test: $(srcs) $(test_srcs)  ## Runs all tests
	@ginkgo $(pkgs)
