# cloud module

Defines the `CloudStorage` interface for all cloud providers. Each provider (Dropbox, Google Drive, etc.) implements this interface for upload, download, and delete operations.

## Key Interface
- `CloudStorage` (UploadChunk, GetChunk, DeleteChunk, StorageSystemID, etc.)

## Example
```go
type DropboxStorage struct {}
func (d *DropboxStorage) UploadChunk(name string, data []byte) error { ... }
```

## Extension
- Add new providers by implementing `CloudStorage` and registering in config.
