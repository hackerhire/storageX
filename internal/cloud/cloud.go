package cloud

// CloudStorage defines the interface for cloud storage providers
// Implement UploadChunk, GetChunk, and DeleteChunk for each provider

type CloudStorage interface {
	UploadChunk(name string, data []byte) error
	GetChunk(name string) ([]byte, error)
	DeleteChunk(name string) error
	GetRemainingSize() (int64, error) // New method to get storage unit size
	StorageSystemID() string          // Returns a unique ID or name for the storage system
}
