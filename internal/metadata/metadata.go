package metadata

import (
	"sync"
)

type ChunkMetadata struct {
	ChunkName string
	Size      int64
	Checksum  string
}

type MetadataService struct {
	mu      sync.RWMutex
	entries map[string]ChunkMetadata // key: chunk name
}

func NewMetadataService() *MetadataService {
	return &MetadataService{
		entries: make(map[string]ChunkMetadata),
	}
}

func (m *MetadataService) Add(meta ChunkMetadata) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[meta.ChunkName] = meta
}

func (m *MetadataService) Get(chunkName string) (ChunkMetadata, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, ok := m.entries[chunkName]
	return meta, ok
}

func (m *MetadataService) List() []ChunkMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ChunkMetadata, 0, len(m.entries))
	for _, v := range m.entries {
		result = append(result, v)
	}
	return result
}
