BINARY_NAME=openspec
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/santif/openspec-go/internal/cli.version=$(VERSION)"

.PHONY: build test lint vet clean install

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/openspec

test:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

install: build
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

all: vet test build
