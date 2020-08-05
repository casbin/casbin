SHELL = /bin/bash
export PATH := $(shell yarn global bin):$(PATH)

default: lint test

test:
	go test -race -v .

benchmark:
	go test -bench=.

lint:
	golangci-lint run --verbose

release:
	yarn global add semantic-release
	semantic-release

