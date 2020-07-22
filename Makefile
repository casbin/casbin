default: lint test

test:
	go test -race -v .

benchmark:
	go test -bench=.

lint:
	golangci-lint run --verbose

release:
	export PATH="$(yarn global bin):$PATH"
	yarn global add semantic-release
	semantic-release

