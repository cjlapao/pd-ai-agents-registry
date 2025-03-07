from fastapi import (
    FastAPI,
    HTTPException,
    UploadFile,
    File,
    Depends,
    status,
    Form,
    Request,
)
from fastapi.responses import JSONResponse, FileResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from typing import List, Optional, Dict, Any
import os
import shutil
import json
from datetime import datetime
import uuid
from pathlib import Path
import hashlib
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Create the FastAPI app
app = FastAPI(
    title="Parallels AI Agents Registry",
    description="A registry server for Parallels AI Agents packages",
    version="0.1.0",
)

# Configuration from environment variables
PORT = int(os.getenv("PORT", 8000))
HOST = os.getenv("HOST", "0.0.0.0")
PACKAGES_DIR = Path(os.getenv("PACKAGES_DIR", "./data/packages"))
TEMP_DIR = Path(os.getenv("TEMP_DIR", "./data/temp"))
METADATA_DIR = Path(os.getenv("METADATA_DIR", "./data/metadata"))
API_KEY_REQUIRED = os.getenv("API_KEY_REQUIRED", "false").lower() == "true"
API_KEY = os.getenv("API_KEY", "")
MAX_UPLOAD_SIZE = int(os.getenv("MAX_UPLOAD_SIZE", 104857600))  # Default 100MB

# Create directories if they don't exist
for directory in [PACKAGES_DIR, TEMP_DIR, METADATA_DIR]:
    directory.mkdir(parents=True, exist_ok=True)


# Error handler
@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    return JSONResponse(status_code=exc.status_code, content={"message": exc.detail})


@app.get("/")
async def root():
    """Root endpoint to check if the service is running"""
    return {"status": "running", "service": "Parallels AI Agents Registry"}


@app.get("/packages")
async def list_packages():
    """List all available packages"""
    packages = []

    # Get all package directories
    if PACKAGES_DIR.exists():
        for pkg_dir in PACKAGES_DIR.iterdir():
            if pkg_dir.is_dir():
                package_name = pkg_dir.name

                # Get package metadata
                metadata_path = METADATA_DIR / f"{package_name}.json"
                metadata = {}
                if metadata_path.exists():
                    with open(metadata_path, "r") as f:
                        metadata = json.load(f)

                # Add package with its metadata
                packages.append(
                    {
                        "name": package_name,
                        "versions": [v.name for v in pkg_dir.iterdir() if v.is_dir()],
                        "metadata": metadata,
                    }
                )

    return packages


@app.get("/packages/{package_name}")
async def get_package(package_name: str):
    """Get details for a specific package"""
    pkg_dir = PACKAGES_DIR / package_name

    if not pkg_dir.exists() or not pkg_dir.is_dir():
        raise HTTPException(
            status_code=404, detail=f"Package '{package_name}' not found"
        )

    # Get versions
    versions = [v.name for v in pkg_dir.iterdir() if v.is_dir()]

    # Get package metadata
    metadata_path = METADATA_DIR / f"{package_name}.json"
    metadata = {}
    if metadata_path.exists():
        with open(metadata_path, "r") as f:
            metadata = json.load(f)

    return {"name": package_name, "versions": versions, "metadata": metadata}


@app.get("/packages/{package_name}/versions")
async def list_versions(package_name: str):
    """List all versions for a specific package"""
    pkg_dir = PACKAGES_DIR / package_name

    if not pkg_dir.exists() or not pkg_dir.is_dir():
        raise HTTPException(
            status_code=404, detail=f"Package '{package_name}' not found"
        )

    versions = []
    for version_dir in pkg_dir.iterdir():
        if version_dir.is_dir():
            # Get version files
            files = []
            for file_path in version_dir.iterdir():
                if file_path.is_file():
                    files.append(
                        {
                            "name": file_path.name,
                            "size": file_path.stat().st_size,
                            "last_modified": datetime.fromtimestamp(
                                file_path.stat().st_mtime
                            ).isoformat(),
                        }
                    )

            versions.append(
                {
                    "version": version_dir.name,
                    "files": files,
                    "upload_date": datetime.fromtimestamp(
                        version_dir.stat().st_mtime
                    ).isoformat(),
                }
            )

    return versions


@app.get("/packages/{package_name}/versions/{version}")
async def get_version(package_name: str, version: str):
    """Get details for a specific package version"""
    version_dir = PACKAGES_DIR / package_name / version

    if not version_dir.exists() or not version_dir.is_dir():
        raise HTTPException(
            status_code=404,
            detail=f"Version '{version}' for package '{package_name}' not found",
        )

    files = []
    for file_path in version_dir.iterdir():
        if file_path.is_file():
            files.append(
                {
                    "name": file_path.name,
                    "size": file_path.stat().st_size,
                    "last_modified": datetime.fromtimestamp(
                        file_path.stat().st_mtime
                    ).isoformat(),
                }
            )

    return {
        "package": package_name,
        "version": version,
        "files": files,
        "upload_date": datetime.fromtimestamp(version_dir.stat().st_mtime).isoformat(),
    }


@app.post("/packages/{package_name}/versions/{version}/upload")
async def upload_package(
    package_name: str,
    version: str,
    file: UploadFile = File(...),
    description: Optional[str] = Form(None),
    author: Optional[str] = Form(None),
    author_email: Optional[str] = Form(None),
):
    """Upload a package file for a specific version"""
    # Create directories if they don't exist
    package_dir = PACKAGES_DIR / package_name
    package_dir.mkdir(parents=True, exist_ok=True)

    version_dir = package_dir / version
    version_dir.mkdir(parents=True, exist_ok=True)

    # Save the file
    file_path = (
        version_dir / file.filename if file.filename else version_dir / "unknown.whl"
    )
    temp_file_path = TEMP_DIR / f"{uuid.uuid4()}.tmp"

    try:
        # Create a temporary file
        with open(temp_file_path, "wb") as buffer:
            # Read and write in chunks to handle large files
            chunk_size = 1024 * 1024  # 1MB chunks
            while True:
                chunk = await file.read(chunk_size)
                if not chunk:
                    break
                buffer.write(chunk)

        # Calculate SHA256 hash
        sha256_hash = hashlib.sha256()
        with open(temp_file_path, "rb") as f:
            # Use binary mode to read the file, ensure we're always reading bytes
            for chunk in iter(lambda: f.read(4096), b""):
                sha256_hash.update(chunk)  # chunk is guaranteed to be bytes
        file_hash = sha256_hash.hexdigest()

        # Move to final location
        shutil.move(temp_file_path, file_path)

        # Update or create package metadata
        metadata_path = METADATA_DIR / f"{package_name}.json"
        metadata = {}

        if metadata_path.exists():
            with open(metadata_path, "r") as f:
                metadata = json.load(f)

        # Update metadata
        metadata.update(
            {
                "name": package_name,
                "latest_version": version,
                "description": description or metadata.get("description", ""),
                "author": author or metadata.get("author", ""),
                "author_email": author_email or metadata.get("author_email", ""),
                "last_updated": datetime.now().isoformat(),
            }
        )

        # Add version-specific info
        if "versions" not in metadata:
            metadata["versions"] = {}

        # Get existing files for this version or initialize empty list
        existing_files = metadata.get("versions", {}).get(version, {}).get("files", [])

        # Add the new file (replacing if exists with same name)
        new_files = [f for f in existing_files if f["name"] != file.filename]
        new_files.append(
            {
                "name": file.filename,
                "size": file_path.stat().st_size,
                "hash": file_hash,
            }
        )

        metadata["versions"][version] = {
            "upload_date": datetime.now().isoformat(),
            "files": new_files,
        }

        # Save updated metadata
        with open(metadata_path, "w") as f:
            json.dump(metadata, f, indent=2)

        return {
            "message": f"File '{file.filename}' uploaded successfully",
            "package": package_name,
            "version": version,
            "file": file.filename,
            "size": file_path.stat().st_size,
            "hash": file_hash,
        }

    except Exception as e:
        # Clean up in case of error
        if os.path.exists(temp_file_path):
            os.remove(temp_file_path)

        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to upload file: {str(e)}",
        )


@app.get("/download/{package_name}/{version}/{filename}")
async def download_package(package_name: str, version: str, filename: str):
    """Download a specific package file"""
    file_path = PACKAGES_DIR / package_name / version / filename

    if not file_path.exists() or not file_path.is_file():
        raise HTTPException(
            status_code=404,
            detail=f"File '{filename}' not found for package '{package_name}' version '{version}'",
        )

    return FileResponse(
        path=file_path, filename=filename, media_type="application/octet-stream"
    )


@app.delete("/packages/{package_name}/versions/{version}/{filename}")
async def delete_package_file(package_name: str, version: str, filename: str):
    """Delete a specific package file"""
    file_path = PACKAGES_DIR / package_name / version / filename

    if not file_path.exists() or not file_path.is_file():
        raise HTTPException(
            status_code=404,
            detail=f"File '{filename}' not found for package '{package_name}' version '{version}'",
        )

    try:
        # Delete the file
        file_path.unlink()

        # Update metadata
        metadata_path = METADATA_DIR / f"{package_name}.json"
        if metadata_path.exists():
            with open(metadata_path, "r") as f:
                metadata = json.load(f)

            # Remove file from metadata
            if "versions" in metadata and version in metadata["versions"]:
                metadata["versions"][version]["files"] = [
                    f
                    for f in metadata["versions"][version]["files"]
                    if f["name"] != filename
                ]

                # Save updated metadata
                with open(metadata_path, "w") as f:
                    json.dump(metadata, f, indent=2)

        # Check if version directory is empty
        version_dir = PACKAGES_DIR / package_name / version
        if not any(version_dir.iterdir()):
            # Remove empty version directory
            version_dir.rmdir()

            # Remove version from metadata
            if metadata_path.exists():
                with open(metadata_path, "r") as f:
                    metadata = json.load(f)

                if "versions" in metadata and version in metadata["versions"]:
                    del metadata["versions"][version]

                    # Update latest version if needed
                    if metadata.get("latest_version") == version:
                        versions = list(metadata.get("versions", {}).keys())
                        if versions:
                            # Sort versions and get the latest
                            metadata["latest_version"] = sorted(versions)[-1]
                        else:
                            metadata["latest_version"] = ""

                    # Save updated metadata
                    with open(metadata_path, "w") as f:
                        json.dump(metadata, f, indent=2)

        # Check if package directory is empty
        package_dir = PACKAGES_DIR / package_name
        if not any(package_dir.iterdir()):
            # Remove empty package directory
            package_dir.rmdir()

            # Remove package metadata
            if metadata_path.exists():
                metadata_path.unlink()

        return {"message": f"File '{filename}' deleted successfully"}

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to delete file: {str(e)}",
        )


# Add more endpoints as needed

if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host=HOST, port=PORT, reload=True)
