package manager

import (
	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
	errorx "github.com/sayuyere/storageX/internal/errors"
)

type StorageManager struct {
	cloudSvcs []cloud.CloudStorage
}

func NewStorageManager(cloudSvcs []cloud.CloudStorage) *StorageManager {
	return &StorageManager{
		cloudSvcs: cloudSvcs,
	}
}

func (sm *StorageManager) AddCloudStorage(storage cloud.CloudStorage) {
	sm.cloudSvcs = append(sm.cloudSvcs, storage)
}

func (sm *StorageManager) SearchStorageID(id string) cloud.CloudStorage {

	for _, svc := range sm.cloudSvcs {
		if svc.StorageSystemID() == id {
			return svc
		}
	}
	return nil
}

func (sm *StorageManager) GetCloudSvcForStorage() cloud.CloudStorage {
	if len(sm.cloudSvcs) == 0 {
		panic("no cloud storage configured")
	}
	return sm.cloudSvcs[0] // Assuming first is the default
}

// UploadChunk uploads a chunk to the selected cloud storage
func (sm *StorageManager) UploadChunk(name string, c chunker.Chunk) (cloud.CloudStorage, error) {
	storageLocation := sm.GetCloudSvcForStorage()
	return storageLocation, storageLocation.UploadChunk(name, c.Bytes())
}

// GetChunk gets a chunk from the selected cloud storage
func (sm *StorageManager) GetChunk(storageSystemID string, name string) ([]byte, error) {
	storageLocation := sm.SearchStorageID(storageSystemID)
	if storageLocation == nil {
		return nil, errorx.ErrStorageNotFound
	}
	return storageLocation.GetChunk(name)
}

// DeleteChunk deletes a chunk from the selected cloud storage
func (sm *StorageManager) DeleteChunk(storageSystemID string, name string) error {
	storageLocation := sm.SearchStorageID(storageSystemID)
	if storageLocation == nil {
		return errorx.ErrStorageNotFound
	}
	return storageLocation.DeleteChunk(name)
}
