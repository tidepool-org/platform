ROOT_DIRECTORY:=$(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
REPOSITORY:=$(ROOT_DIRECTORY:$(realpath $(ROOT_DIRECTORY)/../../../)/%=%)

VERSION_BASE=$(shell cat .version)
VERSION_COMMIT=$(shell git rev-parse HEAD)

GO_LD_FLAGS:=-ldflags "-X $(REPOSITORY)/version.Base=$(VERSION_BASE) -X $(REPOSITORY)/version.Commit=$(VERSION_COMMIT)"

MAIN_FIND_CMD:=find . -not -path './Godeps/*' -name '*.go' -type f -exec egrep -l '^\s*func\s+main\s*\(' {} \;
MAIN_TRANSFORM_CMD:=sed 's/\(.*\/\([^\/]*\)\.go\)/_bin\/\2 \1/'
GO_BUILD_CMD:=godep go build $(GO_BUILD_FLAGS) $(GO_LD_FLAGS) -o

default: test

log:
	@mkdir -p $(ROOT_DIRECTORY)/_log

tmp:
	@mkdir -p $(ROOT_DIRECTORY)/_tmp

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

check-environment: check-go check-gopath

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

# Use godep to install ginkgo
ginkgo: godep
ifeq ($(shell which ginkgo),)
	godep go install github.com/onsi/ginkgo/ginkgo
endif

buildable: godep goimports golint ginkgo

editable: buildable gocode godef oracle

format: check-environment
	@echo "gofmt -d -e -s"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './Godeps/*' -name '*.go' -type f -exec gofmt -d -e -s {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports: goimports
	@echo "goimports -d -e"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './Godeps/*' -name '*.go' -type f -exec goimports -d -e {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

vet: check-environment
	@echo "go tool vet -test"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -mindepth 1 -maxdepth 1 -not -path "./.*" -not -path "./_*" -not -path "./Godeps" -type d -exec go tool vet -test {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

lint: golint tmp
	@echo "golint"
	cat .golintignore && echo $?
	ls -al _tmp && echo $?
	cd $(ROOT_DIRECTORY) && find . -not -path './Godeps/*' -name '*.go' -type f -exec golint {} \; | grep -v 'exported.*should have comment.*or be unexported' &> _tmp/golint.out && echo $?
	cat _tmp/golint.out && echo $?
	diff --changed-group-format='%>' --unchanged-group-format='' .golintignore _tmp/golint.out && echo $?
	ls -al _tmp && echo $?

lint-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/golint.out .golintignore

pre-build: format imports vet lint

build: godep
	@echo "godep go build"
	@cd $(ROOT_DIRECTORY) && mkdir -p _bin && $(MAIN_FIND_CMD) | $(MAIN_TRANSFORM_CMD) | xargs -L1 $(GO_BUILD_CMD)

start: stop build log
	@cd $(ROOT_DIRECTORY) && _bin/dataservices >> _log/service.log 2>&1 &
	@cd $(ROOT_DIRECTORY) && _bin/userservices >> _log/service.log 2>&1 &

stop: check-environment
	@killall -v dataservices userservices &> /dev/null || exit 0

test: ginkgo
	@echo "ginkgo -r"
	@cd $(ROOT_DIRECTORY) && GOPATH=$(shell godep path):$(GOPATH) ginkgo -r $(TEST)

watch: ginkgo
	@echo "ginkgo watch -r -p -randomizeAllSpecs -succinct -notify"
	@cd $(ROOT_DIRECTORY) && GOPATH=$(shell godep path):$(GOPATH) ginkgo watch -r -p -randomizeAllSpecs -succinct -notify $(WATCH)

clean: stop
	@cd $(ROOT_DIRECTORY) && rm -rf _bin _log _tmp

clean-all: clean
	@cd $(ROOT_DIRECTORY) && rm -rf Godeps/_workspace/{bin,pkg}

git-hooks:
	@echo "Installing git hooks..."
	@cd $(ROOT_DIRECTORY) && cp _tools/git/hooks/* .git/hooks/

pre-commit: format imports vet lint

# DO NOT USE THE FOLLOWING TARGETS UNDER NORMAL CIRCUMSTANCES!!!

# Remove everything in GOPATH except REPOSITORY
gopath-implode: check-environment
	cd $(GOPATH) && rm -rf {bin,pkg} && find src -not -path "src/$(REPOSITORY)/*" -type f -delete && find src -not -path "src/$(REPOSITORY)/*" -type d -empty -delete

# Remove saved dependencies in REPOSITORY
dependencies-implode: check-environment
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

.PHONY: default log tmp check-go check-gopath check-environment \
	godep goimports golint gocode godef oracle ginkgo buildable editable \
	format imports vet lint pre-build build start stop test watch clean clean-all git-hooks pre-commit \
	gopath-implode dependencies-implode bootstrap-implode bootstrap-dependencies bootstrap-save bootstrap
