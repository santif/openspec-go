BINARY_NAME=openspec
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/santif/openspec-go/internal/cli.version=$(VERSION)"

.PHONY: build test coverage lint vet clean install

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/openspec

test:
	go test -race ./...

coverage:
	go test -coverprofile=/tmp/coverage.out ./...
	go tool cover -func=/tmp/coverage.out | tail -1
	go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
	@echo "HTML report: /tmp/coverage.html"

vet:
	go vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

install: build
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

all: vet test build
