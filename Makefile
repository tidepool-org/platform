ROOT_DIRECTORY:=$(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
REPOSITORY:=$(ROOT_DIRECTORY:$(realpath $(ROOT_DIRECTORY)/../../../)/%=%)

ifdef TRAVIS_TAG
	VERSION_BASE:=$(TRAVIS_TAG)
else
	VERSION_BASE:=$(shell git describe --abbrev=0 --tags)
endif
VERSION_BASE:=$(VERSION_BASE:v%=%)
VERSION_SHORT_COMMIT:=$(shell git rev-parse --short HEAD)
VERSION_FULL_COMMIT:=$(shell git rev-parse HEAD)

GO_LD_FLAGS:=-ldflags "-X main.VersionBase=$(VERSION_BASE) -X main.VersionShortCommit=$(VERSION_SHORT_COMMIT) -X main.VersionFullCommit=$(VERSION_FULL_COMMIT)"

FIND_MAIN_CMD:=find . -path './$(BUILD)*' -not -path './vendor/*' -name '*.go' -not -name '*_test.go' -type f -exec egrep -l '^\s*package\s+main\s*$$' {} \;
TRANSFORM_MKDIR_CMD:=sed 's/\.\(.*\/\)[^\/]*\.go/_bin\1/'
MKDIR_CMD:=mkdir -p
TRANSFORM_GO_BUILD_CMD:=sed 's/\(\.\(\/.*\)\.go\)/_bin\2 \1/'
GO_BUILD_CMD:=go build $(GO_BUILD_FLAGS) $(GO_LD_FLAGS) -o

GOPATH_REPOSITORY:=$(word 1, $(subst :, ,$(GOPATH)))

default: test

log:
	@mkdir -p $(ROOT_DIRECTORY)/_log

tmp:
	@mkdir -p $(ROOT_DIRECTORY)/_tmp

check-gopath:
ifndef GOPATH
	@echo "FATAL: GOPATH environment variable not defined. Please see http://golang.org/doc/code.html#GOPATH."
	@exit 1
endif
	@exit 0

check-environment: check-gopath

godep: check-environment
ifeq ($(shell which godep),)
	go get -u github.com/tools/godep
endif

goimports: check-environment
ifeq ($(shell which goimports),)
	go get -u golang.org/x/tools/cmd/goimports
endif

golint: check-environment
ifeq ($(shell which golint),)
	go get -u github.com/golang/lint/golint
endif

gocode: check-environment
ifeq ($(shell which gocode),)
	go get -u github.com/nsf/gocode
endif

godef: check-environment
ifeq ($(shell which godef),)
	go get -u github.com/rogpeppe/godef
endif

oracle: check-environment
ifeq ($(shell which oracle),)
	go get -u golang.org/x/tools/cmd/oracle
endif

ginkgo: check-environment
ifeq ($(shell which ginkgo),)
	mkdir -p $(GOPATH_REPOSITORY)/src/github.com/onsi/
	cp -r vendor/github.com/onsi/ginkgo $(GOPATH_REPOSITORY)/src/github.com/onsi/
	go install github.com/onsi/ginkgo/ginkgo
endif

buildable: goimports golint ginkgo

editable: buildable gocode godef oracle

format: check-environment
	@echo "gofmt -d -e -s"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './vendor/*' -name '*.go' -type f -exec gofmt -d -e -s {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports: goimports
	@echo "goimports -d -e"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './vendor/*' -name '*.go' -type f -exec goimports -d -e {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

vet: check-environment tmp
	@echo "go tool vet -test -shadowstrict -printfuncs=Errorf:1"
	@cd $(ROOT_DIRECTORY) && \
		find . -mindepth 1 -maxdepth 1 -not -path "./.*" -not -path "./_*" -not -path "./vendor" -type d -exec go tool vet -test -shadowstrict -printfuncs=Errorf:1 {} \; &> _tmp/govet.out && \
		O=`diff .govetignore _tmp/govet.out` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

vet-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/govet.out .govetignore

lint: golint tmp
	@echo "golint"
	@cd $(ROOT_DIRECTORY) && \
		find . -not -path './vendor/*' -name '*.go' -type f -exec golint {} \; | grep -v 'exported.*should have comment.*or be unexported' &> _tmp/golint.out && \
		diff .golintignore _tmp/golint.out || \
		exit 0

lint-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/golint.out .golintignore

pre-build: format imports vet lint

build-list:
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD)

build:
	@echo "go build"
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_MKDIR_CMD) | xargs -L1 $(MKDIR_CMD)
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_GO_BUILD_CMD) | xargs -L1 $(GO_BUILD_CMD)

ci-build: build

ci-deploy: pre-build ci-build ci-test
ifdef TRAVIS_TAG
	@cd $(ROOT_DIRECTORY) && \
		rm -rf deploy/ && \
		mkdir -p deploy/platform/platform-$(TRAVIS_TAG)/ && \
		cp -r _bin _config _deploy/* deploy/platform/platform-$(TRAVIS_TAG)/ && \
		tar -c -z -f deploy/platform/platform-$(TRAVIS_TAG).tar.gz -C deploy/platform/ platform-$(TRAVIS_TAG)
endif

start: stop build log
	@cd $(ROOT_DIRECTORY) && _bin/dataservices/dataservices >> _log/service.log 2>&1 &

stop: check-environment
	@killall -v dataservices &> /dev/null || exit 0

test: ginkgo
	@echo "ginkgo --slowSpecThreshold=10 -r $(TEST)"
	@cd $(ROOT_DIRECTORY) && TIDEPOOL_ENV=test ginkgo --slowSpecThreshold=10 -r $(TEST)

ci-test: ginkgo
	@echo "ginkgo --slowSpecThreshold=10 -r --randomizeSuites --randomizeAllSpecs -succinct --failOnPending --cover --trace --race --progress -keepGoing $(TEST)"
	@cd $(ROOT_DIRECTORY) && TIDEPOOL_ENV=test ginkgo --slowSpecThreshold=10 -r --randomizeSuites --randomizeAllSpecs -succinct --failOnPending --cover --trace --race --progress -keepGoing $(TEST)

watch: ginkgo
	@echo "ginkgo watch --slowSpecThreshold=10 -r -notify $(WATCH)"
	@cd $(ROOT_DIRECTORY) && TIDEPOOL_ENV=test ginkgo watch --slowSpecThreshold=10 -r -notify $(WATCH)

clean: stop clean-cover
	@cd $(ROOT_DIRECTORY) && rm -rf _bin _log _tmp deploy

clean-cover:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.coverprofile" -delete

clean-all: clean

git-hooks:
	@echo "Installing git hooks..."
	@cd $(ROOT_DIRECTORY) && cp _tools/git/hooks/* .git/hooks/

pre-commit: format imports vet lint

# DO NOT USE THE FOLLOWING TARGETS UNDER NORMAL CIRCUMSTANCES!!!

# Remove everything in GOPATH_REPOSITORY except REPOSITORY
gopath-implode: check-environment
	cd $(GOPATH_REPOSITORY) && rm -rf {bin,pkg} && find src -not -path "src/$(REPOSITORY)/*" -type f -delete && find src -not -path "src/$(REPOSITORY)/*" -type d -empty -delete

# Remove saved dependencies in REPOSITORY
dependencies-implode: check-environment
	cd $(ROOT_DIRECTORY) && rm -rf {Godeps,vendor}

bootstrap-implode: gopath-implode dependencies-implode

bootstrap-dependencies: godep
	go get github.com/onsi/ginkgo
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/ginkgo/extensions/table
	go get github.com/onsi/gomega
	go get github.com/onsi/gomega/ghttp
	go get golang.org/x/sys/unix
	go get ./...

bootstrap-save: bootstrap-dependencies
	cd $(ROOT_DIRECTORY) && godep save ./... github.com/onsi/ginkgo/ginkgo github.com/onsi/ginkgo/extensions/table

# Bootstrap REPOSITORY with initial dependencies
bootstrap:
	@$(MAKE) bootstrap-implode
	@$(MAKE) bootstrap-save
	@$(MAKE) gopath-implode

.PHONY: default log tmp check-gopath check-environment \
	godep goimports golint gocode godef oracle ginkgo buildable editable \
	format imports vet vet-ignore lint lint-ignore pre-build build-list build ci-build ci-deploy start stop test ci-test watch clean clean-cover clean-all git-hooks pre-commit \
	gopath-implode dependencies-implode bootstrap-implode bootstrap-dependencies bootstrap-save bootstrap
