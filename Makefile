default: lint test

test:
	go test -race -v .

benchmark:
	go test -bench=.

benchmark-regressions-check:
	go test -bench=. -count=5

lint:
	golangci-lint run --verbose

release:
	export PATH="$(yarn global bin):$PATH"
	yarn global add semantic-release
	semantic-release

