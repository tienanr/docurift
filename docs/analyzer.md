Analyzer module design

For each request received by the proxy server, analyzer will be invoked with full request and response.

Analyzer maintains an internal data structure to track each (http method, url) combination, for each combination, it would maintain following structure:

* HTTP method
* URL
* Request header schema store
* Request payload JSON schema store.
* Response status data store
* For each response status, maintain response header schema store and response payload JSON schema store.

JSON schema store is a map keyed by JSON schema path.

example:
{
    "user": {
        "friends": [
            {
                "name": "John",
                "age": 25
            }
        ]
    }
}

JSON schema path of name field should be "user.friends[].name", all nested objects or arrays need to be expanded until we get primitives.

for each discovered path, store a list of example values we have seen under this path and a boolean value optional, which is true if all request/response contain this field, otherwise false.

Header store is similar to schema store, where headers keys are the keys, values are stored as examples, an optional flag to track if it always exists.

Expose an analyzer endpoint on port 8082, which provide a JSON view of the data structure.