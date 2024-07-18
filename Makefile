SHELL = /bin/bash
export PATH := $(shell yarn global bin):$(PATH)

default: lint test

test:
	go test -race -v ./...

benchmark:
	go test -bench=.

lint:
	golangci-lint run --verbose

release:
	npx semantic-release@v19.0.2
