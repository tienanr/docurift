.PHONY: build install test clean version

# Build variables
BINARY_NAME=docurift
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(shell git rev-parse --short HEAD) -X main.date=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/docurift

install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/docurift

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

version:
	@echo "Version: $(VERSION)" 