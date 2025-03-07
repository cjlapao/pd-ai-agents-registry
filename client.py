#!/usr/bin/env python3
"""
Parallels AI Agents Registry Client

A command line client to interact with the Parallels AI Agents Registry Server.
"""

import argparse
import requests
import os
import sys
import json
from pathlib import Path
from typing import Optional, Dict, Any, List
from tabulate import tabulate


class RegistryClient:
    """Client for interacting with the Parallels AI Agents Registry Server."""

    def __init__(
        self, base_url: str = "http://localhost:8000", api_key: Optional[str] = None
    ):
        """Initialize the client with the server URL and optional API key."""
        self.base_url = base_url.rstrip("/")
        self.headers = {}
        if api_key:
            self.headers["X-API-Key"] = api_key

    def list_packages(self) -> List[Dict[str, Any]]:
        """List all available packages."""
        response = requests.get(f"{self.base_url}/packages", headers=self.headers)
        response.raise_for_status()
        return response.json()

    def get_package(self, package_name: str) -> Dict[str, Any]:
        """Get details for a specific package."""
        response = requests.get(
            f"{self.base_url}/packages/{package_name}", headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def list_versions(self, package_name: str) -> List[Dict[str, Any]]:
        """List all versions for a specific package."""
        response = requests.get(
            f"{self.base_url}/packages/{package_name}/versions", headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def get_version(self, package_name: str, version: str) -> Dict[str, Any]:
        """Get details for a specific package version."""
        response = requests.get(
            f"{self.base_url}/packages/{package_name}/versions/{version}",
            headers=self.headers,
        )
        response.raise_for_status()
        return response.json()

    def upload_package(
        self,
        package_name: str,
        version: str,
        file_path: str,
        description: Optional[str] = None,
        author: Optional[str] = None,
        author_email: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Upload a package file for a specific version."""
        files = {"file": open(file_path, "rb")}
        data = {}
        if description:
            data["description"] = description
        if author:
            data["author"] = author
        if author_email:
            data["author_email"] = author_email

        response = requests.post(
            f"{self.base_url}/packages/{package_name}/versions/{version}/upload",
            headers=self.headers,
            files=files,
            data=data,
        )
        response.raise_for_status()
        return response.json()

    def download_package(
        self, package_name: str, version: str, filename: str, output_dir: str = "."
    ) -> str:
        """Download a package file."""
        output_path = os.path.join(output_dir, filename)
        response = requests.get(
            f"{self.base_url}/download/{package_name}/{version}/{filename}",
            headers=self.headers,
            stream=True,
        )
        response.raise_for_status()

        with open(output_path, "wb") as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)

        return output_path

    def delete_package(
        self, package_name: str, version: str, filename: str
    ) -> Dict[str, Any]:
        """Delete a package file."""
        response = requests.delete(
            f"{self.base_url}/packages/{package_name}/versions/{version}/{filename}",
            headers=self.headers,
        )
        response.raise_for_status()
        return response.json()


def parse_args():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description="Parallels AI Agents Registry Client")

    # Server configuration
    parser.add_argument(
        "--server", default="http://localhost:8000", help="Registry server URL"
    )
    parser.add_argument("--api-key", help="API key for authentication")

    # Commands
    subparsers = parser.add_subparsers(dest="command", help="Command to execute")

    # List packages
    list_parser = subparsers.add_parser("list", help="List packages")

    # Get package details
    get_parser = subparsers.add_parser("get", help="Get package details")
    get_parser.add_argument("package", help="Package name")

    # List versions
    versions_parser = subparsers.add_parser("versions", help="List package versions")
    versions_parser.add_argument("package", help="Package name")

    # Get version details
    version_parser = subparsers.add_parser("version", help="Get version details")
    version_parser.add_argument("package", help="Package name")
    version_parser.add_argument("version", help="Version to get")

    # Upload package
    upload_parser = subparsers.add_parser("upload", help="Upload a package")
    upload_parser.add_argument("package", help="Package name")
    upload_parser.add_argument("version", help="Version to upload")
    upload_parser.add_argument("file", help="Path to package file")
    upload_parser.add_argument("--description", help="Package description")
    upload_parser.add_argument("--author", help="Package author")
    upload_parser.add_argument("--author-email", help="Package author email")

    # Download package
    download_parser = subparsers.add_parser("download", help="Download a package")
    download_parser.add_argument("package", help="Package name")
    download_parser.add_argument("version", help="Version to download")
    download_parser.add_argument("filename", help="Filename to download")
    download_parser.add_argument("--output-dir", default=".", help="Output directory")

    # Delete package
    delete_parser = subparsers.add_parser("delete", help="Delete a package")
    delete_parser.add_argument("package", help="Package name")
    delete_parser.add_argument("version", help="Version to delete")
    delete_parser.add_argument("filename", help="Filename to delete")

    return parser.parse_args()


def display_packages(packages: List[Dict[str, Any]]):
    """Display packages in a formatted table."""
    rows = []
    for package in packages:
        description = package.get("metadata", {}).get("description", "")
        latest_version = package.get("metadata", {}).get("latest_version", "")
        rows.append(
            [
                package["name"],
                latest_version,
                description,
                ", ".join(package["versions"]),
            ]
        )

    print(
        tabulate(
            rows,
            headers=["Name", "Latest Version", "Description", "Available Versions"],
            tablefmt="pretty",
        )
    )


def display_versions(versions: List[Dict[str, Any]]):
    """Display versions in a formatted table."""
    rows = []
    for version in versions:
        files = ", ".join([f["name"] for f in version["files"]])
        rows.append([version["version"], version["upload_date"], files])

    print(
        tabulate(rows, headers=["Version", "Upload Date", "Files"], tablefmt="pretty")
    )


def display_files(files: List[Dict[str, Any]]):
    """Display files in a formatted table."""
    rows = []
    for file in files:
        size_kb = file["size"] / 1024
        rows.append([file["name"], f"{size_kb:.2f} KB", file["last_modified"]])

    print(
        tabulate(rows, headers=["Filename", "Size", "Last Modified"], tablefmt="pretty")
    )


def main():
    """Main entry point for the client."""
    args = parse_args()

    # Create client
    client = RegistryClient(args.server, args.api_key)

    try:
        # Execute command
        if args.command == "list":
            packages = client.list_packages()
            display_packages(packages)

        elif args.command == "get":
            package = client.get_package(args.package)
            print(f"Package: {package['name']}")
            print(f"Available Versions: {', '.join(package['versions'])}")
            print("Metadata:")
            print(json.dumps(package["metadata"], indent=2))

        elif args.command == "versions":
            versions = client.list_versions(args.package)
            display_versions(versions)

        elif args.command == "version":
            version = client.get_version(args.package, args.version)
            print(f"Package: {version['package']}")
            print(f"Version: {version['version']}")
            print(f"Upload Date: {version['upload_date']}")
            print("Files:")
            display_files(version["files"])

        elif args.command == "upload":
            result = client.upload_package(
                args.package,
                args.version,
                args.file,
                args.description,
                args.author,
                args.author_email,
            )
            print(f"Upload successful: {result['message']}")
            print(f"File: {result['file']}")
            print(f"Size: {result['size']} bytes")
            print(f"Hash: {result['hash']}")

        elif args.command == "download":
            output_path = client.download_package(
                args.package, args.version, args.filename, args.output_dir
            )
            print(f"Downloaded to: {output_path}")

        elif args.command == "delete":
            result = client.delete_package(args.package, args.version, args.filename)
            print(result["message"])

        else:
            print("No command specified. Use --help for usage information.")

    except requests.exceptions.RequestException as e:
        print(f"Error: {e}")
        if hasattr(e, "response") and e.response is not None:
            print(f"Server response: {e.response.text}")
        sys.exit(1)


if __name__ == "__main__":
    main()
