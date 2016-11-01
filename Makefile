#! /usr/bin/make
#
# Makefile for goa
#
# Targets:
# - "depend" retrieves the Go packages needed to run the linter and tests
# - "lint" runs the linter and checks the code format using goimports
# - "test" runs the tests
#
# Meta targets:
# - "all" is the default target, it runs all the targets in the order above.
#
DIRS=$(shell go list -f {{.Dir}} ./...)

# Only list test and build dependencies
# Standard dependencies are installed via go get
DEPEND=\
	github.com/golang/lint/golint \
	github.com/on99/gocyclo \
	golang.org/x/tools/cmd/cover \
	golang.org/x/tools/cmd/goimports

.PHONY: goagen

all: depend lint cyclo goagen test

depend:
	@go get -v ./...
	@go get -v $(DEPEND)

lint:
	@for d in $(DIRS) ; do \
		if [ "`goimports -l $$d/*.go | tee /dev/stderr`" ]; then \
			echo "^ - Repo contains improperly formatted go files" && echo && exit 1; \
		fi \
	done
	@if [ "`golint ./... | grep -vf .golint_exclude | tee /dev/stderr`" ]; then \
		echo "^ - Lint errors!" && echo && exit 1; \
	fi

cyclo:
	@if [ "`gocyclo -over 20 . | grep -v _integration_tests | tee /dev/stderr`" ]; then \
		echo "^ - Cyclomatic complexity exceeds 20, refactor the code!" && echo && exit 1; \
	fi

test:
	go test ./...
	go test ./_integration_tests

goagen:
	@cd goagen && \
	go install
