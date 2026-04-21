.PHONY: build install clean test lint

BINARY=smartxctl
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%SZ')
LDFLAGS=-ldflags "-s -w -X smartxctl/internal/cli.version=$(VERSION) -X smartxctl/internal/cli.buildTime=$(BUILD_TIME)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd

install: build
	install -m 755 $(BINARY) /usr/local/bin/

clean:
	rm -f $(BINARY)

test:
	go test -v ./...

lint:
	go vet ./...

run:
	./$(BINARY) $(ARGS)

help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  install - Install to /usr/local/bin/"
	@echo "  clean   - Remove binary"
	@echo "  test    - Run tests"
	@echo "  lint    - Run linter"
	@echo "  run     - Run with arguments"
