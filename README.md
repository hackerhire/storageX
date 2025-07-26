# storageX

A modular Go system to chunk and store files in multiple cloud services (Dropbox, Google Drive, etc.) with unified authentication, error handling, and extensible storage backends.

## Features
- Streaming chunking logic for large files
- Pluggable, unified cloud storage backends (Dropbox, Google Drive, etc.)
- Unified authentication and error handling
- Persistent metadata tracking (SQLite)
- Storage manager for multi-cloud orchestration
- Robust logging (zap)
- CI/CD with test coverage and security scanning
- Easy to add new cloud providers

## Structure
- `cmd/` - Main application entrypoint
- `internal/chunker/` - Streaming chunking logic (singleton, config-driven)
- `internal/cloud/` - Cloud storage abstractions and implementations (Dropbox, Google Drive, etc.)
- `internal/manager/` - Storage manager for orchestrating multi-cloud operations
- `internal/metadata/` - Persistent chunk and file metadata (SQLite)
- `internal/storage/` - High-level file storage/retrieval API
- `internal/log/` - Logging utilities (zap)
- `internal/config/` - App config management (JSON)
- `config/` - Configuration files (JSON)

## Usage
```sh
make build
./bin/storageX --config config/config.json
```

## Requirements
- Go 1.21+
- Dropbox/Google Drive credentials (for respective providers)
- SQLite (for metadata)

## Extending
To add a new cloud provider, implement the `CloudStorage` interface in `internal/cloud/` and register it in your config/manager.

## Testing
```sh
go test ./...
```

## CI/CD
- GitHub Actions: build, test, coverage, security (CodeQL)

---
For more details, see the code and comments in each module.
