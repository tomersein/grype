TEMPDIR = ./.tmp
RESULTSDIR = $(TEMPDIR)/results
COVER_REPORT = $(RESULTSDIR)/cover.report
COVER_TOTAL = $(RESULTSDIR)/cover.total
LICENSES_REPORT = $(RESULTSDIR)/licenses.json
LINTCMD = $(TEMPDIR)/golangci-lint run --tests=false --config .golangci.yaml
BOLD := $(shell tput -T linux bold)
PURPLE := $(shell tput -T linux setaf 5)
GREEN := $(shell tput -T linux setaf 2)
CYAN := $(shell tput -T linux setaf 6)
RED := $(shell tput -T linux setaf 1)
RESET := $(shell tput -T linux sgr0)
TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)
# the quality gate lower threshold for unit test total % coverage (by function statements)
COVERAGE_THRESHOLD := 60

## Build variables
DISTDIR=./dist
SNAPSHOTDIR=./snapshot
GITTREESTATE=$(if $(shell git status --porcelain),dirty,clean)

ifeq "$(strip $(VERSION))" ""
 override VERSION = $(shell git describe --always --tags --dirty)
endif

## Variable assertions

ifndef TEMPDIR
	$(error TEMPDIR is not set)
endif

ifndef RESULTSDIR
	$(error RESULTSDIR is not set)
endif

ifndef DISTDIR
	$(error DISTDIR is not set)
endif

ifndef SNAPSHOTDIR
	$(error SNAPSHOTDIR is not set)
endif

define title
    @printf '$(TITLE)$(1)$(RESET)\n'
endef

.PHONY: all bootstrap lint lint-fix unit coverage integration check-pipeline clear-cache help test

all: clean lint check-licenses test ## Run all checks (linting, license check, unit, and integration tests)
	@printf '$(SUCCESS)All checks pass!$(RESET)\n'

.PHONY: compare
compare:
	@cd test/inline-compare && make

.PHONY: test
test: unit integration ## Run all tests (currently unit & integration tests )

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'

ci-bootstrap: bootstrap
	sudo apt install -y bc

.PHONY: boostrap
bootstrap: ## Download and install all go dependencies (+ prep tooling in the ./tmp dir)
	$(call title,Boostrapping dependencies)
	@pwd
	# prep temp dirs
	mkdir -p $(TEMPDIR)
	mkdir -p $(RESULTSDIR)
	# install go dependencies
	go mod download
	# install utilities
	[ -f "$(TEMPDIR)/golangci" ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TEMPDIR)/ v1.26.0
	[ -f "$(TEMPDIR)/bouncer" ] || curl -sSfL https://raw.githubusercontent.com/wagoodman/go-bouncer/master/bouncer.sh | sh -s -- -b $(TEMPDIR)/ v0.1.0
	[ -f "$(TEMPDIR)/goreleaser" ] || curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh -s -- -b $(TEMPDIR)/ v0.140.0

.PHONY: lint
lint: ## Run gofmt + golangci lint checks
	$(call title,Running linters)
	# ensure there are no go fmt differences
	@printf "files with gofmt issues: [$(shell gofmt -l -s .)]\n"
	@test -z "$(shell gofmt -l -s .)"

	# run all golangci-lint rules
	$(LINTCMD)

	# go tooling does not play well with certain filename characters, ensure the common cases don't result in future "go get" failures
	$(eval MALFORMED_FILENAMES := $(shell find . | grep -e ':'))
	@bash -c "[[ '$(MALFORMED_FILENAMES)' == '' ]] || (printf '\nfound unsupported filename characters:\n$(MALFORMED_FILENAMES)\n\n' && false)"

lint-fix: ## Auto-format all source code + run golangci lint fixers
	$(call title,Running lint fixers)
	gofmt -w -s .
	$(LINTCMD) --fix

.PHONY: check-licenses
check-licenses:
	$(TEMPDIR)/bouncer check

unit: ## Run unit tests (with coverage)
	$(call title,Running unit tests)
	mkdir -p $(RESULTSDIR)
	go test -v -coverprofile $(COVER_REPORT) ./...
	@go tool cover -func $(COVER_REPORT) | grep total |  awk '{print substr($$3, 1, length($$3)-1)}' > $(COVER_TOTAL)
	@echo "Coverage: $$(cat $(COVER_TOTAL))"
	@if [ $$(echo "$$(cat $(COVER_TOTAL)) >= $(COVERAGE_THRESHOLD)" | bc -l) -ne 1 ]; then echo "$(RED)$(BOLD)Failed coverage quality gate (> $(COVERAGE_THRESHOLD)%)$(RESET)" && false; fi

integration: ## Run integration tests
	$(call title,Running integration tests)
	go test -v -tags=integration ./test/integration

integration/test-fixtures/tar-cache.key, integration-fingerprint:
	find test/integration/test-fixtures/image-* -type f -exec md5sum {} + | awk '{print $1}' | sort | md5sum | tee test/integration/test-fixtures/tar-cache.fingerprint

clear-test-cache: ## Delete all test cache (built docker image tars)
	find . -type f -wholename "**/test-fixtures/tar-cache/*.tar" -delete

check-pipeline: ## Run local CircleCI pipeline locally (sanity check)
	$(call title,Check pipeline)
	# note: this is meant for local development & testing of the pipeline, NOT to be run in CI
	mkdir -p $(TEMPDIR)
	circleci config process .circleci/config.yml > .tmp/circleci.yml
	circleci local execute -c .tmp/circleci.yml --job "Static Analysis"
	circleci local execute -c .tmp/circleci.yml --job "Unit & Integration Tests (go-latest)"
	@printf '$(SUCCESS)Pipeline checks pass!$(RESET)\n'

.PHONY: build
build: $(SNAPSHOTDIR) ## Build release snapshot binaries and packages

$(SNAPSHOTDIR): ## Build snapshot release binaries and packages
	$(call title,Building snapshot artifacts)
	# create a config with the dist dir overridden
	echo "dist: $(SNAPSHOTDIR)" > $(TEMPDIR)/goreleaser.yaml
	cat .goreleaser.yaml >> $(TEMPDIR)/goreleaser.yaml

	# build release snapshots
	BUILD_GIT_TREE_STATE=$(GITTREESTATE) \
	$(TEMPDIR)/goreleaser release --skip-publish --rm-dist --snapshot --config $(TEMPDIR)/goreleaser.yaml

# TODO: this is not releasing yet
.PHONY: release
release: clean-dist ## Build and publish final binaries and packages
	$(call title,Publishing release artifacts)
	# create a config with the dist dir overridden
	echo "dist: $(DISTDIR)" > $(TEMPDIR)/goreleaser.yaml
	cat .goreleaser.yaml >> $(TEMPDIR)/goreleaser.yaml

	# release
	BUILD_GIT_TREE_STATE=$(GITTREESTATE) \
	$(TEMPDIR)/goreleaser --skip-publish --rm-dist --config $(TEMPDIR)/goreleaser.yaml

	# create a version file for version-update checks
	echo "$(VERSION)" > $(DISTDIR)/VERSION
	# TODO: add upload to bucket

.PHONY: clean
clean: clean-dist clean-shapshot  ## Remove previous builds and result reports
	rm -rf $(RESULTSDIR)/*

.PHONY: clean-shapshot
clean-shapshot:
	rm -rf $(SNAPSHOTDIR) $(TEMPDIR)/goreleaser.yaml

.PHONY: clean-dist
clean-dist:
	rm -rf $(DISTDIR) $(TEMPDIR)/goreleaser.yaml