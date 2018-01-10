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

ifeq ($(TRAVIS_BRANCH),master)
	DOCKER:=true
else ifdef TRAVIS_TAG
	DOCKER:=true
endif
ifdef DOCKER_FILE
	DOCKER_REPO:="tidepool/$(patsubst .%,%,$(suffix $(DOCKER_FILE)))"
endif

default: test

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

CompileDaemon: check-environment
ifeq ($(shell which CompileDaemon),)
	go get -u github.com/tidepool-org/CompileDaemon
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
		find . -not -path './vendor/*' -name '*.go' -type f | sort | xargs -I {} golint {} | grep -v 'exported.*should have comment.*or be unexported' 2> _tmp/golint.out > _tmp/golint.out && \
		diff .golintignore _tmp/golint.out || \
		exit 0

lint-ignore:
	@cd $(ROOT_DIRECTORY) && cp _tmp/golint.out .golintignore

pre-build: format imports vet lint

build-list:
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD)

build: check-environment
	@echo "go build $(BUILD)"
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_MKDIR_CMD) | while read LINE; do $(MKDIR_CMD) $${LINE}; done
	@cd $(ROOT_DIRECTORY) && $(FIND_MAIN_CMD) | $(TRANSFORM_GO_BUILD_CMD) | while read LINE; do $(GO_BUILD_CMD) $${LINE}; done

ci-build: pre-build build

service-build:
ifdef SERVICE
	@$(MAKE) build BUILD=$${SERVICE}
endif

service-start: CompileDaemon tmp
ifdef SERVICE
	@cd $(ROOT_DIRECTORY) && BUILD=$(SERVICE) CompileDaemon -build-dir='.' -build='make build' -command='_bin/$(SERVICE)/$(notdir $(SERVICE))' -directory='_tmp' -pattern='^$$' -include='$(subst /,.,$(SERVICE)).restart' -recursive=false -log-prefix=false -graceful-kill=true -graceful-timeout=60
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

test-watch: ginkgo
	@echo "ginkgo watch -requireSuite -slowSpecThreshold=10 -r $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch -requireSuite -slowSpecThreshold=10 -r $(TEST)

ci-test: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 -r -randomizeSuites -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress -keepGoing $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 -r -randomizeSuites -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress -keepGoing $(TEST)

ci-test-watch: ginkgo
	@echo "ginkgo watch -requireSuite -slowSpecThreshold=10 -r -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch -requireSuite -slowSpecThreshold=10 -r -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress $(TEST)

deploy: clean-deploy deploy-services deploy-migrations deploy-tools

deploy-services:
	@cd $(ROOT_DIRECTORY) && for SERVICE in $(shell ls -1 _bin/services); do $(MAKE) bundle-deploy DEPLOY=$${SERVICE} SOURCE=services/$${SERVICE}; done

deploy-migrations:
	@$(MAKE) bundle-deploy DEPLOY=migrations SOURCE=migrations

deploy-tools:
	@$(MAKE) bundle-deploy DEPLOY=tools SOURCE=tools

ci-deploy: deploy

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

docker:
ifdef DOCKER
	@echo "$(DOCKER_PASSWORD)" | docker login --username "$(DOCKER_USERNAME)" --password-stdin
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-build DOCKER_FILE="$${DOCKER_FILE}"; done
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-push DOCKER_FILE="$${DOCKER_FILE}"; done
endif

docker-build:
ifdef DOCKER
ifdef DOCKER_FILE
	@docker build --tag "$(DOCKER_REPO):development" --target=development --file "$(DOCKER_FILE)" .
	@docker build --tag "$(DOCKER_REPO)" --file "$(DOCKER_FILE)" .
ifdef TRAVIS_TAG
	@docker tag "$(DOCKER_REPO)" "$(DOCKER_REPO):$(TRAVIS_TAG:v%=%)"
endif
endif
endif

docker-push:
ifdef DOCKER
ifdef DOCKER_REPO
ifeq ($(TRAVIS_BRANCH),master)
	@docker push "$(DOCKER_REPO):development"
	@docker push "$(DOCKER_REPO)"
endif
ifdef TRAVIS_TAG
	@docker push "$(DOCKER_REPO):$(TRAVIS_TAG:v%=%)"
endif
endif
endif

ci-docker: docker

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
	go get github.com/onsi/ginkgo/extensions/table
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega
	go get github.com/onsi/gomega/gbytes
	go get github.com/onsi/gomega/gexec
	go get github.com/onsi/gomega/ghttp
	go get github.com/onsi/gomega/gstruct
	go get golang.org/x/sys/unix
	go get ./...

bootstrap-save: bootstrap-dependencies
	cd $(ROOT_DIRECTORY) && godep save ./... \
		github.com/onsi/ginkgo/extensions/table \
		github.com/onsi/ginkgo/ginkgo \
		github.com/onsi/gomega/gbytes \
		github.com/onsi/gomega/gexec \
		github.com/onsi/gomega/ghttp \
		github.com/onsi/gomega/gstruct

# Bootstrap REPOSITORY with initial dependencies
bootstrap:
	@$(MAKE) bootstrap-implode
	@$(MAKE) bootstrap-save
	@$(MAKE) gopath-implode

.PHONY: default tmp check-gopath check-environment \
	godep goimports golint gocode godef CompileDaemon ginkgo buildable editable \
	format format-write imports vet vet-ignore lint lint-ignore pre-build build-list build ci-build \
	service-build service-start service-restart service-restart-all test test-watch ci-test c-test-watch \
	deploy deploy-services deploy-migrations deploy-tools ci-deploy bundle-deploy \
	docker docker-build docker-push ci-docker \
	clean clean-bin clean-cover clean-debug clean-deploy clean-all pre-commit \
	gopath-implode dependencies-implode bootstrap-implode bootstrap-dependencies bootstrap-save bootstrap
