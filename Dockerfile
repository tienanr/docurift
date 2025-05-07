FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy only the necessary files
COPY cmd/docurift cmd/docurift
COPY internal internal
COPY go.mod go.sum ./

RUN go mod download
RUN go build -o docurift ./cmd/docurift

FROM alpine:latest

WORKDIR /app

# Install curl for health checks
RUN apk add --no-cache curl

COPY --from=builder /app/docurift /usr/local/bin/
COPY config.yaml /app/

# Verify config file exists and is readable
RUN cat /app/config.yaml

EXPOSE 8080 8082

CMD ["docurift"] 