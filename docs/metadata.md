# metadata module

The `MetadataService` manages persistent file and chunk metadata using SQLite. Supports CRUD, existence checks, and transactional safety for file/chunk operations.

## Key Types
- `MetadataService`: Main service
- `ChunkMetadata`, `FileMetadata`: Data models

## Example
```go
meta := metadata.NewMetadataService("meta.db")
meta.AddFile("file.txt", 1234)
meta.AddChunk("file.txt", chunkMeta)
```

## Extension
- Add support for other databases (e.g., Postgres)
- Add advanced queries (e.g., chunk deduplication)
