package storage

import (
	"io"
	"os"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/log"
	"github.com/sayuyere/storageX/internal/manager"
	"github.com/sayuyere/storageX/internal/metadata"
)

type StorageService struct {
	manager *manager.StorageManager
	metaSvc *metadata.MetadataService
	chunker *chunker.FileChunker
}

func NewStorageService(mgr *manager.StorageManager, meta *metadata.MetadataService, ch *chunker.FileChunker) *StorageService {
	return &StorageService{
		manager: mgr,
		metaSvc: meta,
		chunker: ch,
	}
}

// UploadFile splits the file into chunks and uploads them
func (s *StorageService) UploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	chunks, err := s.chunker.ChunkFileStream(file)
	if err != nil {
		return err
	}
	for chunk := range chunks {
		if chunk.Err != nil {
			err := s.metaSvc.DeleteFile(file.Name())
			if err != nil {
				log.Info("Error While deleting item from metadata service :: ", err)
			}
			return chunk.Err
		}
		c := string(chunk.Checksum[:])
		systemToUse := s.manager.GetCloudSvcForStorage().StorageSystemID() // Relly bad as we are reserving have some reservation logic with sometimeout if not used but keeping this simple for now
		s.metaSvc.AddChunk(file.Name(), metadata.ChunkMetadata{
			ChunkName: chunk.Name,
			Size:      int64(len(chunk.Data)),
			Checksum:  c,
			Index:     int(chunk.Index),
			FileName:  file.Name(),
			Storage:   systemToUse,
		})
		err := s.manager.UploadChunk(file.Name(), chunk.Name, chunk, c)
		if err != nil {
			err := s.metaSvc.DeleteFile(file.Name())
			return err
		}
	}
	return nil
}

// GetFile reconstructs the file from chunks and writes to writer
func (s *StorageService) GetFile(fileName string, w io.Writer) error {
	metas, err := s.metaSvc.ListChunks(fileName)
	if err != nil {
		return err
	}
	for _, meta := range metas {
		data, _, err := s.manager.GetChunkFrom(meta.Storage, meta.ChunkName)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteFile deletes all chunks for a file
func (s *StorageService) DeleteFile(fileName string) error {
	metas, err := s.metaSvc.ListChunks(fileName)
	if err != nil {
		return err
	}
	for _, meta := range metas {
		err := s.manager.DeleteChunk(meta.ChunkName)
		if err != nil {
			return err
		}
	}
	// Optionally, delete the file metadata
	err = s.metaSvc.DeleteFile(fileName)
	return err
}
