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
VERSION_PACKAGE:=$(REPOSITORY)/version

GO_LD_FLAGS:=-ldflags "-X $(VERSION_PACKAGE).Base=$(VERSION_BASE) -X $(VERSION_PACKAGE).ShortCommit=$(VERSION_SHORT_COMMIT) -X $(VERSION_PACKAGE).FullCommit=$(VERSION_FULL_COMMIT)"

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

ginkgo: check-environment
ifeq ($(shell which ginkgo),)
	mkdir -p $(GOPATH_REPOSITORY)/src/github.com/onsi/
	cp -r vendor/github.com/onsi/ginkgo $(GOPATH_REPOSITORY)/src/github.com/onsi/
	go install github.com/onsi/ginkgo/ginkgo
endif

buildable: goimports golint ginkgo

editable: buildable gocode godef

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
	@echo "go tool vet -all -shadow -shadowstrict"
	@cd $(ROOT_DIRECTORY) && \
		find . -mindepth 1 -maxdepth 1 -not -path "./.*" -not -path "./_*" -not -path "./vendor" -type d -exec go tool vet -all -shadow -shadowstrict {} \; 2> _tmp/govet.out > _tmp/govet.out && \
		O=`diff .govetignore _tmp/govet.out` || (echo "$${O}" && exit 1)

vet-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/govet.out .govetignore

lint: golint tmp
	@echo "golint"
	@cd $(ROOT_DIRECTORY) && \
		find . -not -path './vendor/*' -name '*.go' -type f -exec golint {} \; | grep -v 'exported.*should have comment.*or be unexported' 2> _tmp/golint.out > _tmp/golint.out && \
		diff .golintignore _tmp/golint.out || \
		exit 0

lint-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/golint.out .golintignore

pre-build: format imports vet lint

build-list:
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD)

build: check-environment
	@echo "go build"
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_MKDIR_CMD) | while read LINE; do $(MKDIR_CMD) $${LINE}; done
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_GO_BUILD_CMD) | while read LINE; do $(GO_BUILD_CMD) $${LINE}; done

ci-build: pre-build build

start: start-dataservices start-userservices

start-dataservices: stop-dataservices log
	@cd $(ROOT_DIRECTORY) && _bin/dataservices/dataservices >> _log/service.log 2>&1 &

start-userservices: stop-userservices log
	@cd $(ROOT_DIRECTORY) && _bin/userservices/userservices >> _log/service.log 2>&1 &

stop: stop-dataservices stop-userservices

stop-dataservices: check-environment
	@killall -v dataservices &> /dev/null || exit 0

stop-userservices: check-environment
	@killall -v userservices &> /dev/null || exit 0

test: ginkgo
	@echo "ginkgo --slowSpecThreshold=10 -r $(TEST)"
	@cd $(ROOT_DIRECTORY) && TIDEPOOL_ENV=test ginkgo --slowSpecThreshold=10 -r $(TEST)

ci-test: ginkgo
	@echo "ginkgo --slowSpecThreshold=10 -r --randomizeSuites --randomizeAllSpecs -succinct --failOnPending --cover --trace --race --progress -keepGoing $(TEST)"
	@cd $(ROOT_DIRECTORY) && TIDEPOOL_ENV=test ginkgo --slowSpecThreshold=10 -r --randomizeSuites --randomizeAllSpecs -succinct --failOnPending --cover --trace --race --progress -keepGoing $(TEST)

watch: ginkgo
	@echo "ginkgo watch --slowSpecThreshold=10 -r -notify $(WATCH)"
	@cd $(ROOT_DIRECTORY) && TIDEPOOL_ENV=test ginkgo watch --slowSpecThreshold=10 -r -notify $(WATCH)

deploy: clean-deploy deploy-dataservices deploy-userservices deploy-tools

deploy-dataservices:
	@$(MAKE) bundle-deploy DEPLOY=dataservices

deploy-userservices:
	@$(MAKE) bundle-deploy DEPLOY=userservices

deploy-tools:
	@$(MAKE) bundle-deploy DEPLOY=tools

ci-deploy: ci-build ci-test deploy

bundle-deploy: check-environment
ifdef DEPLOY
ifdef TRAVIS_TAG
	@cd $(ROOT_DIRECTORY) && \
		DEPLOY_TAG=$(DEPLOY)-$(TRAVIS_TAG) && \
		DEPLOY_DIR=deploy/$(DEPLOY)/$${DEPLOY_TAG} && \
		mkdir -p $${DEPLOY_DIR}/ && \
		cp -r _deploy/$(DEPLOY)/* $${DEPLOY_DIR}/ && \
		for DIR in _bin _config; do if [ -d "$${DIR}/$(DEPLOY)" ]; then mkdir -p $${DEPLOY_DIR}/$${DIR}; cp -r $${DIR}/$(DEPLOY)/ $${DEPLOY_DIR}/$${DIR}/$(DEPLOY)/; fi; done && \
		tar -c -z -f $${DEPLOY_DIR}.tar.gz -C deploy/$(DEPLOY)/ $${DEPLOY_TAG}
endif
endif

clean: clean-bin clean-cover clean-deploy
	@cd $(ROOT_DIRECTORY) && rm -rf _log _tmp

clean-bin: stop
	@cd $(ROOT_DIRECTORY) && rm -rf _bin

clean-cover:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.coverprofile" -delete

clean-deploy:
	@cd $(ROOT_DIRECTORY) && rm -rf deploy

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
