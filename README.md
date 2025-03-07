# Parallels AI Agents Registry

A simple, self-hosted package registry server for Parallels AI Agents. This server allows you to store and distribute private Python packages for AI Agents without exposing source code.

## Features

- Upload and download Python packages (.whl, .tar.gz)
- Package versioning
- Metadata storage
- RESTful API for integration
- Docker support for easy deployment

## Setup and Run

### Using Docker

The simplest way to run the registry server is with Docker:

```bash
# Clone the repository
git clone https://github.com/parallels/ai-agents-registry.git
cd ai-agents-registry

# Run with Docker Compose
docker-compose up -d
```

This will start the server on port 8000 (<http://localhost:8000>).

### Manual Setup

If you prefer to run directly without Docker:

```bash
# Clone the repository
git clone https://github.com/parallels/ai-agents-registry.git
cd ai-agents-registry

# Create and activate virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Run the server
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

## Configuration

The server can be configured via environment variables or a `.env` file:

```
# Server configuration
PORT=8000
HOST=0.0.0.0

# Storage paths
PACKAGES_DIR=./data/packages
TEMP_DIR=./data/temp
METADATA_DIR=./data/metadata

# Security settings
API_KEY_REQUIRED=false
API_KEY=
ADMIN_USERNAME=admin
ADMIN_PASSWORD_HASH=

# Other settings
MAX_UPLOAD_SIZE=104857600  # 100MB in bytes
```

## API Usage

### Upload a Package

```bash
curl -X POST \
  "http://localhost:8000/packages/my-agent/versions/1.0.0/upload" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@./my-agent-1.0.0-py3-none-any.whl" \
  -F "description=My awesome AI agent" \
  -F "author=John Doe" \
  -F "author_email=john@example.com"
```

### List Available Packages

```bash
curl "http://localhost:8000/packages"
```

### Download a Package

```bash
curl -O "http://localhost:8000/download/my-agent/1.0.0/my-agent-1.0.0-py3-none-any.whl"
```

### Delete a Package

```bash
curl -X DELETE "http://localhost:8000/packages/my-agent/versions/1.0.0/my-agent-1.0.0-py3-none-any.whl"
```

## For Agent Developers

To prepare your agent for distribution via this registry:

1. Structure your project as a standard Python package
2. Use setuptools or poetry to build your package
3. Include metadata about your agent in your package
4. Upload your package to the registry

Example agent package structure:

```
my-agent/
  my_agent/
    __init__.py
    agent.py
  pyproject.toml
  setup.py
  setup.cfg
  README.md
```

## Client Integration

To integrate with the Parallels AI Agents client, your agent package should:

1. Include necessary metadata in a standardized location
2. Follow the Parallels AI Agents interface
3. Provide documentation on usage and parameters

## License

[MIT License](LICENSE)
