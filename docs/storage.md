# storage module

The `StorageService` orchestrates file upload/download/delete. It coordinates chunking, metadata, and cloud operations, ensuring transactional safety and rollback on failure.

## Key Types
- `StorageService`: Main orchestration service

## Example
```go
svc := storage.NewStorageService(mgr, meta, chunker)
svc.UploadFile("file.txt")
svc.GetFile("file.txt", writer)
```

## Extension
- Add more orchestration strategies (e.g., parallel upload)
- Add integration with new cloud providers
