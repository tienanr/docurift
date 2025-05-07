# DocuRift

DocuRift is an intelligent API documentation generator that automatically analyzes your API traffic and generates comprehensive documentation in multiple formats. It works as a proxy server that captures API requests and responses, then generates OpenAPI and Postman collection documentation.

## Features

- 🔄 Automatic API traffic analysis
- 📝 Generates OpenAPI (Swagger) documentation
- 📦 Creates Postman collections
- 🔍 Tracks request/response examples
- 📊 Supports query parameters, headers, and request bodies
- ⚡ Real-time documentation updates
- 🎯 Configurable example limits

## Installation

### Using Go Install (Recommended)

```bash
go install github.com/tienanr/docurift/cmd/docurift@latest
```

This will install the latest version of DocuRift to your `$GOPATH/bin` directory.

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/tienanr/docurift.git
cd docurift
```

2. Build the project:
```bash
make build
```

Or install directly:
```bash
make install
```

## Configuration

The application uses a YAML configuration file (`src/config.yaml`) with the following settings:

```yaml
# Proxy server configuration
proxy:
  port: 8080
  backend_url: "http://localhost:8081"

# Analyzer server configuration
analyzer:
  port: 8082
  max_examples: 10  # Maximum number of example values to keep per field
```

You can customize:
- Proxy server port
- Backend API URL
- Analyzer server port
- Maximum number of examples to keep per field

## Usage

1. Start the application:
```bash
./docurift
```

Or specify a custom config file:
```bash
./docurift -config /path/to/config.yaml
```

2. Configure your application to use the proxy server (default: `http://localhost:8080`)

3. Access the generated documentation:
- OpenAPI documentation: `http://localhost:8082/openapi.json`
- Postman collection: `http://localhost:8082/postman.json`

## How It Works

1. The proxy server (`:8080`) intercepts all API traffic
2. The analyzer server (`:8082`) processes the traffic and generates documentation
3. Documentation is automatically updated as new API calls are made
4. Access the documentation through the analyzer server endpoints

## Documentation Formats

### OpenAPI (Swagger)
- Comprehensive API specification
- Includes endpoints, methods, parameters, and examples
- Compatible with Swagger UI and other OpenAPI tools
- View interactive documentation at `http://localhost:8082/swagger-ui/`
- Test API endpoints directly from the Swagger UI interface

### Postman Collection
- Ready-to-use Postman collection
- Includes example requests and responses
- Can be imported directly into Postman

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 