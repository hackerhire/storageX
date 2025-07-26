package cloud

// CloudStorage defines the interface for cloud storage providers
// Implement UploadChunk, GetChunk, and DeleteChunk for each provider

type CloudStorage interface {
	UploadChunk(name string, data []byte) error
	GetChunk(name string) ([]byte, error)
	DeleteChunk(name string) error
}
