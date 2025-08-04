# chunker module

Handles splitting files or byte slices into fixed-size chunks with metadata (checksum, index, name). Supports streaming and serialization for efficient upload/download.

## Key Types
- `Chunk`: Data, Checksum, Index, Name, etc.
- `FileChunker`: Main chunking logic

## Example
```go
chunker := chunker.NewFileChunker(1024*1024)
chunks, err := chunker.ChunkFileStream(file)
```

## Extension
- Add new chunking strategies (e.g., variable size, deduplication)
