package manager

import (
	"crypto/sha256"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
	"github.com/sayuyere/storageX/internal/metadata"
)

type StorageManager struct {
	cloudSvcs []cloud.CloudStorage
	metaSvc   *metadata.MetadataService
}

func NewStorageManager(cloudSvcs []cloud.CloudStorage, metaSvc *metadata.MetadataService) *StorageManager {
	return &StorageManager{
		cloudSvcs: cloudSvcs,
		metaSvc:   metaSvc,
	}
}

func (sm *StorageManager) GetCloudSvcForStorage() cloud.CloudStorage {
	if len(sm.cloudSvcs) == 0 {
		panic("no cloud storage configured")
	}
	return sm.cloudSvcs[0] // Assuming first is the default

}

// UploadChunk uploads to all configured cloud storages
func (sm *StorageManager) UploadChunk(filename, name string, c chunker.Chunk, checksum string) error {
	storageLocation := sm.GetCloudSvcForStorage()
	return storageLocation.UploadChunk(name, c.Bytes())
}

// // GetChunk tries to get from the first available storage
// func (sm *StorageManager) GetChunk(name string) ([]byte, *metadata.ChunkMetadata, error) {
// 	for _, svc := range sm.cloudSvcs {
// 		data, err := svc.GetChunk(name)
// 		if err == nil {
// 			meta, _ := sm.metaSvc.Get(name)
// 			return data, &meta, nil
// 		}
// 	}
// 	return nil, nil, cloud.ErrDriveDownload // or a custom error
// }

// GetChunk gets a chunk from the storage system with the given ID
func (sm *StorageManager) GetChunkFrom(storageSystemID, name string) ([]byte, *metadata.ChunkMetadata, error) {
	for _, svc := range sm.cloudSvcs {
		if svc.StorageSystemID() == storageSystemID {
			data, err := svc.GetChunk(name)
			if err == nil {
				chunk := chunker.ChunkFromBytes(data)
				checksum := sha256.Sum256(chunk.Data)
				checksumString := string(checksum[:]) // Convert to string for metadata
				// meta, _ := sm.metaSvc.Get(name)
				return data, &metadata.ChunkMetadata{
					ChunkName: chunk.Name,
					Size:      int64(len(chunk.Data)),
					Checksum:  checksumString,
					Index:     int(chunk.Index),
					Storage:   svc.StorageSystemID(),
					FileName:  name, // Assuming chunk.Name is the file name

				}, nil
			}
			return nil, nil, err
		}
	}
	return nil, nil, cloud.ErrDriveDownload // or a custom error
}

// DeleteChunk deletes from all storages
func (sm *StorageManager) DeleteChunk(name string) error {
	var lastErr error
	for _, svc := range sm.cloudSvcs {
		err := svc.DeleteChunk(name)
		if err != nil {
			lastErr = err
		}
	}
	// Optionally remove metadata
	return lastErr
}
