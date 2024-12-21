TIMESTAMP ?= $(shell date +%s)
# ensure that we use the same timestamps in sub-makes. We've seen cases where
# these can vary by 1 second
export TIMESTAMP

SERVICES_SEPARATOR=,
SERVICES_TO_BUILD?=auth,blob,data,migrations,prescription,task,tools
SERVICES_TO_BUILD:=$(subst $(SERVICES_SEPARATOR), ,$(SERVICES_TO_BUILD))

MAKEFILE:=$(realpath $(lastword $(MAKEFILE_LIST)))
ROOT_DIRECTORY:=$(realpath $(dir $(MAKEFILE)))

REPOSITORY_GOPATH:=$(word 1, $(subst :, ,$(GOPATH)))
REPOSITORY_PACKAGE:=github.com/tidepool-org/platform
REPOSITORY_NAME:=$(notdir $(REPOSITORY_PACKAGE))

BIN_DIRECTORY := ${ROOT_DIRECTORY}/_bin
PATH := ${PATH}:${BIN_DIRECTORY}

VERSION_BASE:=platform
VERSION_SHORT_COMMIT:=$(shell git rev-parse --short HEAD || echo "dev")
VERSION_FULL_COMMIT:=$(shell git rev-parse HEAD || echo "dev")
VERSION_PACKAGE:=$(REPOSITORY_PACKAGE)/application

GO_BUILD_FLAGS:=-buildvcs=false
GO_LD_FLAGS:=-ldflags '-X $(VERSION_PACKAGE).VersionBase=$(VERSION_BASE) -X $(VERSION_PACKAGE).VersionShortCommit=$(VERSION_SHORT_COMMIT) -X $(VERSION_PACKAGE).VersionFullCommit=$(VERSION_FULL_COMMIT)'

FIND_MAIN_CMD:=find . -path './$(BUILD)*' -not -path './.gvm_local/*' -not -path './vendor/*' -name '*.go' -not -name '*_test.go' -type f -exec egrep -l '^\s*func\s+main\s*(\s*)' {} \;
TRANSFORM_GO_BUILD_CMD:=sed 's|\.\(.*\)\(/[^/]*\)/[^/]*|_bin\1\2\2 .\1\2/.|'

GINKGO_FLAGS += --require-suite --poll-progress-after=10s --poll-progress-interval=20s -r
GINKGO_CI_WATCH_FLAGS += --randomize-all --succinct --fail-on-pending --cover --trace --race
GINKGO_CI_FLAGS += $(GINKGO_CI_WATCH_FLAGS) --randomize-suites --keep-going

GOTEST_PKGS ?= ./...
GOTEST_FLAGS ?=

DOCKER_LOGIN_CMD ?= docker login
DOCKER_BUILD_CMD ?= docker build
DOCKER_PUSH_CMD ?= docker push
DOCKER_TAG_CMD ?= docker tag

TIMING_CMD ?=

PLUGINS=redwood

ifneq ($(shell go env GOWORK),)
ifneq ($(shell go env GOWORK),off)
	GOWORK_OFF := GOWORK=off
	GO_VET_FLAGS += -mod=readonly
	GO_BUILD_FLAGS += -mod=readonly
	GINKGO_FLAGS += -mod=readonly
	GOTEST_FLAGS += -mod=readonly
	DOCKER_REPOSITORY_SUFFIX = -private
endif
endif

GO_BUILD_CMD:=go build $(GO_BUILD_FLAGS) $(GO_LD_FLAGS) -o

ifdef TRAVIS_BRANCH
ifdef TRAVIS_COMMIT
    DOCKER:=true
endif
endif

ifeq ($(TRAVIS_BRANCH),master)
ifeq ($(TRAVIS_PULL_REQUEST_BRANCH),)
	DOCKER:=true
endif
else ifdef TRAVIS_TAG
	DOCKER:=true
endif
ifdef DOCKER_FILE
	SERVICE_NAME:=$(patsubst .%,%,$(suffix $(DOCKER_FILE)))
	ifneq ($(filter $(SERVICE_NAME),$(SERVICES_TO_BUILD)),)
		BUILD_SERVICE:=true
	endif
	DOCKER_REPOSITORY:=tidepool/$(REPOSITORY_NAME)-$(SERVICE_NAME)$(DOCKER_REPOSITORY_SUFFIX)
endif

default: test

tmp:
	@mkdir -p $(ROOT_DIRECTORY)/_tmp

bindir:
	@mkdir -p $(ROOT_DIRECTORY)/_bin

CompileDaemon:
ifeq ($(shell which CompileDaemon),)
	cd vendor/github.com/githubnemo/CompileDaemon && $(GOWORK_OFF) go install -mod=vendor .
endif

mockgen:
ifeq ($(shell which mockgen),)
	cd vendor/go.uber.org/mock/mockgen && $(GOWORK_OFF) go install -mod=vendor .
endif

ginkgo:
ifeq ($(shell which ginkgo),)
	cd vendor/github.com/onsi/ginkgo/v2/ginkgo && $(GOWORK_OFF) go install -mod=vendor .
endif

goimports:
ifeq ($(shell which goimports),)
	cd vendor/golang.org/x/tools/cmd/goimports && $(GOWORK_OFF) go install -mod=vendor .
endif

golint:
ifeq ($(shell which golint),)
	cd vendor/golang.org/x/lint/golint && $(GOWORK_OFF) go install -mod=vendor .
endif

buildable: export GOBIN = ${BIN_DIRECTORY}
buildable: bindir CompileDaemon ginkgo goimports golint

plugins-visibility:
	@cd $(ROOT_DIRECTORY) && \
		for PLUGIN in $(PLUGINS); do $(MAKE) plugin-visibility PLUGIN="$${PLUGIN}"; done

plugin-visibility:
ifdef PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		echo "Plugin $(PLUGIN) is `go run $(GO_BUILD_FLAGS) plugin/test/visibility.go`."
endif

plugins-visibility-public:
	@cd $(ROOT_DIRECTORY) && \
		for PLUGIN in $(PLUGINS); do $(MAKE) plugin-visibility-public PLUGIN="$${PLUGIN}"; done

plugin-visibility-public:
ifdef PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z "`git status -s .gitmodules`" ] || { echo '.gitmodules currently modified' && exit 1; } } && \
		{ [ ! -e go.work ] || go work edit -dropuse=./private/plugin/$(PLUGIN); } && \
		{ [ "`go list -m -mod=readonly`" != "${REPOSITORY_PACKAGE}" ] || rm go.work go.work.sum 2> /dev/null || true; } && \
		git config set --local submodule.private/plugin/$(PLUGIN).update none && \
		git config set --file=.gitmodules submodule.private/plugin/$(PLUGIN).update none && \
		git update-index --assume-unchanged .gitmodules && \
		$(MAKE) plugin-visibility
endif

plugins-visibility-private:
	@cd $(ROOT_DIRECTORY) && \
		for PLUGIN in $(PLUGINS); do $(MAKE) plugin-visibility-private PLUGIN="$${PLUGIN}"; done

plugin-visibility-private:
ifdef PLUGIN
	@cd $(ROOT_DIRECTORY) && \
		{ [ -z "`git status -s .gitmodules`" ] || { echo '.gitmodules currently modified' && exit 1; } } && \
		{ git config unset --local submodule.private/plugin/$(PLUGIN).update || true; } && \
		{ git config unset --file=.gitmodules submodule.private/plugin/$(PLUGIN).update || true; } && \
		git submodule update --init private/plugin/$(PLUGIN) && \
		git update-index --assume-unchanged .gitmodules && \
		{ [ -e go.work ] || go work init .; } && \
		go work edit -use=./private/plugin/$(PLUGIN) && \
		go work edit -go=`sed -n 's/^go //p' go.mod` && \
		go work edit -toolchain=`sed -n 's/^toolchain //p' go.mod` && \
		$(MAKE) plugin-visibility
endif

fail-if-changed:
	@cd $(ROOT_DIRECTORY) && \
		O=`git status -s` && [ "$${O}" = "" ] || (echo "$${O}" && exit 1)

go-generate: mockgen
	@echo "go generate ./..."
	@cd $(ROOT_DIRECTORY) && $(GOWORK_OFF) go generate ./...

generate: go-generate format-write imports-write vet

ci-generate: generate fail-if-changed

format:
	@echo "gofmt -d -e -s"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -not -path './vendor/*' -name '*.go' -type f -exec gofmt -d -e -s {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format-write:
	@echo "gofmt -e -s -w"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -not -path './vendor/*' -name '*.go' -type f -exec gofmt -e -s -w {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format-write-changed:
	@cd $(ROOT_DIRECTORY) && \
		git diff --name-only | grep '\.go$$' | xargs -I{} gofmt -e -s -w {}

imports: goimports
	@echo "goimports -d -e -local github.com/tidepool-org/platform"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -not -path './vendor/*' -not -path '**/test/mock.go' -not -name '*mock.go' -not -name '**_gen.go' -name '*.go' -type f -exec goimports -d -e -local github.com/tidepool-org/platform {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports-write: goimports
	@echo "goimports -e -w -local github.com/tidepool-org/platform"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -not -path './vendor/*' -name '*.go' -type f -exec goimports -e -w -local github.com/tidepool-org/platform {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports-write-changed: goimports
	@cd $(ROOT_DIRECTORY) && \
		git diff --name-only | grep '\.go$$' | xargs -I{} goimports -e -w -local github.com/tidepool-org/platform {}

vet: tmp
	@echo "go vet"
	@cd $(ROOT_DIRECTORY) && \
		go vet $(GO_VET_FLAGS) ./... > _tmp/govet.out 2>&1 || \
		(diff .govetignore _tmp/govet.out && exit 1)

vet-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/govet.out .govetignore

lint: golint tmp
	@echo "golint"
	@cd $(ROOT_DIRECTORY) && \
		find . -not -path './.gvm_local/*' -not -path './vendor/*' -name '*.go' -type f | sort -d | xargs -I {} golint {} | grep -v 'exported.*should have comment.*or be unexported' 2> _tmp/golint.out > _tmp/golint.out || [ $${?} == 1 ] && \
		diff .golintignore _tmp/golint.out || \
		exit 0

lint-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/golint.out .golintignore

build-list:
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD)

build:
	@echo "go build $(BUILD)"
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_GO_BUILD_CMD) | while read LINE; do $(GO_BUILD_CMD) $${LINE}; done

build-watch: CompileDaemon
	@cd $(ROOT_DIRECTORY) && BUILD=$(BUILD) CompileDaemon -build-dir='.' -build='make build' -color -directory='.' -exclude-dir='.git' -exclude='*_test.go' -include='Makefile' -recursive=true

ci-build: build

ci-build-watch: CompileDaemon
	@cd $(ROOT_DIRECTORY) && BUILD=$(BUILD) CompileDaemon -build-dir='.' -build='make ci-build' -color -directory='.' -exclude-dir='.git' -include='Makefile' -recursive=true

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
	@cd $(ROOT_DIRECTORY) && for SERVICE in migrations tools; do $(MAKE) service-restart SERVICE="$${SERVICE}"; done

test: go-test

ci-test: ci-go-test

ginkgo-test: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo $(GINKGO_FLAGS) $(TEST)

ginkgo-test-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo $(GINKGO_FLAGS) -untilItFails $(TEST)

ginkgo-test-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch $(GINKGO_FLAGS) $(TEST)

ci-ginkgo-test: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && $(TIMING_CMD) ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $(TEST)

ci-ginkgo-test-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) -untilItFails $(TEST)

ci-ginkgo-test-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $(TEST)

go-test:
	. ./env.test.sh && $(TIMING_CMD) go test $(GOTEST_FLAGS) $(GOTEST_PKGS)

ci-go-test: GOTEST_FLAGS += -count=1 -race -shuffle=on -cover
ci-go-test: GOTEST_PKGS = ./...
ci-go-test: go-test

deploy: clean-deploy deploy-services deploy-migrations deploy-tools

deploy-services:
ifdef TRAVIS_TAG
	@cd $(ROOT_DIRECTORY) && for SERVICE in $(shell ls -1 _bin/services); do $(MAKE) bundle-deploy DEPLOY=$${SERVICE} SOURCE=services/$${SERVICE}; done
endif

deploy-migrations:
ifdef TRAVIS_TAG
	@$(MAKE) bundle-deploy DEPLOY=migrations SOURCE=migrations
endif

deploy-tools:
ifdef TRAVIS_TAG
	@$(MAKE) bundle-deploy DEPLOY=tools SOURCE=tools
endif

ci-deploy: deploy

bundle-deploy:
ifdef DEPLOY
ifdef TRAVIS_TAG
	@cd $(ROOT_DIRECTORY) && \
		DEPLOY_TAG=$(DEPLOY)-$(TRAVIS_TAG) && \
		DEPLOY_DIR=deploy/$(DEPLOY)/$${DEPLOY_TAG} && \
		mkdir -p $${DEPLOY_DIR}/_bin/$(SOURCE) && \
		cp -R _bin/$(SOURCE)/* $${DEPLOY_DIR}/_bin/$(SOURCE)/ && \
		find $(SOURCE) -type f -name 'README.md' -exec cp {} $${DEPLOY_DIR}/_bin/{} \; && \
		cp $(SOURCE)/start.sh $${DEPLOY_DIR}/ && \
		tar -c -z -f $${DEPLOY_DIR}.tar.gz -C deploy/$(DEPLOY)/ $${DEPLOY_TAG}
endif
endif

docker:
ifdef DOCKER
	@echo "$(DOCKER_PASSWORD)" | $(DOCKER_LOGIN_CMD) --username "$(DOCKER_USERNAME)" --password-stdin
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-build DOCKER_FILE="$${DOCKER_FILE}" TIMESTAMP="$(TIMESTAMP)";done
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-push DOCKER_FILE="$${DOCKER_FILE}" TIMESTAMP="$(TIMESTAMP)";done
endif

docker-build:
ifdef DOCKER
ifdef DOCKER_FILE
ifdef BUILD_SERVICE
	$(DOCKER_BUILD_CMD) --tag $(DOCKER_REPOSITORY):development --target=development --file "$(DOCKER_FILE)" .
	$(DOCKER_BUILD_CMD) --tag $(DOCKER_REPOSITORY) --file "$(DOCKER_FILE)" .
ifdef TRAVIS_BRANCH
ifdef TRAVIS_COMMIT
ifdef TRAVIS_PULL_REQUEST_BRANCH
	$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
else
	$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-latest
	$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
endif
endif
endif
ifdef TRAVIS_TAG
	$(DOCKER_TAG_CMD) $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(TRAVIS_TAG:v%=%)
endif
else
	@echo skipping $(DOCKER_FILE)
endif
endif
endif

docker-push:
ifdef DOCKER
ifdef BUILD_SERVICE
	@echo "DOCKER_REPOSITORY = $(DOCKER_REPOSITORY)"
	@echo "TRAVIS_BRANCH = $(TRAVIS_BRANCH)"
	@echo "TRAVIS_PULL_REQUEST_BRANCH = $(TRAVIS_PULL_REQUEST_BRANCH)"
	@echo "TRAVIS_COMMIT = $(TRAVIS_COMMIT)"
	@echo "TRAVIS_TAG= $(TRAVIS_TAG)"
ifdef DOCKER_REPOSITORY
ifeq ($(TRAVIS_BRANCH),master)
ifeq ($(TRAVIS_PULL_REQUEST_BRANCH),)
	$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY)
endif
endif
ifdef TRAVIS_BRANCH
ifdef TRAVIS_COMMIT
ifdef TRAVIS_PULL_REQUEST_BRANCH
	$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
else
	$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-latest
	$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
endif
endif
endif
ifdef TRAVIS_TAG
	$(DOCKER_PUSH_CMD) $(DOCKER_REPOSITORY):$(TRAVIS_TAG:v%=%)
endif
endif
endif
endif

ci-docker: docker

clean: clean-bin clean-cover clean-debug clean-deploy clean-test
	@cd $(ROOT_DIRECTORY) && rm -rf _tmp

clean-bin:
	@cd $(ROOT_DIRECTORY) && rm -rf _bin _log

clean-cover:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.coverprofile" -o -name "coverprofile.out" -delete

clean-debug:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "debug" -o -name "__debug_bin*" -delete

clean-deploy:
	@cd $(ROOT_DIRECTORY) && rm -rf deploy

clean-test:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.test" -o -name "*.report" -delete

clean-all: clean

pre-commit: format imports vet

gopath-implode:
	cd $(REPOSITORY_GOPATH) && rm -rf bin pkg && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type f -delete && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type d -empty -delete

phony:
	@egrep '^[^ #]+:( |$$)' $(MAKEFILE) | sed -E 's/^([^ #]+):.*/\1/' | sort -u | grep -v '^.PHONY' | xargs echo '.PHONY:' | fold -s -w 80 | sed '$$!s/$$/\\/;2,$$s/^/    /g'

.PHONY: CompileDaemon bindir build build-list build-watch buildable \
    bundle-deploy ci-build ci-build-watch ci-deploy ci-docker ci-generate \
    ci-ginkgo-test ci-ginkgo-test-until-failure ci-ginkgo-test-watch ci-go-test \
    ci-test clean clean-all clean-bin clean-cover clean-debug clean-deploy \
    clean-test default deploy deploy-migrations deploy-services deploy-tools docker \
    docker-build docker-push format format-write format-write-changed generate \
    ginkgo ginkgo-test ginkgo-test-until-failure ginkgo-test-watch go-generate \
    go-test goimports golint gopath-implode imports imports-write \
    imports-write-changed lint lint-ignore mockgen phony plugin-visibility \
    plugin-visibility-private plugin-visibility-public plugins-visibility \
    plugins-visibility-private plugins-visibility-public pre-build pre-commit \
    service-build service-debug service-restart service-restart-all service-start \
    test tmp vet vet-ignore
