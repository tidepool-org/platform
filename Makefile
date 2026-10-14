export TIMESTAMP ?= $(shell date +%s)
TIMING_CMD ?=

MAKEFILE:=$(realpath $(lastword $(MAKEFILE_LIST)))
ROOT_DIRECTORY:=$(realpath $(dir $(MAKEFILE)))

PLATFORM_PACKAGE=github.com/tidepool-org/platform

PLATFORM_PLUGINS ?= abbott
PLATFORM_SERVICES ?= auth blob data migrations prescription task tools

ifeq ($(shell go env GOWORK),)
PLATFORM_PLUGIN_VISIBILITY:=public
else
PLATFORM_PLUGIN_VISIBILITY:=private
endif

REPOSITORY_PACKAGE:=$(shell sed -En 's/^module (.*)$$/\1/p' $(ROOT_DIRECTORY)/go.mod)
REPOSITORY_NAME:=$(notdir $(REPOSITORY_PACKAGE))

BIN_DIRECTORY := ${ROOT_DIRECTORY}/_bin
PATH := ${PATH}:${BIN_DIRECTORY}

ifneq ($(wildcard ./version.env),)
include ./version.env
endif

VERSION_BASE?=$(REPOSITORY_NAME)
VERSION_SHORT_COMMIT?=$(shell git rev-parse --short HEAD 2> /dev/null || echo "dev")
VERSION_FULL_COMMIT?=$(shell git rev-parse HEAD 2> /dev/null || echo "dev")
VERSION_PACKAGE?=$(PLATFORM_PACKAGE)/application

GOIMPORTS_LOCAL:=$(PLATFORM_PACKAGE)

GO_BUILD_FLAGS+=-buildvcs=false
ifdef DELVE_PORT
GO_BUILD_FLAGS+=-gcflags 'all=-N -l'
endif

GO_LD_FLAGS:=-ldflags '-X $(VERSION_PACKAGE).VersionBase=$(VERSION_BASE) -X $(VERSION_PACKAGE).VersionShortCommit=$(VERSION_SHORT_COMMIT) -X $(VERSION_PACKAGE).VersionFullCommit=$(VERSION_FULL_COMMIT)'

FIND_CMD=find . -not -path '*/.git/*' -not -path '*/.gvm_local/*' -not -path '*/.vscode/*'
FIND_MAIN_CMD:=$(FIND_CMD) -not -path './private/*' -path './$(BUILD)*' -type f -name '*.go' -not -name '*_test.go' -exec grep -E -l '^\s*func\s+main\s*(\s*)' {} \;
TRANSFORM_GO_BUILD_CMD:=sed 's|\.\(.*\)\(/[^/]*\)/[^/]*|_bin\1\2\2 .\1\2/.|'

GO_BUILD_CMD:=go build $(GO_BUILD_FLAGS) $(GO_LD_FLAGS)

ifndef GOTEST_PKGS
ifdef TEST
GOTEST_PKGS=$(REPOSITORY_PACKAGE)/$(TEST)
else
GOTEST_PKGS=./...
endif
endif
GOTEST_FLAGS ?=

GINKGO_FLAGS += --require-suite --poll-progress-after=10s --poll-progress-interval=20s -r
GINKGO_CI_WATCH_FLAGS += --randomize-all --succinct --fail-on-pending --cover --trace --race
GINKGO_CI_FLAGS += $(GINKGO_CI_WATCH_FLAGS) --randomize-suites --keep-going

ifdef TEST
TEST_PACKAGE=$(shell [ -d "$(TEST)" ] && echo "$(TEST)" || dirname "$(TEST)")
TEST_PACKAGE_NAME=$(notdir $(TEST_PACKAGE))
TEST_DIRECTORY=$(realpath $(TEST_PACKAGE))
TEST_FILE=$(notdir $(TEST))
endif

TEST_REPEAT ?= 1

# Help prevent erroneous Docker tag or push
ifneq ($(origin DOCKER),undefined)
$(error DOCKER must not be defined outside Makefile)
endif

# If a service Docker build
ifneq ($(DOCKER_SERVICE),)

# If not on Travis
ifneq ($(TRAVIS),true)

# Enable Docker for local builds
override DOCKER := local

# Otherwise, only on Travis push or API-triggered non-PR, non-tag builds of the official repository
else ifneq ($(filter push api,$(TRAVIS_EVENT_TYPE)),)
ifeq (|$(TRAVIS_PULL_REQUEST)|$(TRAVIS_TAG)|$(TRAVIS_REPO_SLUG)|,|false||tidepool-org/platform|)

# Ensure required environment variables
ifeq ($(value TRAVIS_BRANCH),)
$(error TRAVIS_BRANCH not defined for Travis CI Docker build)
else ifeq ($(TRAVIS_COMMIT),)
$(error TRAVIS_COMMIT not defined for Travis CI Docker build)
endif

# Docker tags limited to 128 [A-Za-z0-9_.-] characters; ensure branch matches same (with '/' translated to '-' below),
# but limit to 76 characters to allow for '-<commit>-<timestamp>' suffix below
ifneq ($(shell printf '%s' "$${TRAVIS_BRANCH}" | LC_ALL=C grep -Eq '^[A-Za-z0-9_./-]{1,76}$$' && echo valid),valid)
$(error TRAVIS_BRANCH not valid; limited to 76 [A-Za-z0-9_./-] characters)
else ifneq ($(filter master-% master/%,$(value TRAVIS_BRANCH)),)
$(error TRAVIS_BRANCH not valid; 'master-' and 'master/' prefix are reserved for the 'master' branch)
endif

# If master branch or private plugin visibility, enable Docker for Travis builds
ifeq ($(value TRAVIS_BRANCH),master)
override DOCKER := travis
else ifeq ($(PLATFORM_PLUGIN_VISIBILITY),private)
override DOCKER := travis
endif

endif
endif

# If on any Docker build
ifneq ($(DOCKER),)

DOCKER_BUILDX_CMD ?= docker buildx

# Calculate base Docker repository, including '-private' suffix if the plugin visibility is private
override DOCKER_REPOSITORY := $(REPOSITORY_NAME)-$(DOCKER_SERVICE)
ifeq ($(PLATFORM_PLUGIN_VISIBILITY),private)
override DOCKER_REPOSITORY := $(DOCKER_REPOSITORY)-private
endif

override DOCKER_BUILD_FLAGS := --target=platform-$(DOCKER_SERVICE) --build-arg=SERVICE=$(DOCKER_SERVICE) --build-arg=PLUGIN_VISIBILITY=$(PLATFORM_PLUGIN_VISIBILITY)

# If Travis Docker build
ifeq ($(DOCKER),travis)

DOCKER_LOGIN_CMD ?= docker login
DOCKER_BUILDER ?= platform
DOCKER_PLATFORM ?= linux/amd64,linux/arm64

# Push to the Tidepool Docker repository
override DOCKER_REPOSITORY := tidepool/$(DOCKER_REPOSITORY)

# Calculate the Docker build flags for the service
override DOCKER_BUILD_FLAGS := $(DOCKER_BUILD_FLAGS) --builder=$(DOCKER_BUILDER) --platform=$(DOCKER_PLATFORM) --provenance=false

# Calculate the Docker branch name for the Travis build; use the unexpanded value so make does not interpret it
override DOCKER_TRAVIS_BRANCH := $(subst /,-,$(value TRAVIS_BRANCH))

# Calculate the Docker tags for the Travis build based on the branch, commit, and timestamp
override DOCKER_TRAVIS_TAGS := $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-$(TRAVIS_COMMIT)-$(TIMESTAMP) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-$(TRAVIS_COMMIT) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-latest
ifeq ($(DOCKER_TRAVIS_BRANCH),master)
override DOCKER_TRAVIS_TAGS+=$(DOCKER_REPOSITORY):latest
endif

# Otherwise, if a local Docker build
else ifeq ($(DOCKER),local)

# Calculate the Docker build flags for the service
override DOCKER_BUILD_FLAGS := $(DOCKER_BUILD_FLAGS) --tag=$(DOCKER_REPOSITORY):local --output=type=docker

endif
endif
endif

default: generate build test

tmp:
	@mkdir -p $(ROOT_DIRECTORY)/_tmp

bindir:
	@mkdir -p $(ROOT_DIRECTORY)/_bin

CompileDaemon:
ifeq ($(shell which CompileDaemon),)
	@cd $(ROOT_DIRECTORY) && \
		echo "go install github.com/githubnemo/CompileDaemon" && \
		GOWORK=off go install github.com/githubnemo/CompileDaemon
endif

mockgen:
ifeq ($(shell which mockgen),)
	@cd $(ROOT_DIRECTORY) && \
		echo "go install go.uber.org/mock/mockgen" && \
		GOWORK=off go install go.uber.org/mock/mockgen
endif

ginkgo:
ifeq ($(shell which ginkgo),)
	@cd $(ROOT_DIRECTORY) && \
		echo "go install github.com/onsi/ginkgo/v2/ginkgo" && \
		GOWORK=off go install github.com/onsi/ginkgo/v2/ginkgo
endif

goimports:
ifeq ($(shell which goimports),)
	@cd $(ROOT_DIRECTORY) && \
		echo "go install golang.org/x/tools/cmd/goimports" && \
		GOWORK=off go install golang.org/x/tools/cmd/goimports
endif

buildable: export GOBIN = ${BIN_DIRECTORY}
buildable: bindir CompileDaemon ginkgo goimports

plugins-visibility:
	@cd $(ROOT_DIRECTORY) && \
		for PLATFORM_PLUGIN in $(PLATFORM_PLUGINS); do $(MAKE) plugin-visibility PLATFORM_PLUGIN="$${PLATFORM_PLUGIN}" || exit 1; done

plugin-visibility:
ifdef PLATFORM_PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		VISIBILITY=`go run $(GO_BUILD_FLAGS) $${GOWORK_FLAGS:-} plugin/visibility/visibility.go $(PLATFORM_PLUGIN)` && \
		echo "Plugin $(PLATFORM_PLUGIN) is $${VISIBILITY}."
endif

plugins-visibility-public:
	@cd $(ROOT_DIRECTORY) && \
		for PLATFORM_PLUGIN in $(PLATFORM_PLUGINS); do $(MAKE) plugin-visibility-public PLATFORM_PLUGIN="$${PLATFORM_PLUGIN}" || exit 1; done

plugin-visibility-public:
ifdef PLATFORM_PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ [ ! -e go.work ] || go work edit -dropuse=./private/plugin/$(PLATFORM_PLUGIN); } && \
		{ [ "`go list -m -mod=readonly`" != "${REPOSITORY_PACKAGE}" ] || rm go.work go.work.sum 2> /dev/null || true; } && \
		{ ! git rev-parse --git-dir > /dev/null 2>&1 || git config --local submodule.private/plugin/$(PLATFORM_PLUGIN).update none; } && \
		$(MAKE) plugin-visibility
endif

plugins-visibility-private:
	@cd $(ROOT_DIRECTORY) && \
		for PLATFORM_PLUGIN in $(PLATFORM_PLUGINS); do $(MAKE) plugin-visibility-private PLATFORM_PLUGIN="$${PLATFORM_PLUGIN}" || exit 1; done

plugin-visibility-private:
ifdef PLATFORM_PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ ! git rev-parse --git-dir > /dev/null 2>&1 || { \
			git config --local submodule.private/plugin/$(PLATFORM_PLUGIN).update checkout && \
			git submodule update --init private/plugin/$(PLATFORM_PLUGIN); }; } && \
		{ [ -e go.work ] || go work init .; } && \
		go work edit -use=./private/plugin/$(PLATFORM_PLUGIN) && \
		go work edit -go=`sed -n 's/^go //p' go.mod` && \
		go work edit -toolchain=`grep -q '^toolchain' go.mod && sed -n 's/^toolchain //p' go.mod || sed -n 's/^go /go/p' go.mod` && \
		$(MAKE) plugin-visibility
endif

ci: ci-init ci-generate ci-build ci-test ci-docker

init: go-mod-download

ci-init: init mockgen goimports

go-mod-tidy:
	@echo "go mod tidy"
	@cd $(ROOT_DIRECTORY) && \
		$(TIMING_CMD) go mod tidy

go-mod-download:
	@echo "go mod download"
	@cd $(ROOT_DIRECTORY) && \
		$(TIMING_CMD) go mod download

go-generate: mockgen
	@echo "go generate ./..."
	@cd $(ROOT_DIRECTORY) && \
		GOWORK=off $(TIMING_CMD) go generate ./...

generate: go-generate format-write imports-write vet

ci-generate: generate
	@cd $(ROOT_DIRECTORY) && \
		O=`git status -s | grep -E -v '(\.gitmodules|go\.sum)' || true` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format:
	@echo "gofmt -d -e -s"
	@cd $(ROOT_DIRECTORY) && \
		O=`$(FIND_CMD) -type f -name '*.go' -exec gofmt -d -e -s {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format-write:
	@echo "gofmt -e -s -w"
	@cd $(ROOT_DIRECTORY) && \
		O=`$(FIND_CMD) -type f -name '*.go' -exec gofmt -e -s -w {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format-write-changed:
	@cd $(ROOT_DIRECTORY) && \
		git diff --name-only | grep '\.go$$' | xargs -I{} gofmt -e -s -w {}

imports: goimports
	@echo "goimports -d -e -local $(GOIMPORTS_LOCAL)"
	@cd $(ROOT_DIRECTORY) && \
		O=`$(FIND_CMD) -type f -name '*.go' -exec goimports -d -e -local $(GOIMPORTS_LOCAL) {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports-write: goimports
	@echo "goimports -e -w -local $(GOIMPORTS_LOCAL)"
	@cd $(ROOT_DIRECTORY) && \
		O=`$(FIND_CMD) -type f -name '*.go' -exec goimports -e -w -local $(GOIMPORTS_LOCAL) {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports-write-changed: goimports
	@cd $(ROOT_DIRECTORY) && \
		git diff --name-only | grep '\.go$$' | xargs -I{} goimports -e -w -local $(GOIMPORTS_LOCAL) {}

vet: tmp
	@echo "go vet ./..."
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		go vet $${GOWORK_FLAGS:-} ./... > _tmp/govet.out 2>&1 || \
		{ LC_ALL=C sort -o _tmp/govet.out _tmp/govet.out && diff .govetignore _tmp/govet.out; }

vet-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/govet.out .govetignore

build-list:
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD)

build:
	@echo "go build $(BUILD)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		$(TIMING_CMD) $(FIND_MAIN_CMD) | $(TRANSFORM_GO_BUILD_CMD) | while read LINE; do \
			$(GO_BUILD_CMD) $${GOWORK_FLAGS:-} -o $${LINE}; \
		done

build-watch: CompileDaemon
	@cd $(ROOT_DIRECTORY) && BUILD=$(BUILD) CompileDaemon -build-dir='.' -build='make build' -color -directory='.' -exclude-dir='.git' -exclude-dir='.gvm_local' -exclude-dir='.vscode' -exclude='*_test.go' -include='Makefile' -recursive=true

ci-build: build

ci-build-watch: CompileDaemon
	@cd $(ROOT_DIRECTORY) && BUILD=$(BUILD) CompileDaemon -build-dir='.' -build='make ci-build' -color -directory='.' -exclude-dir='.git' -exclude-dir='.gvm_local' -exclude-dir='.vscode' -include='Makefile' -recursive=true

test: test-go

ci-test: ci-test-go

test-ginkgo: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off $(TIMING_CMD) ginkgo $(GINKGO_FLAGS) $(TEST)

test-ginkgo-repeat: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) --repeat $(TEST_REPEAT) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off ginkgo $(GINKGO_FLAGS) --repeat $(TEST_REPEAT) $(TEST)

test-ginkgo-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) --until-it-fails $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off ginkgo $(GINKGO_FLAGS) --until-it-fails $(TEST)

test-ginkgo-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off ginkgo watch $(GINKGO_FLAGS) $(TEST)

ci-test-ginkgo: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off $(TIMING_CMD) ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $(TEST)

ci-test-ginkgo-repeat: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) --repeat $(TEST_REPEAT) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) --repeat $(TEST_REPEAT) $(TEST)

ci-test-ginkgo-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) --until-it-fails $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) --until-it-fails $(TEST)

ci-test-ginkgo-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		. ./env.test.sh && GOWORK=off ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $(TEST)

test-go:
	@echo "go test $(GOTEST_FLAGS) $(GOTEST_PKGS)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && $(TIMING_CMD) go test $(GOTEST_FLAGS) $${GOWORK_FLAGS:-} $(GOTEST_PKGS)

ci-test-go: GOTEST_FLAGS += -count=1 -race -shuffle=on -cover
ci-test-go: GOTEST_PKGS = ./...
ci-test-go: test-go

ginkgo-bootstrap: ginkgo
ifdef TEST
	@echo "ginkgo bootstrap $(TEST_PACKAGE)"
	@cd "$(TEST_DIRECTORY)" && \
		[ -f "$(TEST_PACKAGE_NAME)_suite_test.go" ] || \
		ginkgo bootstrap --template "$(ROOT_DIRECTORY)/.ginkgo/templates/bootstrap"
endif

ginkgo-generate: ginkgo-bootstrap
ifdef TEST
	@echo "ginkgo generate $(TEST)"
	@cd "$(TEST_DIRECTORY)" && \
		ginkgo generate --template "$(ROOT_DIRECTORY)/.ginkgo/templates/generate" $(TEST_FILE)
endif

service-build:
ifdef PLATFORM_SERVICE
	@$(MAKE) build BUILD=$${PLATFORM_SERVICE}
endif

service-start: CompileDaemon tmp
ifdef PLATFORM_SERVICE
	@cd $(ROOT_DIRECTORY) && BUILD=$(PLATFORM_SERVICE) CompileDaemon -build-dir='.' -build='make build' -command='_bin/$(PLATFORM_SERVICE)/$(notdir $(PLATFORM_SERVICE))' -directory='_tmp' -pattern='^$$' -include='$(subst /,.,$(PLATFORM_SERVICE)).restart' -recursive=false -log-prefix=false -graceful-kill=true -graceful-timeout=60
endif

service-debug: CompileDaemon tmp
ifdef PLATFORM_SERVICE
ifdef DEBUG_PORT
	@cd $(ROOT_DIRECTORY) && BUILD=$(PLATFORM_SERVICE) CompileDaemon -build-dir='.' -build='make build' -command='dlv exec --headless --log --listen=:$(DEBUG_PORT) --api-version=2 _bin/$(PLATFORM_SERVICE)/$(notdir $(PLATFORM_SERVICE))' -directory='_tmp' -pattern='^$$' -include='$(subst /,.,$(PLATFORM_SERVICE)).restart' -recursive=false -log-prefix=false -graceful-kill=true -graceful-timeout=60
endif
endif

service-restart: tmp
ifdef PLATFORM_SERVICE
	@cd $(ROOT_DIRECTORY) && date +'%Y-%m-%dT%H:%M:%S%z' > _tmp/$(subst /,.,$(PLATFORM_SERVICE)).restart
endif

service-restart-all:
	@cd $(ROOT_DIRECTORY) && for PLATFORM_SERVICE in $(shell ls -1 services) ; do $(MAKE) service-restart PLATFORM_SERVICE="services/$${PLATFORM_SERVICE}"; done

docker-dump:
	@echo "DOCKER=$(DOCKER)"
	@echo "DOCKER_SERVICE=$(DOCKER_SERVICE)"
	@echo "DOCKER_REPOSITORY=$(DOCKER_REPOSITORY)"
	@echo "DOCKER_BUILD_FLAGS=$(DOCKER_BUILD_FLAGS)"
	@echo "DOCKER_TRAVIS_TAGS=$(DOCKER_TRAVIS_TAGS)"

docker:
	@cd $(ROOT_DIRECTORY) && \
		for PLATFORM_SERVICE in $(PLATFORM_SERVICES); do $(MAKE) docker-build DOCKER_SERVICE="$${PLATFORM_SERVICE}" || exit 1; done

docker-build: docker-dump
ifeq ($(DOCKER),local)
	@cd $(ROOT_DIRECTORY) && \
		$(TIMING_CMD) $(DOCKER_BUILDX_CMD) build $(DOCKER_BUILD_FLAGS) .
endif

ci-docker: version-write
	@cd $(ROOT_DIRECTORY) && \
		for PLATFORM_SERVICE in $(PLATFORM_SERVICES); do $(MAKE) ci-docker-build DOCKER_SERVICE="$${PLATFORM_SERVICE}" TIMESTAMP="$(TIMESTAMP)" || exit 1; done && \
		for PLATFORM_SERVICE in $(PLATFORM_SERVICES); do $(MAKE) ci-docker-push DOCKER_SERVICE="$${PLATFORM_SERVICE}" TIMESTAMP="$(TIMESTAMP)" || exit 1; done

ci-docker-login:
ifeq ($(DOCKER),travis)
	@echo "$(DOCKER_PASSWORD)" | $(DOCKER_LOGIN_CMD) --username "$(DOCKER_USERNAME)" --password-stdin
endif

ci-docker-builder:
ifeq ($(DOCKER),travis)
	@$(DOCKER_BUILDX_CMD) inspect $(DOCKER_BUILDER) > /dev/null 2>&1 || \
		$(DOCKER_BUILDX_CMD) create --name $(DOCKER_BUILDER) --driver docker-container --bootstrap
endif

# Multi-platform images cannot be loaded locally; build every service into the builder cache before push
ci-docker-build: docker-dump ci-docker-login ci-docker-builder
ifeq ($(DOCKER),travis)
	@cd $(ROOT_DIRECTORY) && \
		$(TIMING_CMD) $(DOCKER_BUILDX_CMD) build $(DOCKER_BUILD_FLAGS) .
endif

ci-docker-push: docker-dump ci-docker-login ci-docker-builder
ifeq ($(DOCKER),travis)
	@cd $(ROOT_DIRECTORY) && \
		$(TIMING_CMD) $(DOCKER_BUILDX_CMD) build $(DOCKER_BUILD_FLAGS) $(addprefix --tag=,$(DOCKER_TRAVIS_TAGS)) --push .
endif

version-write:
	@cd $(ROOT_DIRECTORY) && \
		echo "VERSION_BASE=$(VERSION_BASE)" > version.env && \
		echo "VERSION_SHORT_COMMIT=$(VERSION_SHORT_COMMIT)" >> version.env && \
		echo "VERSION_FULL_COMMIT=$(VERSION_FULL_COMMIT)" >> version.env && \
		echo "VERSION_PACKAGE=$(VERSION_PACKAGE)" >> version.env

clean: clean-bin clean-cover clean-debug clean-generate clean-test clean-version
	@cd $(ROOT_DIRECTORY) && rm -rf _tmp

clean-bin:
	@cd $(ROOT_DIRECTORY) && rm -rf _bin _log

clean-cover:
	@cd $(ROOT_DIRECTORY) && $(FIND_CMD) -type f \( -name '*.coverprofile' -o -name 'coverprofile.out' \) -delete
	@cd $(ROOT_DIRECTORY) && $(FIND_CMD) -type d -name 'coverage' -empty -delete

clean-debug:
	@cd $(ROOT_DIRECTORY) && $(FIND_CMD) -type f \( -name 'debug' -o -name '__debug_bin*' \) -delete

clean-generate:
	@cd $(ROOT_DIRECTORY) && $(FIND_CMD) -type f -path '*/gomock_reflect_*/*' -delete
	@cd $(ROOT_DIRECTORY) && $(FIND_CMD) -type d -name 'gomock_reflect_*' -delete

clean-test:
	@cd $(ROOT_DIRECTORY) && $(FIND_CMD) -type f \( -name '*.test' -o -name '*.report' \) -delete

clean-version:
	@cd $(ROOT_DIRECTORY) && rm -rf version.env

clean-all: clean

pre-commit: format imports vet

phony:
	@{ rm $(MAKEFILE) && sed -E -n '/^.PHONY: /q;p' > $(MAKEFILE); } < $(MAKEFILE)
	@grep -E '^[^ #]+:( |$$)' $(MAKEFILE) | sed -E 's/^([^ #]+):.*/\1/' | sort -u | xargs echo '.PHONY:' | fold -s -w 80 | sed '$$!s/$$/\\/;2,$$s/^/\t/g' >> $(MAKEFILE)

.PHONY: bindir build build-list build-watch buildable ci ci-build \
	ci-build-watch ci-docker ci-docker-build ci-docker-builder ci-docker-login \
	ci-docker-push ci-generate ci-init ci-test ci-test-ginkgo ci-test-ginkgo-repeat \
	ci-test-ginkgo-until-failure ci-test-ginkgo-watch ci-test-go clean clean-all \
	clean-bin clean-cover clean-debug clean-generate clean-test clean-version \
	CompileDaemon default docker docker-build docker-dump format format-write \
	format-write-changed generate ginkgo ginkgo-bootstrap ginkgo-generate \
	go-generate go-mod-download go-mod-tidy goimports imports imports-write \
	imports-write-changed init mockgen phony plugin-visibility \
	plugin-visibility-private plugin-visibility-public plugins-visibility \
	plugins-visibility-private plugins-visibility-public pre-commit service-build \
	service-debug service-restart service-restart-all service-start test \
	test-ginkgo test-ginkgo-repeat test-ginkgo-until-failure test-ginkgo-watch \
	test-go tmp version-write vet vet-ignore
