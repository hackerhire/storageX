# storageX

A Golang system to chunk a service into multiple chunks and store them into multiple cloud services (e.g., AWS S3, Google Cloud Storage).

## Features
- Chunking logic for large files/services
- Pluggable cloud storage backends
- Example implementations for AWS S3 and Google Cloud Storage
- Easy to add new cloud providers

## Structure
- `cmd/` - Main application entrypoint
- `internal/chunker/` - Chunking logic
- `internal/cloud/` - Cloud storage abstractions and implementations
- `config/` - Configuration files

## Usage
```sh
make build
./bin/storageX --config config/config.yaml
```

## Requirements
- Go 1.21+
- AWS/GCP credentials (for respective providers)

## Extending
To add a new cloud provider, implement the `CloudStorage` interface in `internal/cloud/`.
