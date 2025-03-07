#!/bin/bash

# Parallels AI Agents Registry Client wrapper script
# This script makes it easier to run the client

# Get the directory of this script
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

# Default server URL
SERVER_URL="http://localhost:8000"

# Check if a custom server URL is specified in .env
if [ -f "$SCRIPT_DIR/.env" ]; then
  source "$SCRIPT_DIR/.env"
  if [ -n "$REGISTRY_SERVER_URL" ]; then
    SERVER_URL="$REGISTRY_SERVER_URL"
  fi
fi

# Check if a server URL is passed as an environment variable
if [ -n "$REGISTRY_SERVER_URL" ]; then
  SERVER_URL="$REGISTRY_SERVER_URL"
fi

# Run the client
python "$SCRIPT_DIR/client.py" --server "$SERVER_URL" "$@"
