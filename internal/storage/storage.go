package storage

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/config"
	errorx "github.com/sayuyere/storageX/internal/errors" // new error package alias
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
		log.Error("%v: %v", errorx.ErrFileInfoFetchFailed, err)
		return errorx.Wrap(errorx.ErrFileInfoFetchFailed, err)
	}
	fileName := info.Name()
	if fileName == "" || fileName == "." {
		fileName = filePath
	}

	// Check if file already exists
	exists, err := s.metaSvc.FileExists(fileName)
	if err != nil {
		return err
	}
	if exists {
		return errorx.WrapWithDetails(errorx.ErrFileAlreadyExists, fileName)
	}

	// Add file entry to metadata
	if err := s.metaSvc.AddFile(fileName, uint64(info.Size())); err != nil {
		return err
	}

	// Rollback function
	rollback := func(uploadedChunks []string) {
		for _, chunkName := range uploadedChunks {
			chunkMetaData, found := s.metaSvc.GetChunk(chunkName)
			if !found {
				log.Error("%v: %s, skipping deletion", errorx.ErrChunkNotFound, chunkName)
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
		maxParallel    = config.GetConfig().Parallel.Upload
		sem            = make(chan struct{}, maxParallel)
	)

	for chunk := range chunks {
		if chunk.Err != nil {
			errOnce.Do(func() { uploadErr = chunk.Err })
			break
		}
		c := string(chunk.Checksum[:])
		if exists, _ := s.metaSvc.ChunkExists(chunk.Name); exists {
			errOnce.Do(func() {
				uploadErr = errorx.WrapWithDetails(errorx.ErrChunkAlreadyExists, chunk.Name)
			})
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(chunk chunker.Chunk) {
			defer wg.Done()
			defer func() { <-sem }()
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
			log.Info("Retrieving chunk: %s", meta.ChunkName)
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
			continue
		}
		_, err := w.Write(chunkData)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteFile deletes all chunks for a file and removes metadata
func (s *StorageService) DeleteFile(fileName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	metas, err := s.metaSvc.ListChunks(fileName)
	if err != nil {
		return err
	}

	var (
		deleteErrs  []error
		wg          sync.WaitGroup
		maxParallel = config.GetConfig().Parallel.Upload
		sem         = make(chan struct{}, maxParallel)
		mu          sync.Mutex
	)
	for _, meta := range metas {
		sem <- struct{}{}
		wg.Add(1)
		go func(meta metadata.ChunkMetadata) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := s.manager.DeleteChunk(meta.Storage, meta.ChunkName); err != nil {
				mu.Lock()
				deleteErrs = append(deleteErrs, errorx.WrapWithDetails(errorx.ErrChunkDeleteFailed, meta.ChunkName))
				mu.Unlock()
			}
		}(meta)
	}
	wg.Wait()

	if err := s.metaSvc.DeleteFile(fileName); err != nil {
		deleteErrs = append(deleteErrs, errorx.Wrap(errorx.ErrFileDeleteFailed, err))
	}

	if len(deleteErrs) > 0 {
		return errorx.WrapWithDetails(errorx.ErrFileDeleteFailed, fmt.Sprintf("file: %s, errors: %v", fileName, deleteErrs))
	}
	return nil
}
