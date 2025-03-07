#!/usr/bin/env python3
"""
Parallels AI Agent Packager

A utility to help developers package their AI agents for distribution
through the Parallels AI Agents Registry.
"""

import argparse
import os
import sys
import json
import shutil
import subprocess
from pathlib import Path
from typing import Optional, Dict, Any, List


def parse_args():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description="Parallels AI Agent Packager")

    parser.add_argument("source_dir", help="Source directory containing the agent code")
    parser.add_argument(
        "--name", help="Agent package name (defaults to directory name)"
    )
    parser.add_argument("--version", default="0.1.0", help="Agent version")
    parser.add_argument("--author", help="Agent author")
    parser.add_argument("--author-email", help="Agent author email")
    parser.add_argument("--description", help="Agent description")
    parser.add_argument(
        "--output-dir", default="./dist", help="Output directory for the package"
    )
    parser.add_argument(
        "--upload", action="store_true", help="Upload the package to the registry"
    )
    parser.add_argument(
        "--server", default="http://localhost:8000", help="Registry server URL"
    )
    parser.add_argument("--api-key", help="API key for authentication")

    return parser.parse_args()


def validate_source_dir(source_dir: str) -> Path:
    """Validate that the source directory exists and is a directory."""
    path = Path(source_dir).resolve()
    if not path.exists():
        print(f"Error: Source directory '{source_dir}' does not exist.")
        sys.exit(1)
    if not path.is_dir():
        print(f"Error: Source path '{source_dir}' is not a directory.")
        sys.exit(1)
    return path


def ensure_output_dir(output_dir: str) -> Path:
    """Ensure that the output directory exists."""
    path = Path(output_dir).resolve()
    path.mkdir(parents=True, exist_ok=True)
    return path


def create_pyproject_toml(
    source_dir: Path,
    name: str,
    version: str,
    author: Optional[str] = None,
    author_email: Optional[str] = None,
    description: Optional[str] = None,
) -> Path:
    """Create or update a pyproject.toml file for the agent."""
    pyproject_path = source_dir / "pyproject.toml"

    # Basic pyproject.toml template
    pyproject_content = f"""[build-system]
requires = ["setuptools>=42", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "{name}"
version = "{version}"
description = "{description or f'Parallels AI Agent - {name}'}"
authors = [{f'{{name = "{author}"}}' if author else ""}]
{f'author-email = "{author_email}"' if author_email else ""}
requires-python = ">=3.8"
classifiers = [
    "Programming Language :: Python :: 3",
    "License :: OSI Approved :: MIT License",
    "Operating System :: OS Independent",
]

[project.entry-points."parallels.ai.agents"]
# Add your agent entry points here
# agent_id = "{name}.agent:AgentClass"
"""

    # Write to file
    with open(pyproject_path, "w") as f:
        f.write(pyproject_content)

    print(f"Created pyproject.toml at {pyproject_path}")
    return pyproject_path


def create_setup_py(
    source_dir: Path,
    name: str,
    version: str,
    author: Optional[str] = None,
    author_email: Optional[str] = None,
    description: Optional[str] = None,
) -> Path:
    """Create or update a setup.py file for the agent."""
    setup_path = source_dir / "setup.py"

    # Basic setup.py template
    setup_content = f"""from setuptools import setup, find_packages

setup(
    name="{name}",
    version="{version}",
    packages=find_packages(),
    install_requires=[
        # Add your dependencies here
    ],
    author="{author or ''}",
    author_email="{author_email or ''}",
    description="{description or f'Parallels AI Agent - {name}'}",
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    entry_points={{
        'parallels.ai.agents': [
            # Add your agent entry points here
            # 'agent_id = {name}.agent:AgentClass',
        ],
    }},
)
"""

    # Write to file
    with open(setup_path, "w") as f:
        f.write(setup_content)

    print(f"Created setup.py at {setup_path}")
    return setup_path


def create_agent_metadata(
    source_dir: Path,
    name: str,
    version: str,
    author: Optional[str] = None,
    author_email: Optional[str] = None,
    description: Optional[str] = None,
) -> Path:
    """Create or update an agent_metadata.json file for the agent."""
    metadata_path = source_dir / f"{name.replace('-', '_')}" / "agent_metadata.json"

    # Ensure the package directory exists
    metadata_path.parent.mkdir(parents=True, exist_ok=True)

    # Basic metadata template
    metadata = {
        "name": name,
        "version": version,
        "description": description or f"Parallels AI Agent - {name}",
        "author": author or "",
        "author_email": author_email or "",
        "agents": [
            # Template for an agent entry
            # {
            #     "id": "example_agent",
            #     "name": "Example Agent",
            #     "description": "An example agent",
            #     "type": "chat",  # or "task", etc.
            #     "class_path": f"{name.replace('-', '_')}.agent:AgentClass"
            # }
        ],
    }

    # Write to file
    with open(metadata_path, "w") as f:
        json.dump(metadata, f, indent=2)

    print(f"Created agent_metadata.json at {metadata_path}")
    return metadata_path


def ensure_init_py(source_dir: Path, name: str) -> None:
    """Ensure __init__.py files exist in all necessary directories."""
    # Package directory
    package_dir = source_dir / name.replace("-", "_")
    package_dir.mkdir(parents=True, exist_ok=True)

    # Create __init__.py in package directory
    init_path = package_dir / "__init__.py"
    if not init_path.exists():
        with open(init_path, "w") as f:
            f.write(f'"""Parallels AI Agent - {name}."""\n\n')
        print(f"Created {init_path}")

    # Check for subdirectories that might need __init__.py
    for subdir in package_dir.iterdir():
        if subdir.is_dir() and not (subdir / "__init__.py").exists():
            with open(subdir / "__init__.py", "w") as f:
                f.write(
                    f'"""Parallels AI Agent - {name} - {subdir.name} module."""\n\n'
                )
            print(f"Created {subdir / '__init__.py'}")


def build_package(source_dir: Path, output_dir: Path) -> Optional[str]:
    """Build the Python package."""
    try:
        # First, make sure build is installed
        subprocess.run(
            [sys.executable, "-m", "pip", "install", "build"],
            check=True,
            capture_output=True,
            text=True,
        )

        # Build the package
        result = subprocess.run(
            [sys.executable, "-m", "build", "--outdir", str(output_dir)],
            cwd=source_dir,
            check=True,
            capture_output=True,
            text=True,
        )

        print("Package built successfully.")
        print(result.stdout)

        # Find the wheel file
        for file in output_dir.iterdir():
            if file.suffix == ".whl":
                return str(file)

        return None

    except subprocess.CalledProcessError as e:
        print(f"Error building package: {e}")
        print(f"Output: {e.stdout}")
        print(f"Error: {e.stderr}")
        return None


def upload_package(
    server_url: str,
    package_name: str,
    version: str,
    file_path: str,
    description: Optional[str] = None,
    author: Optional[str] = None,
    author_email: Optional[str] = None,
    api_key: Optional[str] = None,
) -> bool:
    """Upload the package to the registry server."""
    try:
        # First, make sure requests is installed
        subprocess.run(
            [sys.executable, "-m", "pip", "install", "requests"],
            check=True,
            capture_output=True,
            text=True,
        )

        import requests

        # Prepare upload data
        url = f"{server_url.rstrip('/')}/packages/{package_name}/versions/{version}/upload"
        files = {"file": open(file_path, "rb")}
        data = {}
        if description:
            data["description"] = description
        if author:
            data["author"] = author
        if author_email:
            data["author_email"] = author_email

        headers = {}
        if api_key:
            headers["X-API-Key"] = api_key

        # Upload the package
        response = requests.post(url, files=files, data=data, headers=headers)
        response.raise_for_status()

        result = response.json()
        print(f"Package uploaded successfully: {result.get('message', '')}")
        print(f"File: {result.get('file', '')}")
        print(f"Size: {result.get('size', 0)} bytes")
        print(f"Hash: {result.get('hash', '')}")

        return True

    except Exception as e:
        print(f"Error uploading package: {e}")
        return False


def main():
    """Main entry point for the packager."""
    args = parse_args()

    # Validate source directory
    source_dir = validate_source_dir(args.source_dir)

    # Determine package name
    name = args.name or source_dir.name

    # Ensure output directory exists
    output_dir = ensure_output_dir(args.output_dir)

    # Create package files
    create_pyproject_toml(
        source_dir, name, args.version, args.author, args.author_email, args.description
    )
    create_setup_py(
        source_dir, name, args.version, args.author, args.author_email, args.description
    )
    create_agent_metadata(
        source_dir, name, args.version, args.author, args.author_email, args.description
    )
    ensure_init_py(source_dir, name)

    # Build the package
    package_file = build_package(source_dir, output_dir)

    # Upload if requested
    if args.upload and package_file:
        upload_package(
            args.server,
            name,
            args.version,
            package_file,
            args.description,
            args.author,
            args.author_email,
            args.api_key,
        )


if __name__ == "__main__":
    main()
