Configuration file can be specified as:
```sh
docurift -config config.yaml
```

Here is an example of DocuRift config file

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
```

The configuration file controls DocuRift's behavior:

### Proxy Section
- `port`: The port number that DocuRift's proxy server will listen on (e.g. 9876)
- `backend-url`: The URL of your backend service that DocuRift will forward requests to

### Analyzer Section  
- `port`: The port number for DocuRift's analyzer API endpoint (e.g. 9877)
- `max-examples`: Maximum number of example values to store for each field in the schema
- `redacted-fields`: A list of the fields to redact in the documentation. Their values will be shown as "REDACTED" (e.g. authorization header or api_keys that you don't want to expose in the doc) 
