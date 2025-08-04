# manager module

The `StorageManager` coordinates cloud operations. It selects the right provider, uploads/downloads/deletes chunks, and abstracts multi-cloud logic from the orchestration layer.

## Key Types
- `StorageManager`: Holds a list of `CloudStorage` providers

## Example
```go
mgr := manager.NewStorageManager([]cloud.CloudStorage{dropbox, gdrive})
mgr.UploadChunk(name, chunk)
```

## Extension
- Add provider selection strategies (e.g., round-robin, by available space)
