package storage

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/config"
	"github.com/sayuyere/storageX/internal/log"
	"github.com/sayuyere/storageX/internal/manager"
	"github.com/sayuyere/storageX/internal/metadata"
)

type StorageService struct {
	manager *manager.StorageManager
	metaSvc *metadata.MetadataService
	chunker *chunker.FileChunker
	lock    sync.RWMutex
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
	s.lock.Lock()
	defer s.lock.Unlock()

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
	// Use only the base name for fileName, not the full path
	fileName := info.Name()
	if fileName == "" || fileName == "." {
		fileName = filePath // fallback
	}

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

	var (
		uploadedChunks []string
		errOnce        sync.Once
		uploadErr      error
		mu             sync.Mutex
		wg             sync.WaitGroup
		maxParallel    = config.GetConfig().Parallel.Upload // adjust as needed
		sem            = make(chan struct{}, maxParallel)
	)

	for chunk := range chunks {
		if chunk.Err != nil {
			errOnce.Do(func() { uploadErr = chunk.Err })
			break
		}
		c := string(chunk.Checksum[:])
		if exists, _ := s.metaSvc.ChunkExists(chunk.Name); exists {
			errOnce.Do(func() { uploadErr = fmt.Errorf("chunk %s already exists in metadata", chunk.Name) })
			break
		}
		sem <- struct{}{} // acquire slot
		wg.Add(1)
		go func(chunk chunker.Chunk) {
			defer wg.Done()
			defer func() { <-sem }() // release slot
			storageLocation, err := s.manager.UploadChunk(chunk.Name, chunk)
			if err != nil {
				errOnce.Do(func() { uploadErr = err })
				return
			}
			mu.Lock()
			uploadedChunks = append(uploadedChunks, chunk.Name)
			mu.Unlock()
			err = s.metaSvc.AddChunk(fileName, metadata.ChunkMetadata{
				ChunkName: chunk.Name,
				Size:      int64(len(chunk.Data)),
				Checksum:  c,
				Index:     int(chunk.Index),
				FileName:  fileName,
				Storage:   storageLocation.StorageSystemID(),
			})
			if err != nil {
				errOnce.Do(func() { uploadErr = err })
			}
		}(chunk)
	}
	wg.Wait()
	if uploadErr != nil {
		rollback(uploadedChunks)
		return uploadErr
	}
	return nil
}

// GetFile reconstructs the file from chunks and writes to writer
func (s *StorageService) GetFile(fileName string, w io.Writer) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	metas, err := s.metaSvc.ListChunks(fileName)
	if err != nil {
		return err
	}
	var (
		errOnce     sync.Once
		getErr      error
		wg          sync.WaitGroup
		maxParallel = config.GetConfig().Parallel.Download
		sem         = make(chan struct{}, maxParallel)
		results     = make([][]byte, len(metas))
	)
	for i, meta := range metas {
		sem <- struct{}{}
		wg.Add(1)
		go func(i int, meta metadata.ChunkMetadata) {
			defer wg.Done()
			defer func() { <-sem }()
			log.Info("Retrieving chunk: %s", meta.ChunkName, meta.Size)
			data, err := s.manager.GetChunk(meta.Storage, meta.ChunkName)
			if err != nil {
				errOnce.Do(func() { getErr = err })
				return
			}
			results[i] = data[chunker.ChunkMetadataSize:]
		}(i, meta)
	}
	wg.Wait()
	if getErr != nil {
		return getErr
	}
	for _, chunkData := range results {
		if chunkData == nil {
			continue // skip missing chunks
		}
		_, err := w.Write(chunkData)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteFile deletes all chunks for a file and removes metadata.
// If any chunk deletion fails, it continues deleting the rest and collects errors.
func (s *StorageService) DeleteFile(fileName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

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
