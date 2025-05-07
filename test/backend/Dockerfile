FROM golang:1.24-alpine

WORKDIR /app

# Install curl for health checks
RUN apk add --no-cache curl

# Copy only the necessary files
COPY test/backend/backend.go backend.go
COPY test/backend/backend_test.go backend_test.go
COPY go.mod go.sum ./

RUN go mod download

EXPOSE 8081

CMD ["go", "run", "backend.go"] 