# variables that have to be setup for Docker
# DOCKER_REGISTRY
# DOCKER_USERNAME
# DOCKER_PASSWORD
# OPS_DOCKER_REGISTRY
# OPS_DOCKER_USERNAME
# OPS_DOCKER_PASSWORD
# VERSION

ROOT_DIRECTORY:=$(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

REPOSITORY_GOPATH:=$(word 1, $(subst :, ,$(GOPATH)))
REPOSITORY_PACKAGE:=github.com/tidepool-org/platform
REPOSITORY_NAME:=$(notdir $(REPOSITORY_PACKAGE))

BIN_DIRECTORY := ${ROOT_DIRECTORY}/_bin
PATH := ${PATH}:${BIN_DIRECTORY}

ifeq ($(VERSION),)
  VERSION:=0.0.0
endif
VERSION_SHORT_COMMIT:=$(shell git rev-parse --short HEAD)
VERSION_FULL_COMMIT:=$(shell git rev-parse HEAD)
VERSION_PACKAGE:=$(REPOSITORY_PACKAGE)/application

GO_LD_FLAGS:=-ldflags '-X $(VERSION_PACKAGE).VersionBase=$(VERSION) -X $(VERSION_PACKAGE).VersionShortCommit=$(VERSION_SHORT_COMMIT) -X $(VERSION_PACKAGE).VersionFullCommit=$(VERSION_FULL_COMMIT)'

FIND_MAIN_CMD:=find . -path './$(BUILD)*' -not -path './vendor/*' -name '*.go' -not -name '*_test.go' -type f -exec egrep -l '^\s*func\s+main\s*(\s*)' {} \;
TRANSFORM_GO_BUILD_CMD:=sed 's|\.\(.*\)\(/[^/]*\)/[^/]*|_bin\1\2\2 .\1\2/.|'
GO_BUILD_CMD:=go build $(GO_BUILD_FLAGS) $(GO_LD_FLAGS) -o

ifeq ($(TRAVIS_BRANCH),dblp)
ifeq ($(TRAVIS_PULL_REQUEST_BRANCH),)
	DOCKER:=true
endif
else ifdef TRAVIS_TAG
	DOCKER:=true
endif
ifdef DOCKER
	DOCKER_REPOSITORY:="${DOCKER_REGISTRY}/$(REPOSITORY_NAME)-$(patsubst .%,%,$(suffix $(DOCKER_FILE)))"
ifdef OPS_DOCKER_REGISTRY
	OPS_DOCKER_REPOSITORY:="${OPS_DOCKER_REGISTRY}/$(REPOSITORY_NAME)-$(patsubst .%,%,$(suffix $(DOCKER_FILE)))"
endif
endif

default: test

tmp:
	@mkdir -p $(ROOT_DIRECTORY)/_tmp

bindir:
	@mkdir -p $(ROOT_DIRECTORY)/_bin

CompileDaemon:
ifeq ($(shell which CompileDaemon),)
	go install github.com/githubnemo/CompileDaemon
endif

esc:
ifeq ($(shell which esc),)
	go install github.com/mjibson/esc
endif

ginkgo:
ifeq ($(shell which ginkgo),)
	go install github.com/onsi/ginkgo/ginkgo
endif

goimports:
ifeq ($(shell which goimports),)
	go install golang.org/x/tools/cmd/goimports
endif

golint:
ifeq ($(shell which golint),)
	go install golang.org/x/lint/golint
endif

buildable: export GOBIN = ${BIN_DIRECTORY}
buildable: bindir CompileDaemon esc ginkgo goimports golint

generate: esc
	@echo "go generate ./..."
	@cd $(ROOT_DIRECTORY) && go generate ./...

ci-generate: generate
	@cd $(ROOT_DIRECTORY) && \
		O=`git diff` && [ "$${O}" = "" ] || (echo "$${O}" && exit 1)

format:
	@echo "gofmt -d -e -s"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './vendor/*' -name '*.go' -type f -exec gofmt -d -e -s {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

format-write:
	@echo "gofmt -e -s -w"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './vendor/*' -name '*.go' -type f -exec gofmt -e -s -w {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports: goimports
	@echo "goimports -d -e -local 'github.com/tidepool-org/platform'"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './vendor/*' -name '*.go' -type f -exec goimports -d -e -local 'github.com/tidepool-org/platform' {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

imports-write: goimports
	@echo "goimports -e -w -local 'github.com/tidepool-org/platform'"
	@cd $(ROOT_DIRECTORY) && \
		O=`find . -not -path './vendor/*' -name '*.go' -type f -exec goimports -e -w -local 'github.com/tidepool-org/platform' {} \; 2>&1` && \
		[ -z "$${O}" ] || (echo "$${O}" && exit 1)

vet: tmp
	@echo "go vet"
	cd $(ROOT_DIRECTORY) && \
		go vet ./... 2> _tmp/govet.out > _tmp/govet.out && \
		O=`diff -w .govetignore _tmp/govet.out` || (echo "$${O}" && exit 1)

vet-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/govet.out .govetignore

lint: golint tmp
	@echo "golint"
	@cd $(ROOT_DIRECTORY) && \
		find . -not -path './vendor/*' -name '*.go' -type f | sort -d | xargs -I {} golint {} | grep -v 'exported.*should have comment.*or be unexported' 2> _tmp/golint.out > _tmp/golint.out || [ $${?} == 1 ] && \
		diff .golintignore _tmp/golint.out || \
		exit 0

lint-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/golint.out .golintignore

pre-build: format imports vet lint

build-list:
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD)

build:
	@echo "go build $(BUILD)"
	@go mod tidy && cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_GO_BUILD_CMD) | while read LINE; do $(GO_BUILD_CMD) $${LINE}; done

build-watch: CompileDaemon
	@cd $(ROOT_DIRECTORY) && BUILD=$(BUILD) CompileDaemon -build-dir='.' -build='make build' -color -directory='.' -exclude-dir='.git' -exclude='*_test.go' -include='Makefile' -recursive=true

ci-build: pre-build build

ci-build-watch: CompileDaemon
	@cd $(ROOT_DIRECTORY) && BUILD=$(BUILD) CompileDaemon -build-dir='.' -build='make ci-build' -color -directory='.' -exclude-dir='.git' -include='Makefile' -recursive=true

service-build:
ifeq ($(TARGETPLATFORM),linux/arm64)
	export GOOS=darwin && export GOARCH=arm64 && export CGO_ENABLED=0
else
	export CGO_ENABLED=1
endif
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

test: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 -r $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 -r $(TEST)

ci-test: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 --compilers=2 -r -randomizeSuites -randomizeAllSpecs -failOnPending -race -timeout=8m --reportFile=junit-report/report.xml $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 --compilers=2 -r -randomizeSuites -randomizeAllSpecs -failOnPending -race -timeout=8m --reportFile=junit-report/report.xml $(TEST)


ci-soups: clean-soup-doc generate-soups

generate-soups:
	@cd $(ROOT_DIRECTORY) && \
		$(MAKE) service-soup SERVICE_DIRECTORY=data SERVICE=platform TARGET=soup

service-soup:
	@cd ${SERVICE_DIRECTORY} && \
		echo "# SOUPs List for ${SERVICE}@${VERSION}" > soup.md && \
		go list -f '## {{printf "%s \n\t* description: \n\t* version: %s\n\t* webSite: https://%s\n\t* sources:" .Path .Version .Path}}' -m all >> soup.md && \
		mkdir -p $(ROOT_DIRECTORY)/${TARGET}/${SERVICE} && \
		mv soup.md $(ROOT_DIRECTORY)/${TARGET}/${SERVICE}/${SERVICE}-${VERSION}-soup.md


clean: clean-bin clean-cover clean-debug clean-deploy
	@cd $(ROOT_DIRECTORY) && rm -rf _tmp

clean-bin:
	@cd $(ROOT_DIRECTORY) && rm -rf _bin

clean-cover:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "*.coverprofile" -delete

clean-debug:
	@cd $(ROOT_DIRECTORY) && find . -type f -name "debug" -delete

clean-deploy:
	@cd $(ROOT_DIRECTORY) && rm -rf deploy

clean-soup-doc:
	@cd $(ROOT_DIRECTORY) && rm -rf soup

clean-all: clean

pre-commit: format imports vet lint

gopath-implode:
	cd $(REPOSITORY_GOPATH) && rm -rf bin pkg && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type f -delete && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type d -empty -delete

.PHONY: default tmp bindir CompileDaemon esc ginkgo goimports golint buildable \
	format format-write imports vet vet-ignore lint lint-ignore pre-build build-list build ci-build \
	service-build service-start service-restart service-restart-all test test-watch ci-test c-test-watch \
	clean clean-bin clean-cover clean-debug clean-deploy clean-all pre-commit \
	gopath-implode