# DocuRift

DocuRift is a powerful tool that automatically generates and maintains REST API documentation by observing real API traffic. It acts as a proxy between your clients and API server, capturing and analyzing requests and responses to create comprehensive documentation.

This is particular useful for existing REST API service with live traffic but lack of documentation. Setup DocuRift as a proxy to run in front of your existing service and wait for a period of time, and come back to get your Open API spec and Postman Collection! 

This is safe to run in production environment as it does not send any data out, and masks sensitive user information by default. Analyzer and document generator runs asynchronously without additional performance penalty to your existing service!

```mermaid
sequenceDiagram
    participant Client
    participant DocuRift
    participant Backend
    participant Analyzer
    
    Client->>DocuRift: HTTP Request
    DocuRift->>Backend: Forward Request
    Backend->>DocuRift: HTTP Response
    DocuRift->>Analyzer: Process Traffic
```

## Features

- 🔄 **Real-time Documentation**: Automatically generates documentation from actual API usage
- 📝 **OpenAPI/Swagger Support**: Generates OpenAPI 3.0 specifications
- 📦 **Postman Collection**: Creates Postman collections for easy API testing
- 🔍 **Request/Response Examples**: Captures real examples of API usage
- 🛡️ **Security**: Handles sensitive data appropriately
- 📊 **Interactive UI**: Integrated Swagger UI for documentation browsing

## Installation

### Installing from Source Code

Clone, build and install into $GOBIN ($GOPATH/bin or $HOME/go/bin):

```bash
go install github.com/tienanr/docurift/cmd/docurift@latest
```

### Alternatively you can install via Docker

Pull and run the latest version:

```bash
docker pull ghcr.io/tienanr/docurift:latest
docker run -p 9876:9876 -p 9877:9877 -v $(pwd)/config.yaml:/etc/docurift/config.yaml ghcr.io/tienanr/docurift:latest
```

See [docs/configuration.md](docs/configuration.md) for details about config file.

## Quick Start

1. Create a configuration file (e.g., `config.yaml`):
```yaml
proxy:
    port: 9876
    backend-url: http://localhost:8080

analyzer:
    port: 9877
    max-examples: 20
    redacted-fields:
        - password
    storage:
        path: .
        frequency: 10
```
if you run from docker, use this instead:
```yaml
proxy:
    port: 9876
    backend-url: http://host.docker.internal:8080

analyzer:
    port: 9877
    max-examples: 20
    redacted-fields:
        - password
    storage:
        path: /etc/docurift/
        frequency: 10
```

2. Start DocuRift with your configuration:
```bash
docurift -config config.yaml
```
or
```bash
docker run -p 9876:9876 -p 9877:9877 -v $(pwd)/config-docker.yaml:/etc/docurift/config.yaml ghcr.io/tienanr/docurift:latest
```

3. Clone the examples repository to test DocuRift:
```bash
git clone https://github.com/tienanr/docurift.git 
cd docurift/examples/shop
```

4. Start your API server (example using the included shop API) on port 8080 and send test traffic:
```bash
lsof -ti :8080 | xargs kill
go run shop.go &
sleep 3
go test -count=1 .
```

4. Access your automatically generated documentation at http://localhost:9877/ (Swagger UI)

5. Get open API spec: http://localhost:9877/openapi.json and Postman Collection: http://localhost:9877/postman.json

6. Check the local folder for `analyzer.json` file, which is persisted every 10 seconds while analyzer is running, the next analyzer run wil pick up any stored state. For docker usages, add proper mount to expose `analyzer.json` to load persisted state.

## Configuration

DocuRift uses a YAML configuration file to control its behavior. Here's the complete configuration reference:

### Proxy Section
- `port`: The port number that DocuRift's proxy server will listen on (e.g. 9876), point your clients request to this port instead of the real backend.
- `backend-url`: The URL of your backend service that DocuRift will forward requests to.

### Analyzer Section  
- `port`: The port number for DocuRift's analyzer API endpoint (e.g. 9877)
- `max-examples`: Maximum number of example values to store for each field in the schema
- `redacted-fields`: A list of fields to redact in the documentation. Their values will be shown as "REDACTED" (e.g. authorization headers, API keys, passwords). This applies globally to HTTP headers, URL parameters and JSON fields.
- `storage.path`: The directory path where DocuRift will store its analyzer state file (analyzer.json). Defaults to current directory if not specified.
- `storage.frequency`: How often (in seconds) DocuRift should save its state to disk. Defaults to 10 seconds if not specified.

Example configuration:
```yaml
proxy:
    port: 9876
    backend-url: http://localhost:8080

analyzer:
    port: 9877
    max-examples: 10
    redacted-fields:
        - Authorization
        - api_key
        - password
    storage:
        path: .
        frequency: 10
```

## Examples

Check out the `examples` directory for sample implementations:
- `examples/shop`: A complete e-commerce API with various endpoints

## Contributing

If you find any bugs or need more features please let me know!
Contributions are also welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 