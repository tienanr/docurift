FROM golang:1.24-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache curl jq

# Copy only the necessary files
COPY test/backend/backend.go test/backend/
COPY test/backend/backend_test.go test/backend/
COPY test/run_tests.sh /usr/local/bin/run_tests.sh
COPY go.mod go.sum ./

RUN chmod +x /usr/local/bin/run_tests.sh
RUN go mod download

CMD ["run_tests.sh"] 