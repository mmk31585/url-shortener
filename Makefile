BINARY=url-shortener

-include .env

.PHONY: all
all: fmt vet build test

.PHONY: build
build:
	go build -o $(BINARY) ./cmd/api

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	gofmt -s -w .
	@test -z "$(gofmt -s -d .)"

.PHONY: clean
clean:
	rm -f $(BINARY)
	go clean
