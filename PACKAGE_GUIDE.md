# Packaging Agents for the Parallels AI Registry

This guide provides instructions for agent developers on how to package and upload their AI agents to the Parallels AI Agents Registry.

## Prerequisites

- Python 3.8 or higher
- An agent implementation that follows the Parallels AI Agent interface
- Access to the Parallels AI Agents Registry server

## Getting Started

We provide a simple packaging tool that helps you prepare your agent for distribution.

### Basic Usage

```bash
# Navigate to your agent's source directory
cd /path/to/your/agent

# Use the packaging tool
/path/to/package-agent.py . --name "my-agent" --version "1.0.0" --description "My awesome AI agent" --author "Your Name" --author-email "your.email@example.com"
```

This will:

1. Create necessary packaging files in your source directory
2. Build a wheel package
3. Output the package to the `dist` directory

### Uploading to the Registry

To upload your package to the registry:

```bash
/path/to/package-agent.py . --name "my-agent" --version "1.0.0" --description "My awesome AI agent" --author "Your Name" --author-email "your.email@example.com" --upload --server "http://registry-server:8000"
```

## Agent Package Structure

A properly structured agent package should look like this:

```
my-agent/
├── my_agent/
│   ├── __init__.py
│   ├── agent.py
│   └── agent_metadata.json  <-- Important for discovery
├── setup.py
├── pyproject.toml
└── README.md
```

### agent_metadata.json

This file is critical for agent discovery. It should contain:

```json
{
  "name": "my-agent",
  "version": "1.0.0",
  "description": "My awesome AI agent",
  "author": "Your Name",
  "author_email": "your.email@example.com",
  "agents": [
    {
      "id": "my_chat_agent",
      "name": "My Chat Agent",
      "description": "A conversational agent",
      "type": "chat",
      "class_path": "my_agent.agent:MyChatAgent"
    }
  ]
}
```

The `agents` array can contain multiple agent entries, each with:

- `id`: A unique identifier for the agent
- `name`: A human-readable name
- `description`: A short description
- `type`: The agent type (e.g., "chat", "task")
- `class_path`: The Python path to the agent class

## Advanced Configuration

### Entry Points

You can also define your agents using Python entry points in `setup.py`:

```python
setup(
    # ...
    entry_points={
        'parallels.ai.agents': [
            'my_chat_agent = my_agent.agent:MyChatAgent',
        ],
    },
)
```

Or in `pyproject.toml`:

```toml
[project.entry-points."parallels.ai.agents"]
my_chat_agent = "my_agent.agent:MyChatAgent"
```

## Using the Client

The registry comes with a client that can be used to interact with the registry:

```bash
# List all available packages
./registry-client.sh list

# Download a package
./registry-client.sh download my-agent 1.0.0 my-agent-1.0.0-py3-none-any.whl

# Upload a package
./registry-client.sh upload my-agent 1.0.0 ./my-agent-1.0.0-py3-none-any.whl --description "My awesome AI agent"
```

## Agent Implementation

Your agent must follow the Parallels AI Agent interface. Here's a minimal example:

```python
class MyChatAgent:
    """A simple chat agent."""
    
    def __init__(self, config=None):
        self.config = config or {}
    
    async def process_message(self, message, context=None):
        """Process a message from the user."""
        return {
            "response": f"You said: {message}",
            "type": "text"
        }
    
    @classmethod
    def get_manifest(cls):
        """Return the agent manifest."""
        return {
            "name": "My Chat Agent",
            "description": "A simple chat agent",
            "version": "1.0.0",
            "author": "Your Name",
            "type": "chat"
        }
```

## Testing Your Agent

Before uploading your agent, you should test it locally:

```bash
# Build the package
python -m build

# Install the package in development mode
pip install -e .

# Test importing and using your agent
python -c "from my_agent.agent import MyChatAgent; agent = MyChatAgent(); print(agent.get_manifest())"
```

## Troubleshooting

- **Package not found after upload**: Ensure your package name and version match what you uploaded.
- **Agent not discovered**: Check your `agent_metadata.json` and make sure it's correctly formatted.
- **Import errors**: Ensure all dependencies are properly listed in your `setup.py` or `pyproject.toml`.

## Need Help?

Contact the Parallels AI team for assistance with packaging or uploading your agent.
