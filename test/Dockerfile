FROM golang:1.24-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache curl jq diffutils

# Copy test files

COPY test/run_tests.sh /usr/local/bin/run_tests.sh
COPY test/expected_openapi.json /app/expected_openapi.json
COPY go.mod go.sum ./
COPY examples/shop/ ./shop

RUN chmod +x /usr/local/bin/run_tests.sh
RUN go mod download

CMD ["run_tests.sh"] 