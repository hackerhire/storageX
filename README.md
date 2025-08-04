# storageX

A modular Go system for chunking and storing files across multiple cloud services (Dropbox, Google Drive, etc.), with unified authentication, robust error handling, persistent metadata (SQLite), and extensible storage backends.

---

## Features
- **Modular architecture**: Clean separation between chunking, storage orchestration, cloud management, and metadata.
- **Unified CLI**: Upload/download files with a single command-line tool.
- **Config-driven**: All settings via JSON config file (see `config/config.json`).
- **Transactional safety**: Rollback on failed uploads, atomic metadata updates.
- **Persistent metadata**: SQLite-backed file/chunk tracking.
- **Extensible**: Add new cloud providers easily.
- **Robust logging**: Configurable debug/info/error output.
- **CI/CD**: GitHub Actions for test, coverage, and security.

---

## Quick Start

### 1. Install dependencies
- Go 1.20+
- [Graphviz](https://graphviz.gitlab.io/) (for diagrams)
- Python 3 + `pip install diagrams` (for system diagram)

### 2. Build the CLI
```sh
git clone <your-repo-url>
cd storageX
make build
```

### 3. Configure
Edit `config/config.json` to set up cloud credentials, chunk size, logging, etc.

### 4. Usage
#### Upload a file
```sh
./bin/storagex upload /path/to/file.txt
```
#### Download a file
```sh
./bin/storagex download file.txt /path/to/output.txt
```
#### Show version
```sh
./bin/storagex version
```

---

## System Architecture
- See `docs/diagram.py` (Python) or `docs/diagram.go` (Go) for system diagrams.
- Run `python docs/diagram.py` to generate a PNG architecture diagram.

---

## Project Structure
```
cmd/           # CLI entrypoint (main.go)
internal/
  chunker/     # File chunking logic
  cloud/       # Cloud provider interfaces & implementations
  manager/     # StorageManager: cloud ops
  metadata/    # MetadataService: SQLite
  storage/     # StorageService: orchestration
  log/         # Logging
  config/      # Config loading
  defaults/    # Default values
config/        # config.json
.github/       # CI/CD workflows
README.md      # This file
```

---

## Extending
- Add a new provider: implement `CloudStorage` interface in `internal/cloud/` and register in config.
- Add new CLI commands: edit `cmd/main.go` and wire up to storage service.

---

## Testing
```sh
make test
```

---

## License
MIT
