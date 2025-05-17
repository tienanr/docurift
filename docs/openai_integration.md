# Using DocuRift with OpenAI

DocuRift can automatically generate OpenAI function calling specifications from your API traffic, making it easy to integrate your API with OpenAI's models. This guide explains how to use DocuRift with OpenAI.

## Overview

DocuRift acts as a proxy between your clients and API server, analyzing the traffic to generate OpenAI function calling specifications. These specifications can be used to:

1. Enable AI models to understand your API
2. Allow AI to make API calls on behalf of users
3. Provide accurate API documentation to the AI

## Setup

### 1. Start DocuRift

First, start DocuRift with your configuration:

```bash
docurift -config config.yaml
```

Make sure your `config.yaml` points to your backend service:

```yaml
proxy:
    port: 9876
    backend-url: http://localhost:8080

analyzer:
    port: 9877
    max-examples: 10
```

### 2. Send Traffic Through DocuRift

Send your API traffic through DocuRift's proxy port (9876) instead of directly to your backend. This allows DocuRift to analyze the traffic and generate the tools specification.

### 3. Get the Tools Specification

DocuRift provides the tools specification in OpenAI's format at:

```
http://localhost:9877/tools/openai.json
```

## Using with OpenAI

### Basic Integration

Here's a simple example of using DocuRift's tools with OpenAI:

```python
from openai import OpenAI
import requests

# Initialize the OpenAI client
client = OpenAI()

# Load tools from DocuRift
response = requests.get("http://localhost:9877/tools/openai.json")
tools = response.json()["tools"]

# Use in your chat completion
response = client.chat.completions.create(
    model="gpt-4-turbo-preview",
    messages=[{"role": "user", "content": "Create a new user"}],
    tools=tools
)
```

### Interactive Chat Interface

DocuRift provides an interactive chat interface for testing the integration:

1. Install requirements:
```bash
pip install -r examples/requirements.txt
```

2. Set your OpenAI API key:
```bash
export OPENAI_API_KEY='your-api-key-here'
```

3. Run the chat interface:
```bash
python examples/openai_client.py
```

The interface provides several commands:
- `help`: Show available commands
- `tools`: Display available API tools
- `clear`: Clear conversation history
- `exit`: Quit the chat

#### Example Interaction

Here's an example conversation demonstrating the tools interface:

```
Welcome to DocuRift Chat Interface
Type 'help' to see available commands
Type 'exit' to quit
Type 'tools' to see available API tools

You: tools
Available API Tools:
┌─────────────────┬─────────────────────────────────────┬────────────────────┐
│ Name            │ Description                         │ Parameters         │
├─────────────────┼─────────────────────────────────────┼────────────────────┤
│ post_api_users  │ Creates a new user                  │ name (required)    │
│                 │                                     │ age (required)     │
│                 │                                     │ email (optional)   │
└─────────────────┴─────────────────────────────────────┴────────────────────┘

You: Create a new user named John who is 30 years old
Assistant: I'll help you create a new user.

Calling tool: post_api_users
Arguments: {"name": "John", "age": 30}
Result: {"id": 1, "name": "John", "age": 30}

User John has been created successfully with ID 1.

You: What users do we have?
Assistant: I'll retrieve the list of users.

Calling tool: get_api_users
Arguments: {}
Result: [{"id": 1, "name": "John", "age": 30}]

We have 1 user in the system:
- John (ID: 1, Age: 30)

You: Update John's age to 31
Assistant: I'll update John's age.

Calling tool: put_api_users_1
Arguments: {"age": 31}
Result: {"id": 1, "name": "John", "age": 31}

John's age has been updated to 31.

You: clear
Conversation history cleared

You: exit
Goodbye!
```

This example shows:
1. Viewing available tools with the `tools` command
2. Creating a new user with natural language
3. Retrieving user information
4. Updating user data
5. Clearing conversation history
6. Exiting the interface

Each tool call shows:
- The tool being called
- The arguments being passed
- The result received
- A human-readable response

## Tools Specification Format

DocuRift generates tools in OpenAI's function calling format:

```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "post_api_users",
        "description": "Creates a new user. Use this function to add new data to the system.",
        "parameters": {
          "type": "object",
          "properties": {
            "name": {
              "type": "string",
              "description": "User's name (text)"
            },
            "age": {
              "type": "number",
              "minimum": 0,
              "maximum": 150,
              "description": "User's age (number). Must be a number between 0 and 150."
            }
          },
          "required": ["name", "age"]
        }
      }
    }
  ]
}
```

The specification includes:
- Function names derived from HTTP methods and paths
- Detailed parameter descriptions
- Type information and validation rules
- Required/optional parameter flags

## Best Practices

1. **Traffic Analysis**:
   - Send diverse traffic through DocuRift to capture all API endpoints
   - Include examples of different parameter combinations
   - Test error cases to ensure proper handling

2. **Tool Names**:
   - Tool names are automatically generated as `method_path`
   - Example: `POST /api/users` becomes `post_api_users`
   - Use consistent naming in your API for better tool names

3. **Parameter Descriptions**:
   - DocuRift automatically enhances parameter descriptions
   - Adds type information and validation rules
   - Detects common field types (email, URL, date, etc.)

4. **Error Handling**:
   - Include error responses in your traffic
   - DocuRift will capture error patterns
   - AI will learn to handle errors appropriately

## Example Workflow

1. Start your API and DocuRift:
```bash
# Start your API
./your-api-server

# Start DocuRift
docurift -config config.yaml
```

2. Send test traffic:
```bash
# Send requests through DocuRift
curl -X POST http://localhost:9876/api/users -d '{"name": "John", "age": 30}'
```

3. Use with OpenAI:
```python
from openai import OpenAI
import requests

# Initialize the OpenAI client
client = OpenAI()

# Get tools from DocuRift
tools = requests.get("http://localhost:9877/tools/openai.json").json()["tools"]

# Use in your application
response = client.chat.completions.create(
    model="gpt-4-turbo-preview",
    messages=[{"role": "user", "content": "Create a new user"}],
    tools=tools
)
```

## Troubleshooting

1. **Tools Not Available**:
   - Ensure DocuRift is running
   - Check that traffic is being sent through DocuRift
   - Verify the tools endpoint is accessible

2. **Incorrect Tool Names**:
   - Check your API endpoint naming
   - Ensure consistent HTTP methods
   - Review the generated tools specification

3. **Missing Parameters**:
   - Send more diverse traffic through DocuRift
   - Include examples with different parameters
   - Check the analyzer's max-examples setting

4. **API Connection Issues**:
   - Verify backend-url in config.yaml
   - Check network connectivity
   - Ensure proper port forwarding

## Security Considerations

1. **API Keys**:
   - Never expose your OpenAI API key
   - Use environment variables
   - Implement proper access controls

2. **Sensitive Data**:
   - Configure redacted fields in DocuRift
   - Review the tools specification for sensitive data
   - Implement proper validation

3. **Rate Limiting**:
   - Monitor API usage
   - Implement rate limiting
   - Set appropriate quotas

## Additional Resources

- [OpenAI Function Calling Documentation](https://platform.openai.com/docs/guides/function-calling)
- [DocuRift Configuration Guide](configuration.md)
- [Example API Integration](examples/openai_client.py) 