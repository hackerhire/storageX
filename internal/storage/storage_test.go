package storage

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
	"github.com/sayuyere/storageX/internal/manager"
	"github.com/sayuyere/storageX/internal/metadata"
)

const DefaultChunkSize = 64 // Default chunk size for testing

type mockCloudStorage struct {
	chunks     map[string][]byte
	failUpload bool
	failGet    bool
	failDelete bool
}

func (m *mockCloudStorage) UploadChunk(name string, data []byte) error {
	if m.failUpload {
		return errors.New("upload failed")
	}
	m.chunks[name] = data
	return nil
}
func (m *mockCloudStorage) GetChunk(name string) ([]byte, error) {
	if m.failGet {
		return nil, errors.New("get failed")
	}
	data, ok := m.chunks[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return data, nil
}
func (m *mockCloudStorage) DeleteChunk(name string) error {
	if m.failDelete {
		return errors.New("delete failed")
	}
	delete(m.chunks, name)
	return nil
}
func (m *mockCloudStorage) GetRemainingSize() (int64, error) { return 1 << 30, nil }
func (m *mockCloudStorage) StorageSystemID() string          { return "mock" }

func setupStorageService(t *testing.T) (*StorageService, *mockCloudStorage, *metadata.MetadataService, func()) {
	metaSvc, err := metadata.NewMetadataService("test_storage.db")
	if err != nil {
		t.Fatalf("failed to create metadata: %v", err)
	}
	mockCloud := &mockCloudStorage{chunks: make(map[string][]byte)}
	mgr := manager.NewStorageManager([]cloud.CloudStorage{mockCloud})
	ch := chunker.NewFileChunker(DefaultChunkSize)
	cleanup := func() { os.Remove("test_storage.db") }
	return NewStorageService(mgr, metaSvc, ch), mockCloud, metaSvc, cleanup
}

func TestUploadAndGetFile(t *testing.T) {
	ss, _, _, cleanup := setupStorageService(t)
	defer cleanup()

	f, err := os.CreateTemp("", "storage-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	data := []byte("testdata1234567890")

	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek: %v", err)
	}

	if err := ss.UploadFile(f.Name()); err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	var buf bytes.Buffer
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}
	if err := ss.GetFile(info.Name(), &buf); err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("GetFile data mismatch: got %v, want %v", buf.Bytes(), data)
	}
}

func TestDeleteFile(t *testing.T) {
	ss, _, metaSvc, cleanup := setupStorageService(t)
	defer cleanup()

	f, err := os.CreateTemp("", "storage-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	data := []byte("testdata1234567890")
	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek: %v", err)
	}

	if err := ss.UploadFile(f.Name()); err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
	if err := ss.DeleteFile(f.Name()); err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
	if ok, _ := metaSvc.FileExists(f.Name()); ok {
		t.Error("file metadata still exists after delete")
	}
}

func TestUploadFile_RollbackOnChunkError(t *testing.T) {
	ss, mockCloud, _, cleanup := setupStorageService(t)
	defer cleanup()
	mockCloud.failUpload = true

	f, err := os.CreateTemp("", "storage-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	defer os.Remove(f.Name())
	if _, err := f.Write([]byte("testdata")); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek: %v", err)
	}
	if err := ss.UploadFile(f.Name()); err == nil {
		t.Error("expected error on upload, got nil")
	}
}
