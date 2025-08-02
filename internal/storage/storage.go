package storage

import (
	"fmt"
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

// UploadFile splits the file into chunks and uploads them, updating metadata
func (s *StorageService) UploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		log.Error("failed to get file info: %v", err)
		return err
	}
	fileName := file.Name()

	// Check if file already exists in metadata
	exists, err := s.metaSvc.FileExists(fileName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("file %s already exists in metadata", fileName)
	}

	// Add file entry to metadata first to prevent race conditions
	if err := s.metaSvc.AddFile(fileName, uint64(info.Size())); err != nil {
		return err
	}

	// Rollback function to clean up metadata and cloud storage
	rollback := func(uploadedChunks []string) {
		for _, chunkName := range uploadedChunks {
			// Assuming the storage system ID can be obtained from the manager
			chunkMetaData, found := s.metaSvc.GetChunk(chunkName)
			if !found {
				log.Error("chunk %s not found in metadata, skipping deletion", chunkName)
				continue
			}
			_ = s.manager.DeleteChunk(chunkMetaData.Storage, chunkName)
		}
		_ = s.metaSvc.DeleteFile(fileName)

	}

	chunks, err := s.chunker.ChunkFileStream(file)
	if err != nil {
		_ = s.metaSvc.DeleteFile(fileName)
		return err
	}

	var uploadedChunks []string
	for chunk := range chunks {
		log.Info("Processing chunk: %s", chunk.Name, len(chunk.Bytes()))
		if chunk.Err != nil {
			rollback(uploadedChunks)
			return chunk.Err
		}
		c := string(chunk.Checksum[:])
		if exists, _ := s.metaSvc.ChunkExists(chunk.Name); exists {
			rollback(uploadedChunks)
			return fmt.Errorf("chunk %s already exists in metadata", chunk.Name)
		}
		storageLocation, err := s.manager.UploadChunk(chunk.Name, chunk)
		if err != nil {
			rollback(uploadedChunks)
			return err
		}
		uploadedChunks = append(uploadedChunks, chunk.Name)
		err = s.metaSvc.AddChunk(fileName, metadata.ChunkMetadata{
			ChunkName: chunk.Name,
			Size:      int64(len(chunk.Data)),
			Checksum:  c,
			Index:     int(chunk.Index),
			FileName:  fileName,
			Storage:   storageLocation.StorageSystemID(),
		})
		if err != nil {
			rollback(uploadedChunks)
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
		log.Info("Retrieving chunk: %s", meta.ChunkName, meta.Size)
		data, err := s.manager.GetChunk(meta.Storage, meta.ChunkName)
		if err != nil {
			return err
		}
		_, err = w.Write(data[chunker.ChunkMetadataSize:])
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteFile deletes all chunks for a file and removes metadata.
// If any chunk deletion fails, it continues deleting the rest and collects errors.
func (s *StorageService) DeleteFile(fileName string) error {
	metas, err := s.metaSvc.ListChunks(fileName)
	if err != nil {
		return err
	}

	var deleteErrs []error
	for _, meta := range metas {
		if err := s.manager.DeleteChunk(meta.Storage, meta.ChunkName); err != nil {
			deleteErrs = append(deleteErrs, fmt.Errorf("failed to delete chunk %s: %w", meta.ChunkName, err))
		}
	}

	if err := s.metaSvc.DeleteFile(fileName); err != nil {
		deleteErrs = append(deleteErrs, fmt.Errorf("failed to delete file metadata: %w", err))
	}

	if len(deleteErrs) > 0 {
		return fmt.Errorf("DeleteFile encountered errors: %v", deleteErrs)
	}
	return nil
}
