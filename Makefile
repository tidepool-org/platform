TIMESTAMP ?= $(shell date +%s)
# ensure that we use the same timestamps in sub-makes. We've seen cases where
# these can vary by 1 second
export TIMESTAMP

ifneq ($(PRIVATE),)
  REPOSITORY_SUFFIX:=-private
endif

SERVICES_SEPARATOR=,
SERVICES_TO_BUILD?=auth,blob,data,migrations,prescription,task,tools
SERVICES_TO_BUILD:=$(subst $(SERVICES_SEPARATOR), ,$(SERVICES_TO_BUILD))

ROOT_DIRECTORY:=$(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

REPOSITORY_GOPATH:=$(word 1, $(subst :, ,$(GOPATH)))
REPOSITORY_PACKAGE:=github.com/tidepool-org/platform
REPOSITORY_NAME:=$(notdir $(REPOSITORY_PACKAGE))$(REPOSITORY_SUFFIX)

BIN_DIRECTORY := ${ROOT_DIRECTORY}/_bin
PATH := ${PATH}:${BIN_DIRECTORY}

VERSION_BASE:=platform
VERSION_SHORT_COMMIT:=$(shell git rev-parse --short HEAD || echo "dev")
VERSION_FULL_COMMIT:=$(shell git rev-parse HEAD || echo "dev")
VERSION_PACKAGE:=$(REPOSITORY_PACKAGE)/application

GO_BUILD_FLAGS:=-buildvcs=false
GO_LD_FLAGS:=-ldflags '-X $(VERSION_PACKAGE).VersionBase=$(VERSION_BASE) -X $(VERSION_PACKAGE).VersionShortCommit=$(VERSION_SHORT_COMMIT) -X $(VERSION_PACKAGE).VersionFullCommit=$(VERSION_FULL_COMMIT)'

FIND_MAIN_CMD:=find . -path './$(BUILD)*' -not -path './.gvm_local/*' -name '*.go' -not -name '*_test.go' -type f -exec egrep -l '^\s*func\s+main\s*(\s*)' {} \;
TRANSFORM_GO_BUILD_CMD:=sed 's|\.\(.*\)\(/[^/]*\)/[^/]*|_bin\1\2\2 .\1\2/.|'

GO_BUILD_CMD:=go build $(GO_BUILD_FLAGS) $(GO_LD_FLAGS) -o

GINKGO_FLAGS += --require-suite --poll-progress-after=10s --poll-progress-interval=20s -r
GINKGO_CI_WATCH_FLAGS += --randomize-all --succinct --fail-on-pending --cover --trace --race
GINKGO_CI_FLAGS += $(GINKGO_CI_WATCH_FLAGS) --randomize-suites --keep-going

GOTEST_PKGS ?= ./...
GOTEST_FLAGS ?=

TIMING_CMD ?=

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
	DOCKER_REPOSITORY:=tidepool/$(REPOSITORY_NAME)-$(SERVICE_NAME)
endif

default: test

tmp:
	@mkdir -p $(ROOT_DIRECTORY)/_tmp

bindir:
	@mkdir -p $(ROOT_DIRECTORY)/_bin

CompileDaemon:
ifeq ($(shell which CompileDaemon),)
	@cd $(ROOT_DIRECTORY) && \
		echo "go install github.com/githubnemo/CompileDaemon@v1.4.0" && \
		go install github.com/githubnemo/CompileDaemon@v1.4.0
endif

mockgen:
ifeq ($(shell which mockgen),)
	@cd $(ROOT_DIRECTORY) && \
		echo "go install go.uber.org/mock/mockgen@v0.5.0" && \
		go install go.uber.org/mock/mockgen@v0.5.0
endif

ginkgo:
ifeq ($(shell which ginkgo),)
	@cd $(ROOT_DIRECTORY) && \
		echo "github.com/onsi/ginkgo/v2/ginkgo@v2.19.0" && \
		go install github.com/onsi/ginkgo/v2/ginkgo@v2.19.0
endif

goimports:
ifeq ($(shell which goimports),)
	@cd $(ROOT_DIRECTORY) && \
		echo "golang.org/x/tools/cmd/goimports@latest" && \
		go install golang.org/x/tools/cmd/goimports@latest
endif

buildable: export GOBIN = ${BIN_DIRECTORY}
buildable: bindir CompileDaemon ginkgo goimports

generate: mockgen
	@echo "go generate ./..."
	@cd $(ROOT_DIRECTORY) && go generate ./...

ci-generate: generate format-write-changed imports-write-changed
	@cd $(ROOT_DIRECTORY) && \
		O=`git diff` && [ "$${O}" = "" ] || (echo "$${O}" && exit 1)

format:
	@echo "gofmt -d -e -s"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -name '*.go' -type f -exec gofmt -d -e -s {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format-write:
	@echo "gofmt -e -s -w"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -name '*.go' -type f -exec gofmt -e -s -w {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format-write-changed:
	@cd $(ROOT_DIRECTORY) && \
		git diff --name-only | grep '\.go$$' | xargs -I{} gofmt -e -s -w {}

imports: goimports
	@echo "goimports -d -e -local 'github.com/tidepool-org/platform'"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -name '*.go' -type f -exec goimports -d -e -local 'github.com/tidepool-org/platform' {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports-write: goimports
	@echo "goimports -e -w -local 'github.com/tidepool-org/platform'"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './.gvm_local/*' -name '*.go' -type f -exec goimports -e -w -local 'github.com/tidepool-org/platform' {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports-write-changed: goimports
	@cd $(ROOT_DIRECTORY) && \
		git diff --name-only | grep '\.go$$' | xargs -I{} goimports -e -w -local 'github.com/tidepool-org/platform' {}

vet: tmp
	@echo "go vet"
	@cd $(ROOT_DIRECTORY) && \
		go vet ./... > _tmp/govet.out 2>&1 || \
		(diff .govetignore _tmp/govet.out && exit 1)

vet-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/govet.out .govetignore

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

ginkgo-test: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo $(GINKGO_FLAGS) $(TEST)

test-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo $(GINKGO_FLAGS) -untilItFails $(TEST)

test-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch $(GINKGO_FLAGS) $(TEST)

ci-test: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && $(TIMING_CMD) ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) $(TEST)

ci-test-until-failure: ginkgo
	@echo "ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo $(GINKGO_FLAGS) $(GINKGO_CI_FLAGS) -untilItFails $(TEST)

ci-test-watch: ginkgo
	@echo "ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch $(GINKGO_FLAGS) $(GINKGO_CI_WATCH_FLAGS) $(TEST)

go-test:
	. ./env.test.sh && $(TIMING_CMD) go test $(GOTEST_FLAGS) $(GOTEST_PKGS)

go-ci-test: GOTEST_FLAGS += -count=1 -race -shuffle=on -cover
go-ci-test: GOTEST_PKGS = ./...
go-ci-test: go-test

docker:
ifdef DOCKER
	@echo "$(DOCKER_PASSWORD)" | docker login --username "$(DOCKER_USERNAME)" --password-stdin
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-build DOCKER_FILE="$${DOCKER_FILE}" TIMESTAMP="$(TIMESTAMP)";done
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-push DOCKER_FILE="$${DOCKER_FILE}" TIMESTAMP="$(TIMESTAMP)";done
endif

docker-build:
ifdef DOCKER
ifdef DOCKER_FILE
ifdef BUILD_SERVICE
	docker build --tag $(DOCKER_REPOSITORY):development --target=development --file "$(DOCKER_FILE)" .
	docker build --tag $(DOCKER_REPOSITORY) --file "$(DOCKER_FILE)" .
ifdef TRAVIS_BRANCH
ifdef TRAVIS_COMMIT
ifdef TRAVIS_PULL_REQUEST_BRANCH
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
else
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-latest
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
endif
endif
endif
ifdef TRAVIS_TAG
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(TRAVIS_TAG:v%=%)
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
	docker push $(DOCKER_REPOSITORY)
endif
endif
ifdef TRAVIS_BRANCH
ifdef TRAVIS_COMMIT
ifdef TRAVIS_PULL_REQUEST_BRANCH
	docker push $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	docker push $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
else
	docker push $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	docker push $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-latest
	docker push $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)-$(TIMESTAMP)
endif
endif
endif
ifdef TRAVIS_TAG
	docker push $(DOCKER_REPOSITORY):$(TRAVIS_TAG:v%=%)
endif
endif
endif
endif

ci-docker: docker

clean: clean-bin clean-cover clean-debug clean-test
	@cd $(ROOT_DIRECTORY) && rm -rf _tmp

clean-bin:
	@cd $(ROOT_DIRECTORY) && rm -rf _bin _log

clean-cover:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.coverprofile" -o -name "coverprofile.out" -delete

clean-debug:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "debug" -o -name "__debug_bin*" -delete

clean-test:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.test" -o -name "*.report" -delete

clean-all: clean

pre-commit: format imports vet

gopath-implode:
	cd $(REPOSITORY_GOPATH) && rm -rf bin pkg && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type f -delete && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type d -empty -delete

.PHONY: default tmp bindir CompileDaemon ginkgo goimports buildable \
	format format-write imports vet vet-ignore pre-build build-list build ci-build \
	service-build service-start service-restart service-restart-all test test-watch ci-test c-test-watch \
	docker docker-build docker-push ci-docker \
	clean clean-bin clean-cover clean-debug clean-all pre-commit \
	gopath-implode go-test go-ci-test
