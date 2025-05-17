#!/usr/bin/env python3

import json
import requests
from openai import OpenAI
from typing import Dict, Any, List
import os
import sys
from rich.console import Console
from rich.markdown import Markdown
from rich.panel import Panel
from rich.prompt import Prompt
from rich.table import Table

class DocuRiftClient:
    def __init__(self, docurift_url: str = "http://localhost:9877"):
        """Initialize the DocuRift client."""
        self.docurift_url = docurift_url.rstrip('/')
        self.tools = self._load_tools()
        self.console = Console()

    def _load_tools(self) -> List[Dict[str, Any]]:
        """Load tools specification from DocuRift."""
        try:
            response = requests.get(f"{self.docurift_url}/tools/openai.json")
            response.raise_for_status()
            return response.json()["tools"]
        except requests.exceptions.RequestException as e:
            self.console.print(f"[red]Error loading tools from DocuRift: {e}[/red]")
            sys.exit(1)

    def execute_tool(self, tool_name: str, args: Dict[str, Any]) -> Dict[str, Any]:
        """Execute a tool by making a request to the actual API endpoint."""
        # Find the tool specification
        tool_spec = next((t for t in self.tools if t["function"]["name"] == tool_name), None)
        if not tool_spec:
            raise ValueError(f"Tool {tool_name} not found")

        # Extract method and path from tool name
        # Example: post_api_users -> POST /api/users
        method = tool_spec["function"]["name"].split("_")[0].upper()
        path = "/" + "/".join(tool_spec["function"]["name"].split("_")[1:])

        # Make the actual API request
        try:
            response = requests.request(
                method=method,
                url=f"http://localhost:8080{path}",
                json=args
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            self.console.print(f"[red]Error executing tool {tool_name}: {e}[/red]")
            return {"error": str(e)}

    def list_available_tools(self):
        """Display available tools in a table."""
        table = Table(title="Available API Tools")
        table.add_column("Name", style="cyan")
        table.add_column("Description", style="green")
        table.add_column("Parameters", style="yellow")

        for tool in self.tools:
            func = tool["function"]
            params = func["parameters"]
            param_desc = []
            for name, prop in params.get("properties", {}).items():
                required = "required" if name in params.get("required", []) else "optional"
                param_desc.append(f"{name} ({required})")
            
            table.add_row(
                func["name"],
                func["description"],
                "\n".join(param_desc)
            )

        self.console.print(table)

class ChatInterface:
    def __init__(self):
        self.console = Console()
        self.docurift = DocuRiftClient()
        self.client = OpenAI()
        self.messages = [
            {"role": "system", "content": "You are a helpful assistant that can interact with the API using the provided tools. Be concise and clear in your responses."}
        ]

    def display_welcome(self):
        """Display welcome message and available tools."""
        self.console.print(Panel.fit(
            "[bold blue]DocuRift Chat Interface[/bold blue]\n"
            "Type 'help' to see available commands\n"
            "Type 'exit' to quit\n"
            "Type 'tools' to see available API tools",
            title="Welcome"
        ))

    def display_help(self):
        """Display help message."""
        help_text = """
        Available commands:
        - help: Show this help message
        - exit: Exit the chat
        - tools: Show available API tools
        - clear: Clear the conversation history
        """
        self.console.print(Markdown(help_text))

    def clear_conversation(self):
        """Clear the conversation history."""
        self.messages = [
            {"role": "system", "content": "You are a helpful assistant that can interact with the API using the provided tools. Be concise and clear in your responses."}
        ]
        self.console.print("[green]Conversation history cleared[/green]")

    def process_user_input(self, user_input: str) -> bool:
        """Process user input and return True if should continue, False if should exit."""
        if user_input.lower() == 'exit':
            return False
        elif user_input.lower() == 'help':
            self.display_help()
            return True
        elif user_input.lower() == 'tools':
            self.docurift.list_available_tools()
            return True
        elif user_input.lower() == 'clear':
            self.clear_conversation()
            return True

        # Add user message to conversation
        self.messages.append({"role": "user", "content": user_input})

        try:
            # Get response from OpenAI
            response = self.client.chat.completions.create(
                model="gpt-3.5-turbo",
                messages=self.messages,
                tools=self.docurift.tools
            )

            # Process the response
            message = response.choices[0].message
            self.messages.append(message)

            # If the model wants to call a function
            if message.tool_calls:
                for tool_call in message.tool_calls:
                    # Show tool call
                    self.console.print(f"\n[cyan]Calling tool: {tool_call.function.name}[/cyan]")
                    self.console.print(f"[yellow]Arguments: {tool_call.function.arguments}[/yellow]")

                    # Execute the tool
                    result = self.docurift.execute_tool(
                        tool_call.function.name,
                        json.loads(tool_call.function.arguments)
                    )

                    # Add the result to the conversation
                    self.messages.append({
                        "role": "tool",
                        "tool_call_id": tool_call.id,
                        "name": tool_call.function.name,
                        "content": json.dumps(result)
                    })

                    # Show tool result
                    self.console.print(f"[green]Result: {json.dumps(result, indent=2)}[/green]")

                    # Get a new response from the model
                    response = self.client.chat.completions.create(
                        model="gpt-3.5-turbo",
                        messages=self.messages,
                        tools=self.docurift.tools
                    )
                    message = response.choices[0].message
                    self.messages.append(message)

            # Display the response
            self.console.print("\n[bold blue]Assistant:[/bold blue]")
            self.console.print(Markdown(message.content))

        except Exception as e:
            self.console.print(f"[red]Error: {str(e)}[/red]")

        return True

    def run(self):
        """Run the chat interface."""
        self.display_welcome()

        while True:
            try:
                user_input = Prompt.ask("\n[bold green]You[/bold green]")
                if not self.process_user_input(user_input):
                    break
            except KeyboardInterrupt:
                self.console.print("\n[yellow]Use 'exit' to quit[/yellow]")
            except EOFError:
                break

        self.console.print("\n[blue]Goodbye![/blue]")

def main():
    # Check for OpenAI API key
    if not os.getenv("OPENAI_API_KEY"):
        print("Please set your OpenAI API key:")
        print("export OPENAI_API_KEY='your-api-key-here'")
        sys.exit(1)

    # Run the chat interface
    chat = ChatInterface()
    chat.run()

if __name__ == "__main__":
    main() 