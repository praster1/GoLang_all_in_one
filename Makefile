.PHONY: fmt build test vet docs race check

fmt:
	go fmt ./...

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

docs:
	go run ./tools/checkdocs

race:
	go test -race ./...

check: build vet test docs
