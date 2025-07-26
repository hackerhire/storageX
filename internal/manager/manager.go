package manager

import (
	"github.com/sayuyere/storageX/internal/cloud"
	"github.com/sayuyere/storageX/internal/metadata"
)

type StorageManager struct {
	cloudSvc cloud.CloudStorage
	metaSvc  *metadata.MetadataService
}

func NewStorageManager(cloudSvc cloud.CloudStorage, metaSvc *metadata.MetadataService) *StorageManager {
	return &StorageManager{
		cloudSvc: cloudSvc,
		metaSvc:  metaSvc,
	}
}

func (sm *StorageManager) UploadChunk(name string, data []byte, checksum string) error {
	err := sm.cloudSvc.UploadChunk(name, data)
	if err != nil {
		return err
	}
	meta := metadata.ChunkMetadata{
		ChunkName: name,
		Size:      int64(len(data)),
		Checksum:  checksum,
	}
	sm.metaSvc.Add(meta)
	return nil
}

func (sm *StorageManager) GetChunk(name string) ([]byte, *metadata.ChunkMetadata, error) {
	data, err := sm.cloudSvc.GetChunk(name)
	if err != nil {
		return nil, nil, err
	}
	meta, _ := sm.metaSvc.Get(name)
	return data, &meta, nil
}

func (sm *StorageManager) DeleteChunk(name string) error {
	err := sm.cloudSvc.DeleteChunk(name)
	if err != nil {
		return err
	}
	// Optionally remove metadata
	return nil
}
