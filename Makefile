# Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
#
# Created by M. Massenzio, 2022-06-23

bin := out/bin/opatest
version := 0.1.0

# Source files & Test files definitions
#
# Edit only the packages list, when adding new functionality,
# the rest is deduced automatically.
#
pkgs := ./gentests ./gentests/internals
all_go := $(shell for d in $(pkgs); do find $$d -name "*.go"; done)
test_srcs := $(shell for d in $(pkgs); do find $$d -name "*_test.go"; done)
srcs := $(filter-out $(test_srcs),$(all_go))

##@ Development

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

$(bin): cmd/main.go $(srcs)

build: $(bin)  ## Build the opatest binary in the out/bin directory.
	@go build -ldflags "-X main.Release=$(version) -X main.ProgName=$(shell basename $(bin))" \
		-o $(bin) cmd/main.go

test: $(srcs) $(test_srcs)  ## Runs all tests
	@ginkgo $(pkgs)
