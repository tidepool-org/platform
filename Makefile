# variables that have to be setup for Docker
# DOCKER_REGISTRY
# DOCKER_USERNAME
# DOCKER_PASSWORD
# OPS_DOCKER_REGISTRY
# OPS_DOCKER_USERNAME
# OPS_DOCKER_PASSWORD

ROOT_DIRECTORY:=$(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

REPOSITORY_GOPATH:=$(word 1, $(subst :, ,$(GOPATH)))
REPOSITORY_PACKAGE:=github.com/tidepool-org/platform
REPOSITORY_NAME:=$(notdir $(REPOSITORY_PACKAGE))

BIN_DIRECTORY := ${ROOT_DIRECTORY}/_bin
PATH := ${PATH}:${BIN_DIRECTORY}

ifdef TRAVIS_TAG
	VERSION_BASE:=$(TRAVIS_TAG)
else
	VERSION_BASE:=$(shell git describe --abbrev=0 --tags 2> /dev/null || echo 'dblp.0.0.0')
endif
VERSION_BASE:=$(VERSION_BASE:dblp.%=%)
VERSION_SHORT_COMMIT:=$(shell git rev-parse --short HEAD)
VERSION_FULL_COMMIT:=$(shell git rev-parse HEAD)
VERSION_PACKAGE:=$(REPOSITORY_PACKAGE)/application

GO_LD_FLAGS:=-ldflags '-X $(VERSION_PACKAGE).VersionBase=$(VERSION_BASE) -X $(VERSION_PACKAGE).VersionShortCommit=$(VERSION_SHORT_COMMIT) -X $(VERSION_PACKAGE).VersionFullCommit=$(VERSION_FULL_COMMIT)'

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

test-until-failure: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 -r -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 -r -untilItFails $(TEST)

test-watch: ginkgo
	@echo "ginkgo watch -requireSuite -slowSpecThreshold=10 -r $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch -requireSuite -slowSpecThreshold=10 -r $(TEST)

ci-test: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 --compilers=2 -r -randomizeSuites -randomizeAllSpecs -failOnPending -race -timeout=8m --reportFile=junit-report/report.xml $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 --compilers=2 -r -randomizeSuites -randomizeAllSpecs -failOnPending -race -timeout=8m --reportFile=junit-report/report.xml $(TEST)

snyk-test:
	@echo "snyk test --dev --org=tidepool"
	@cd $(ROOT_DIRECTORY) && snyk test --dev --org=tidepool

snyk-monitor:
	@echo "snyk monitor --org=tidepool"
	@cd $(ROOT_DIRECTORY) && snyk monitor --org=tidepool

ci-snyk: snyk-test snyk-monitor

ci-test-until-failure: ginkgo
	@echo "ginkgo -requireSuite -slowSpecThreshold=10 -r -randomizeSuites -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress -keepGoing -untilItFails $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo -requireSuite -slowSpecThreshold=10 -r -randomizeSuites -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress -keepGoing -untilItFails $(TEST)

ci-test-watch: ginkgo
	@echo "ginkgo watch -requireSuite -slowSpecThreshold=10 -r -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress $(TEST)"
	@cd $(ROOT_DIRECTORY) && . ./env.test.sh && ginkgo watch -requireSuite -slowSpecThreshold=10 -r -randomizeAllSpecs -succinct -failOnPending -cover -trace -race -progress $(TEST)

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

ci-soups: clean-soup-doc generate-soups

generate-soups:
	@cd $(ROOT_DIRECTORY) && \
		$(MAKE) service-soup SERVICE_DIRECTORY=data SERVICE=platform TARGET=soup VERSION=${VERSION_BASE}

service-soup:
	@cd ${SERVICE_DIRECTORY} && \
		echo "# SOUPs List for ${SERVICE}@${VERSION}" > soup.md && \
		go list -f '## {{printf "%s \n\t* description: \n\t* version: %s\n\t* webSite: https://%s\n\t* sources:" .Path .Version .Path}}' -m all >> soup.md && \
		mkdir -p $(ROOT_DIRECTORY)/${TARGET}/${SERVICE} && \
		mv soup.md $(ROOT_DIRECTORY)/${TARGET}/${SERVICE}/${SERVICE}-${VERSION}-soup.md

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
	@echo "Login to Docker Default registry..."
	@echo $(DOCKER_PASSWORD) | docker login --username "$(DOCKER_USERNAME)" --password-stdin $(DOCKER_REGISTRY)
ifdef OPS_DOCKER_REPOSITORY
	@echo "Login to Docker Ops registry..."
	@echo $(OPS_DOCKER_PASSWORD) | docker login --username "$(OPS_DOCKER_USERNAME)" --password-stdin $(OPS_DOCKER_REGISTRY)
endif
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-build DOCKER_FILE="$${DOCKER_FILE}"; $(MAKE) docker-scan DOCKER_FILE="$${DOCKER_FILE}"; done
	@cd $(ROOT_DIRECTORY) && for DOCKER_FILE in $(shell ls -1 Dockerfile.*); do $(MAKE) docker-push DOCKER_FILE="$${DOCKER_FILE}"; done
endif

docker-build:
ifdef DOCKER
ifdef DOCKER_FILE
	docker build --tag $(DOCKER_REPOSITORY):development --target=development --file "$(DOCKER_FILE)" .
	docker build --tag $(DOCKER_REPOSITORY) --file "$(DOCKER_FILE)" .
ifdef TRAVIS_BRANCH
ifdef TRAVIS_COMMIT
ifdef TRAVIS_PULL_REQUEST_BRANCH
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
else
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	docker tag $(DOCKER_REPOSITORY) $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-latest
endif
endif
endif
ifdef TRAVIS_TAG
	docker tag "$(DOCKER_REPOSITORY)" "$(DOCKER_REPOSITORY):$(VERSION_BASE)"
ifdef OPS_DOCKER_REPOSITORY
	docker tag "$(DOCKER_REPOSITORY)" "$(OPS_DOCKER_REPOSITORY):$(VERSION_BASE)"
endif
endif
endif
endif

docker-scan:
	@echo "Security scan using Trivy container"
	@echo "Scan Image $(DOCKER_REPOSITORY)"
	@TRIVY_VERSION=$(shell curl --silent "https://api.github.com/repos/aquasecurity/trivy/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/') && \
		docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:$${TRIVY_VERSION} image --exit-code 0 --severity MEDIUM,LOW,UNKNOWN $(DOCKER_REPOSITORY) && \
		docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:$${TRIVY_VERSION} image --exit-code 1 --severity CRITICAL,HIGH $(DOCKER_REPOSITORY)

docker-push:
ifdef DOCKER
	@echo "DOCKER_REPOSITORY = $(DOCKER_REPOSITORY)"
	@echo "OPS_DOCKER_REPOSITORY = $(OPS_DOCKER_REPOSITORY)"
	@echo "TRAVIS_BRANCH = $(TRAVIS_BRANCH)"
	@echo "TRAVIS_PULL_REQUEST_BRANCH = $(TRAVIS_PULL_REQUEST_BRANCH)"
	@echo "TRAVIS_COMMIT = $(TRAVIS_COMMIT)"
	@echo "TRAVIS_TAG= $(TRAVIS_TAG)"
ifdef DOCKER_REPOSITORY
ifeq ($(TRAVIS_BRANCH),dblp)
ifeq ($(TRAVIS_PULL_REQUEST_BRANCH),)
	docker push $(DOCKER_REPOSITORY)
endif
endif
ifdef TRAVIS_BRANCH
ifdef TRAVIS_COMMIT
ifdef TRAVIS_PULL_REQUEST_BRANCH
	docker push $(DOCKER_REPOSITORY):PR-$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
else
	docker push $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-$(TRAVIS_COMMIT)
	docker push $(DOCKER_REPOSITORY):$(subst /,-,$(TRAVIS_BRANCH))-latest
endif
endif
endif
ifdef TRAVIS_TAG
	docker push "$(DOCKER_REPOSITORY):$(VERSION_BASE)"
endif
endif
ifdef OPS_DOCKER_REPOSITORY
ifdef TRAVIS_TAG
	@echo "Pushing to Ops..."
	docker push "$(OPS_DOCKER_REPOSITORY):$(VERSION_BASE)"
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

clean-soup-doc:
	@cd $(ROOT_DIRECTORY) && rm -rf soup

clean-all: clean

pre-commit: format imports vet lint

gopath-implode:
	cd $(REPOSITORY_GOPATH) && rm -rf bin pkg && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type f -delete && find src -not -path "src/$(REPOSITORY_PACKAGE)/*" -type d -empty -delete

.PHONY: default tmp bindir CompileDaemon esc ginkgo goimports golint buildable \
	format format-write imports vet vet-ignore lint lint-ignore pre-build build-list build ci-build \
	service-build service-start service-restart service-restart-all test test-watch ci-test c-test-watch \
	deploy deploy-services deploy-migrations deploy-tools ci-deploy bundle-deploy \
	docker docker-build docker-push ci-docker docker-scan \
	clean clean-bin clean-cover clean-debug clean-deploy clean-all pre-commit \
	gopath-implode