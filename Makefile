.PHONY: build build-dev clean

# Version information
VERSION ?= v0.1.4
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +'%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS = -X github.com/tienanr/docurift/cmd/docurift.version=$(VERSION) \
          -X github.com/tienanr/docurift/cmd/docurift.commit=$(COMMIT) \
          -X github.com/tienanr/docurift/cmd/docurift.date=$(DATE)

# Build the application with version information
build:
	go build -ldflags "$(LDFLAGS)" -o docurift ./cmd/docurift

# Build for development (without version information)
build-dev:
	go build -o docurift ./cmd/docurift

# Clean build artifacts
clean:
	rm -f docurift 