.PHONY: fmt test run

fmt:
	gofmt -w $(shell find . -name '*.go')

test:
	go test ./...

run:
	go run ./cmd/server
