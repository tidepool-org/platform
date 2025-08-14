export TIMESTAMP ?= $(shell date +%s)
TIMING_CMD ?=

MAKEFILE:=$(realpath $(lastword $(MAKEFILE_LIST)))
ROOT_DIRECTORY:=$(realpath $(dir $(MAKEFILE)))

PLATFORM_PACKAGE=github.com/tidepool-org/platform

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

FIND_CMD=find . -not -path '*/.git/*' -not -path '*/.gvm_local/*' -not -path '*/.vs_code/*'
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

DOCKER_LOGIN_CMD ?= docker login
DOCKER_BUILD_CMD ?= docker build
DOCKER_PUSH_CMD ?= docker push
DOCKER_TAG_CMD ?= docker tag

ifdef TRAVIS_COMMIT
ifdef TRAVIS_BRANCH
ifeq ($(TRAVIS_BRANCH),master)
    DOCKER:=true
else # ifneq ($(shell go env GOWORK),) # TODO: Enable after BACK-3295 is merged into master
    DOCKER:=true
endif
ifdef DOCKER
	DOCKER_TRAVIS_BRANCH:=$(subst /,-,$(TRAVIS_BRANCH))
endif
endif
endif

ifdef DOCKER
ifdef DOCKER_SERVICE
	DOCKER_REPOSITORY:=tidepool/$(REPOSITORY_NAME)-$(DOCKER_SERVICE)
ifneq ($(shell go env GOWORK),)
	DOCKER_REPOSITORY:=$(DOCKER_REPOSITORY)-private
endif
endif
endif

PLUGINS=abbott

ifeq ($(shell go env GOWORK),)
	PLUGIN_VISIBILITY:=public
else
	PLUGIN_VISIBILITY:=private
endif

SERVICES=auth blob data migrations prescription task tools

default: test

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
		for PLUGIN in $(PLUGINS); do $(MAKE) plugin-visibility PLUGIN="$${PLUGIN}"; done

plugin-visibility:
ifdef PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		echo "Plugin $(PLUGIN) is `go run $(GO_BUILD_FLAGS) $${GOWORK_FLAGS:-} plugin/visibility/visibility.go`."
endif

plugins-visibility-public:
	@cd $(ROOT_DIRECTORY) && \
		for PLUGIN in $(PLUGINS); do $(MAKE) plugin-visibility-public PLUGIN="$${PLUGIN}"; done

plugin-visibility-public:
ifdef PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ [ ! -e go.work ] || go work edit -dropuse=./private/plugin/$(PLUGIN); } && \
		{ [ "`go list -m -mod=readonly`" != "${REPOSITORY_PACKAGE}" ] || rm go.work go.work.sum 2> /dev/null || true; } && \
		git config set --local submodule.private/plugin/$(PLUGIN).update none && \
		git config set --file=.gitmodules submodule.private/plugin/$(PLUGIN).update none && \
		$(MAKE) plugin-visibility
endif

plugins-visibility-private:
	@cd $(ROOT_DIRECTORY) && \
		for PLUGIN in $(PLUGINS); do $(MAKE) plugin-visibility-private PLUGIN="$${PLUGIN}"; done

plugin-visibility-private:
ifdef PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ git config unset --local submodule.private/plugin/$(PLUGIN).update || true; } && \
		{ git config unset --file=.gitmodules submodule.private/plugin/$(PLUGIN).update || true; } && \
		git submodule update --init private/plugin/$(PLUGIN) && \
		{ [ -e go.work ] || go work init .; } && \
		go work edit -use=./private/plugin/$(PLUGIN) && \
		go work edit -go=`sed -n 's/^go //p' go.mod` && \
		go work edit -toolchain=`sed -n 's/^toolchain //p' go.mod` && \
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
		(diff .govetignore _tmp/govet.out && exit 1)

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
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && $(TIMING_CMD) ginkgo $(GINKGO_FLAGS) $${GOWORK_FLAGS:-} $(TEST)

test-ginkgo-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && ginkgo $(GINKGO_FLAGS) -untilItFails $${GOWORK_FLAGS:-} $(TEST)

test-ginkgo-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && ginkgo watch $(GINKGO_FLAGS) $${GOWORK_FLAGS:-} $(TEST)

ci-test-ginkgo: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && $(TIMING_CMD) ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $${GOWORK_FLAGS:-} $(TEST)

ci-test-ginkgo-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) -untilItFails $${GOWORK_FLAGS:-} $(TEST)

ci-test-ginkgo-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $${GOWORK_FLAGS:-} $(TEST)

test-go:
	@echo "go test $(GOTEST_FLAGS) $(GOTEST_PKGS)"
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z `go env GOWORK` ] || GOWORK_FLAGS=-mod=readonly; } && \
		. ./env.test.sh && $(TIMING_CMD) go test $(GOTEST_FLAGS) $${GOWORK_FLAGS:-} $(GOTEST_PKGS)

ci-test-go: GOTEST_FLAGS += -race -cover
ci-test-go: GOTEST_PKGS = ./...
ci-test-go: test-go

service-build:
ifdef SERVICE
	@$(MAKE) build BUILD=$${SERVICE}
endif

service-start: CompileDaemon tmp
ifdef SERVICE
	@cd $(ROOT_DIRECTORY) && BUILD=$(SERVICE) CompileDaemon -build-dir='.' -build='make build' -command='_bin/$(SERVICE)/$(notdir $(SERVICE))' -directory='_tmp' -pattern='^$$' -include='$(subst /,.,$(SERVICE)).restart' -recursive=false -log-prefix=false -graceful-kill=true -graceful-timeout=60
endif

service-debug: CompileDaemon tmp
ifdef SERVICE
ifdef DEBUG_PORT
	@cd $(ROOT_DIRECTORY) && BUILD=$(SERVICE) CompileDaemon -build-dir='.' -build='make build' -command='dlv exec --headless --log --listen=:$(DEBUG_PORT) --api-version=2 _bin/$(SERVICE)/$(notdir $(SERVICE))' -directory='_tmp' -pattern='^$$' -include='$(subst /,.,$(SERVICE)).restart' -recursive=false -log-prefix=false -graceful-kill=true -graceful-timeout=60
endif
endif

service-restart: tmp
ifdef SERVICE
	@cd $(ROOT_DIRECTORY) && date +'%Y-%m-%dT%H:%M:%S%z' > _tmp/$(subst /,.,$(SERVICE)).restart
endif

service-restart-all:
	@cd $(ROOT_DIRECTORY) && for SERVICE in $(shell ls -1 services) ; do $(MAKE) service-restart SERVICE="services/$${SERVICE}"; done

docker-dump:
	@echo "DOCKER=$(DOCKER)"
	@echo "DOCKER_REPOSITORY=$(DOCKER_REPOSITORY)"

docker:
ifdef DOCKER
	@cd $(ROOT_DIRECTORY) && \
		for SERVICE in $(SERVICES); do $(MAKE) docker-build DOCKER_SERVICE="$${SERVICE}" TIMESTAMP="$(TIMESTAMP)"; done && \
		for SERVICE in $(SERVICES); do $(MAKE) docker-push DOCKER_SERVICE="$${SERVICE}" TIMESTAMP="$(TIMESTAMP)"; done
endif

docker-login:
ifdef DOCKER_REPOSITORY
	@echo "$(DOCKER_PASSWORD)" | $(DOCKER_LOGIN_CMD) --username "$(DOCKER_USERNAME)" --password-stdin
endif

docker-build: docker-dump docker-login
ifdef DOCKER_REPOSITORY
	@cd $(ROOT_DIRECTORY) && \
		$(TIMING_CMD) $(DOCKER_BUILD_CMD) --build-arg=PLUGIN_VISIBILITY=$(PLUGIN_VISIBILITY) --target=platform-${DOCKER_SERVICE} --tag $(DOCKER_REPOSITORY) .
ifdef DOCKER_TRAVIS_BRANCH
	@cd $(ROOT_DIRECTORY) && \
		$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-$(TRAVIS_COMMIT)-$(TIMESTAMP) && \
		$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-$(TRAVIS_COMMIT) && \
		$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-latest
endif
endif

docker-push: docker-dump docker-login
ifdef DOCKER_REPOSITORY
ifdef DOCKER_TRAVIS_BRANCH
ifeq ($(DOCKER_TRAVIS_BRANCH),master)
	@cd $(ROOT_DIRECTORY) && \
		$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY)
endif
	@cd $(ROOT_DIRECTORY) && \
		$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-$(TRAVIS_COMMIT)-$(TIMESTAMP) && \
		$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-$(TRAVIS_COMMIT) && \
		$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):$(DOCKER_TRAVIS_BRANCH)-latest
endif
endif

ci-docker: version-write docker

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
	@grep -E '^[^ #]+:( |$$)' $(MAKEFILE) | sed -E 's/^([^ #]+):.*/\1/' | sort -u | xargs echo '.PHONY:' | fold -s -w 80 | sed '$$!s/$$/\\/;2,$$s/^/    /g' >> $(MAKEFILE)

.PHONY: bindir build build-list build-watch buildable ci ci-build \
    ci-build-watch ci-docker ci-generate ci-init ci-test ci-test-ginkgo \
    ci-test-ginkgo-until-failure ci-test-ginkgo-watch ci-test-go clean clean-all \
    clean-bin clean-cover clean-debug clean-generate clean-test clean-version \
    CompileDaemon default docker docker-build docker-dump docker-login docker-push \
    format format-write format-write-changed generate ginkgo go-generate \
    go-mod-download go-mod-tidy goimports imports imports-write \
    imports-write-changed init mockgen phony plugin-visibility \
    plugin-visibility-private plugin-visibility-public plugins-visibility \
    plugins-visibility-private plugins-visibility-public pre-commit service-build \
    service-debug service-restart service-restart-all service-start test \
    test-ginkgo test-ginkgo-until-failure test-ginkgo-watch test-go tmp \
    version-write vet vet-ignore
