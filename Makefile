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
VERSION_PACKAGE:=$(REPOSITORY)/application/version

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

format-write: check-environment
	@echo "gofmt -e -s -w"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './vendor/*' -name '*.go' -type f -exec gofmt -e -s -w {} \; 2>&1` && \
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

test: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 -r $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 -r $(TEST)

ci-test: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 -r -randomizeSuites -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress -keepGoing $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 -r -randomizeSuites -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress -keepGoing $(TEST)

watch: ginkgo
	@echo "ginkgo watch -requireSuite -slowSpecThreshold=10 -r $(WATCH)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch -requireSuite -slowSpecThreshold=10 -r $(WATCH)

deploy: clean-deploy deploy-services deploy-migrations deploy-tools

deploy-services:
	@for SERVICE in $(shell ls -1 $(ROOT_DIRECTORY)/_bin/services); do $(MAKE) bundle-deploy DEPLOY=$${SERVICE} SOURCE=services/$${SERVICE}; done

deploy-migrations:
	@$(MAKE) bundle-deploy DEPLOY=migrations SOURCE=migrations

deploy-tools:
	@$(MAKE) bundle-deploy DEPLOY=tools SOURCE=tools

ci-deploy: ci-build ci-test deploy

bundle-deploy: check-environment
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

clean: clean-bin clean-cover clean-deploy
	@cd $(ROOT_DIRECTORY) && rm -rf _log _tmp

clean-bin:
	@cd $(ROOT_DIRECTORY) && rm -rf _bin

clean-cover:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.coverprofile" -delete

clean-debug:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "debug" -delete

clean-deploy:
	@cd $(ROOT_DIRECTORY) && rm -rf deploy

clean-all: clean

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
	godep goimports golint gocode godef ginkgo buildable editable \
	format format-write imports vet vet-ignore lint lint-ignore pre-build build-list build ci-build \
	test ci-test watch \
	deploy deploy-services deploy-migrations deploy-tools ci-deploy bundle-deploy \
	clean clean-bin clean-cover clean-deploy clean-all pre-commit \
	gopath-implode dependencies-implode bootstrap-implode bootstrap-dependencies bootstrap-save bootstrap
