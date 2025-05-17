# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o docurift ./cmd/docurift

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/docurift /usr/local/bin/docurift

# Copy example config
COPY config.yaml /etc/docurift/config.yaml

# Create a non-root user
RUN adduser -D -g '' docurift

# Use non-root user
USER docurift

# Set the entrypoint
ENTRYPOINT ["docurift"]

# Default command
CMD ["-config", "/etc/docurift/config.yaml"] 