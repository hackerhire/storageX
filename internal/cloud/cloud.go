package cloud

// CloudStorage defines the interface for cloud storage providers
// Implement UploadChunk for each provider

type CloudStorage interface {
	UploadChunk(name string, data []byte) error
}
