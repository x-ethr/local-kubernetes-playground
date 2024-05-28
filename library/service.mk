SHELL := /usr/bin/env bash

# ====================================================================================
# Colors
# ------------------------------------------------------------------------------------

black        := $(shell printf "\033[30m")
black-bold   := $(shell printf "\033[30;1m")
red          := $(shell printf "\033[31m")
red-bold     := $(shell printf "\033[31;1m")
green        := $(shell printf "\033[32m")
green-bold   := $(shell printf "\033[32;1m")
yellow       := $(shell printf "\033[33m")
yellow-bold  := $(shell printf "\033[33;1m")
blue         := $(shell printf "\033[34m")
blue-bold    := $(shell printf "\033[34;1m")
magenta      := $(shell printf "\033[35m")
magenta-bold := $(shell printf "\033[35;1m")
cyan         := $(shell printf "\033[36m")
cyan-bold    := $(shell printf "\033[36;1m")
white        := $(shell printf "\033[37m")
white-bold   := $(shell printf "\033[37;1m")
reset        := $(shell printf "\033[0m")

# ====================================================================================
# Logger
# ------------------------------------------------------------------------------------

time-long	= $(date +%Y-%m-%d' '%H:%M:%S)
time-short	= $(date +%H:%M:%S)
time		= $(time-short)

information	= echo $(time) $(blue)[ ... ]$(reset)
warning	= echo $(time) $(yellow)[ WARNING ]$(reset)
exception		= echo $(time) $(red)[ ERROR ]$(reset)
ok		= echo $(time) $(green)[ OK ]$(reset)
fail	= (echo $(time) $(red)[ FAILURE ]$(reset) && false)

# ====================================================================================
# Command(s)
# ------------------------------------------------------------------------------------

all :: deploy

service = $(shell basename $(CURDIR))

version = $(shell [ -f VERSION ] && head VERSION || echo "0.0.0")

major      		= $(shell echo $(version) | sed "s/^\([0-9]*\).*/\1/")
minor      		= $(shell echo $(version) | sed "s/[0-9]*\.\([0-9]*\).*/\1/")
patch      		= $(shell echo $(version) | sed "s/[0-9]*\.[0-9]*\.\([0-9]*\).*/\1/")

major-upgrade 	= $(shell expr $(major) + 1).$(minor).$(patch)
minor-upgrade 	= $(major).$(shell expr $(minor) + 1).$(patch)
patch-upgrade 	= $(major).$(minor).$(shell expr $(patch) + 1)

prepare:
	@$(information) Executing Module Vendoring
	@go get -u github.com/x-ethr/server
	@go mod tidy
	@$(ok) Build Preparation
	@printf "\n"

bump: prepare
	@$(information) Service Name: $(service), Version: $(version)
	@echo $(patch-upgrade) > VERSION
	@$(ok) Version Bump
	@printf "\n"

build: bump
	@$(information) Building Container Image
	@docker build --file ../Dockerfile --tag "localhost:5050/$(service):$(version)" --build-arg="SERVICE=$(service)" .
	@docker push "localhost:5050/$(service):$(version)"
	@$(ok) Build
	@printf "\n"

update: build
	@$(information) Updating Kustomization Image Patches
	@ethr-cli kubernetes kustomization update image --file ./kustomize/kustomization.yaml --image service:latest --name $(service) --tag $(version) --registry localhost:5050
	@$(ok) Update
	@printf "\n"

manifests: update
	@$(information) Applying Kuberenetes Manifests
	@kubectl apply --kustomize . --wait
	@$(ok) Manifests
	@printf "\n"

deploy: manifests
	@$(information) Beginning Rollout
	@kubectl --namespace development rollout restart deployments/$(service)
	@kubectl --namespace development rollout status deployments/$(service)
	@$(ok) Deployment
	@printf "\n"

# ====================================================================================
# Help
# ------------------------------------------------------------------------------------

define HELP
Usage: make [make-options] <target> [options]

Common Targets:
    build        Build source code and other artifacts for host platform.
    build.all    Build source code and other artifacts for all platforms.
    clean        Remove all files created during the build.
    distclean    Remove all files created during the build including cached tools.
    lint         Run lint and code analysis tools.
    help         Show this help info.
    test         Runs unit tests.
    e2e          Runs end-to-end integration tests.
    generate     Run code generation.
    reviewable   Validate that a PR is ready for review.
    check-diff   Ensure the reviewable target doesn't create a git diff.

Common Options:
    DEBUG        Whether to generate debug symbols. Default is 0.
    PLATFORM     The platform to build.
    SUITE        The test suite to run.
    TESTFILTER   Tests to run in a suite.
    V            Set to 1 enable verbose build. Default is 0.

Release Targets:
    publish      Build and publish final releasable artifacts
    promote      Promote a release to a release channel
    tag          Tag a release

Release Options:
    VERSION      The version information for binaries and releases.
    CHANNEL      Sets the release channel. Can be set to master, main, alpha, beta, or stable.

endef
export HELP

help-special: ; @:

help:
	@echo "$$HELP"
	@$(MAKE) help-special
