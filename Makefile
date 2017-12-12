BINARY = grawler
PACKAGE = rahulrumalla/grawler
GOLANG_VERSION = 1.8.0

VERSION = 0.0.1
BUILD_DATE = `date +%FT%T%z`
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`

PACKAGES = $(shell go list ./... | grep -vE "/vendor/" | grep "rahulrumalla/")

DEV_DEPS = github.com/tools/godep \
		   github.com/golang/lint/golint \
		   github.com/spf13/cobra/cobra

.DEFAULT_GOAL := help

## Build the project
build:
	go build ${LDFLAGS} ${PACKAGE}

## Clean source directory
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	rm -rf ./.tmp
	go clean
	
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install: ## Installs grawler
	@go install $(PACKAGES)
	
test: ## test entire project
	@go test -v $(PACKAGES)

install-dev-deps:
	go get -v $(DEV_DEPS)

dev-setup: install-dev-deps check ## Initial set up for dev environment