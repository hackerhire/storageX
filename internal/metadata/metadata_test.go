package metadata_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/sayuyere/storageX/internal/metadata"
)

func setupTestDB(t *testing.T) *metadata.MetadataService {
	dbPath := "test_metadata.db"
	_ = os.Remove(dbPath)
	metaSvc, err := metadata.NewMetadataService(dbPath)
	if err != nil {
		t.Fatalf("failed to create metadata service: %v", err)
	}
	t.Cleanup(func() { os.Remove(dbPath) })
	return metaSvc
}

func TestAddAndGetFile(t *testing.T) {
	metaSvc := setupTestDB(t)
	fileName := "testfile.txt"
	f, err := os.Create(fileName) // Ensure file exists for testing
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()
	defer os.Remove(fileName) // Clean up test file after test
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}
	if err := metaSvc.AddFile(fileName, uint64(info.Size())); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}
	meta, ok := metaSvc.GetFile(fileName)
	if !ok {
		t.Fatalf("GetFile failed: not found")
	}
	if meta.FileName != fileName {
		t.Errorf("expected %q, got %q", fileName, meta.FileName)
	}
}

func TestAddChunkAndExists(t *testing.T) {
	metaSvc := setupTestDB(t)
	fileName := "testfile.txt"
	f, err := os.Create(fileName) // Ensure file exists for testing
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()
	defer os.Remove(fileName) // Clean up test file after test
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}
	_ = metaSvc.AddFile(fileName, uint64(info.Size()))
	chunk := metadata.ChunkMetadata{
		ChunkName: "chunk1",
		Size:      123,
		Checksum:  "abc",
		Index:     0,
		FileName:  fileName,
		Storage:   "mock",
	}
	if err := metaSvc.AddChunk(fileName, chunk); err != nil {
		t.Fatalf("AddChunk failed: %v", err)
	}
	ok, err := metaSvc.ChunkExists("chunk1")
	if err != nil || !ok {
		t.Fatalf("ChunkExists failed: %v", err)
	}
	meta, ok := metaSvc.GetChunk("chunk1")
	if !ok || meta.ChunkName != "chunk1" {
		t.Errorf("GetChunk failed: got %+v", meta)
	}
}

func TestListChunksAndFiles(t *testing.T) {
	metaSvc := setupTestDB(t)
	fileName := "testfile.txt"
	f, err := os.Create(fileName) // Ensure file exists for testing
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()
	defer os.Remove(fileName) // Clean up test file after test
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}
	_ = metaSvc.AddFile(fileName, uint64(info.Size()))
	for i := 0; i < 3; i++ {
		chunk := metadata.ChunkMetadata{
			ChunkName: "chunk" + string("A"+strconv.Itoa(i)),
			Size:      int64(i * 10),
			Checksum:  "sum",
			Index:     i,
			FileName:  fileName,
			Storage:   "mock",
		}
		_ = metaSvc.AddChunk(fileName, chunk)
	}
	chunks, err := metaSvc.ListChunks(fileName)
	if err != nil || len(chunks) != 3 {
		t.Fatalf("ListChunks failed: %v, got %d", err, len(chunks))
	}
	files, err := metaSvc.ListFiles()
	if err != nil || len(files) != 1 {
		t.Fatalf("ListFiles failed: %v, got %d", err, len(files))
	}
}

func TestDeleteChunkAndFile(t *testing.T) {
	metaSvc := setupTestDB(t)
	fileName := "testfile.txt"
	f, err := os.Create(fileName) // Ensure file exists for testing
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()
	defer os.Remove(fileName) // Clean up test file after test
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}

	_ = metaSvc.AddFile(fileName, uint64(info.Size()))
	chunk := metadata.ChunkMetadata{
		ChunkName: "chunk1",
		Size:      123,
		Checksum:  "abc",
		Index:     0,
		FileName:  fileName,
		Storage:   "mock",
	}
	_ = metaSvc.AddChunk(fileName, chunk)
	if err := metaSvc.DeleteChunk("chunk1"); err != nil {
		t.Errorf("DeleteChunk failed: %v", err)
	}
	ok, _ := metaSvc.ChunkExists("chunk1")
	if ok {
		t.Errorf("Chunk should not exist after delete")
	}
	_ = metaSvc.AddChunk(fileName, chunk)
	if err := metaSvc.DeleteFile(fileName); err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}
	ok, _ = metaSvc.FileExists(fileName)
	if ok {
		t.Errorf("File should not exist after delete")
	}
}
