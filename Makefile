SHELL = /bin/bash

default: lint test

test:
	go test -race -v ./...

benchmark:
	go test -bench=.

lint:
	golangci-lint run --verbose
