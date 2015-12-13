REPOSITORY:=github.com/tidepool-org/platform

ROOT_DIRECTORY:=$(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

# TODO: Need to make this work
VERSION_STRING:=0.0.1

GO_LD_FLAGS:=-ldflags "-X github.com/tidepool-org/platform/version.String=$(VERSION_STRING)"

# Commands for later use
MAIN_FIND_CMD:=find . -type f -name '*.go' -not -path './Godeps/*' -exec egrep -l '^\s*func\s+main\s*\(' {} \;
MAIN_TRANSFORM_CMD:=sed 's/\(.*\/\([^\/]*\)\.go\)/bin\/\2 \1/'
GO_BUILD_CMD:=godep go build $(GO_BUILD_FLAGS) $(GO_LD_FLAGS) -o

default: build test

check-go:
ifeq ($(GO15VENDOREXPERIMENT), 1)
	@echo "FATAL: GO15VENDOREXPERIMENT not supported."
	@exit 1
endif
	@exit 0

check-gopath:
ifndef GOPATH
	@echo "FATAL: GOPATH environment variable not defined. Please see http://golang.org/doc/code.html#GOPATH."
	@exit 1
endif
	@exit 0

check-env: check-go check-gopath

godep: check-env
	@[ "$(shell which godep)" ] || go get -u github.com/tools/godep

goimports: check-env
	@[ "$(shell which goimports)" ] || go get -u golang.org/x/tools/cmd/goimports

golint: check-env
	@[ "$(shell which golint)" ] || go get -u github.com/golang/lint/golint

gocode: check-env
	@[ "$(shell which gocode)" ] || go get -u github.com/nsf/gocode

godef: check-env
	@[ "$(shell which godef)" ] || go get -u github.com/rogpeppe/godef

oracle: check-env
	@[ "$(shell which oracle)" ] || go get -u golang.org/x/tools/cmd/oracle

buildable: godep goimports golint

editable: buildable gocode godef oracle

ginkgo: godep
	@[ "$(shell which ginkgo)" ] || cd $(ROOT_DIRECTORY) && godep go install github.com/onsi/ginkgo/ginkgo

imports: goimports
	@cd $(ROOT_DIRECTORY) && find . -type f -name '*.go' -not -path './Godeps/*' -exec goimports -d -e {} \;

format: imports

lint: golint
	@cd $(ROOT_DIRECTORY) && golint ./...

precommit: imports lint

build: godep
	@cd $(ROOT_DIRECTORY) && mkdir -p bin && $(MAIN_FIND_CMD) | $(MAIN_TRANSFORM_CMD) | xargs -L1 $(GO_BUILD_CMD)

test: ginkgo
	@cd $(ROOT_DIRECTORY) && GOPATH=$(shell godep path):$(GOPATH) ginkgo -r $(TEST)

watch: check-env
	@cd $(ROOT_DIRECTORY) && GOPATH=$(shell godep path):$(GOPATH) ginkgo watch -r -p -randomizeAllSpecs -succinct -notify $(WATCH)

clean: check-env
	@cd $(ROOT_DIRECTORY) && rm -rf bin

clean-all: clean
	@cd $(ROOT_DIRECTORY) && rm -rf Godeps/_workspace/{bin,pkg}

# DO NOT USE THE FOLLOWING TARGETS UNDER NORMAL CIRCUMSTANCES!!!

# Remove everything in GOPATH except REPOSITORY
gopath-implode: check-env
	cd $(GOPATH) && rm -rf {bin,pkg} && find src -type f -not -path "src/$(REPOSITORY)/*" -delete && find src -type d -not -path "src/$(REPOSITORY)/*" -empty -delete

# Remove saved dependencies in REPOSITORY
dependencies-implode: check-env
	cd $(ROOT_DIRECTORY) && rm -rf {Godeps,vendor}

bootstrap-implode: gopath-implode dependencies-implode

bootstrap-dependencies: godep
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega

bootstrap-save: bootstrap-dependencies
	cd $(ROOT_DIRECTORY) && godep save ./...

# Bootstrap REPOSITORY with initial dependencies
bootstrap: 
	@$(MAKE) bootstrap-implode 
	@$(MAKE) bootstrap-save 
	@$(MAKE) gopath-implode

.PHONY: default check-go check-gopath check-repository check-env \
	godep goimports golint gocode godef oracle buildable editable ginkgo \
	imports format lint precommit build test watch clean clean-all \
	gopath-implode dependencies-implode bootstrap-implode bootstrap-dependencies bootstrap-save bootstrap
