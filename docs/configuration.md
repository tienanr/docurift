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
    storage:
        path: .
        frequency: 10
```

The configuration file controls DocuRift's behavior:

### Proxy Section
- `port`: The port number that DocuRift's proxy server will listen on (e.g. 9876)
- `backend-url`: The URL of your backend service that DocuRift will forward requests to

### Analyzer Section  
- `port`: The port number for DocuRift's analyzer API endpoint (e.g. 9877)
- `max-examples`: Maximum number of example values to store for each field in the schema
- `redacted-fields`: A list of the fields to redact in the documentation. Their values will be shown as "REDACTED" (e.g. authorization header or api_keys that you don't want to expose in the doc) 

- `storage.path`: The directory path where DocuRift will store its analyzer state file (analyzer.json). Defaults to current directory if not specified.
- `storage.frequency`: How often (in seconds) DocuRift should save its state to disk. Defaults to 10 seconds if not specified.

