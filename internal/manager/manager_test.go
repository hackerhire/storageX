package manager_test

import (
	"errors"
	"testing"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
	"github.com/sayuyere/storageX/internal/manager"
)

type mockCloudStorage struct {
	id      string
	chunks  map[string][]byte
	failOps map[string]bool
}

func (m *mockCloudStorage) UploadChunk(name string, data []byte) error {
	if m.failOps["upload"] {
		return errors.New("upload failed")
	}
	m.chunks[name] = data
	return nil
}
func (m *mockCloudStorage) GetChunk(name string) ([]byte, error) {
	if m.failOps["get"] {
		return nil, errors.New("get failed")
	}
	data, ok := m.chunks[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return data, nil
}
func (m *mockCloudStorage) DeleteChunk(name string) error {
	if m.failOps["delete"] {
		return errors.New("delete failed")
	}
	delete(m.chunks, name)
	return nil
}
func (m *mockCloudStorage) GetRemainingSize() (int64, error) { return 0, nil }
func (m *mockCloudStorage) StorageSystemID() string          { return m.id }

func newMockCloudStorage(id string) *mockCloudStorage {
	return &mockCloudStorage{
		id:      id,
		chunks:  make(map[string][]byte),
		failOps: make(map[string]bool),
	}
}

func TestManager_UploadGetDeleteChunk(t *testing.T) {
	mock := newMockCloudStorage("mock1")
	mgr := manager.NewStorageManager([]cloud.CloudStorage{mock})
	chunk := chunker.Chunk{Data: []byte("hello")}

	// Upload
	storage, err := mgr.UploadChunk("chunk1", chunk)
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}
	if storage.StorageSystemID() != "mock1" {
		t.Errorf("expected storage id 'mock1', got %q", storage.StorageSystemID())
	}

	// Get
	data, err := mgr.GetChunk("mock1", "chunk1")
	if err != nil {
		t.Fatalf("GetChunk failed: %v", err)
	}
	if string(data) != string(chunk.Bytes()) {
		t.Errorf("expected chunk data 'hello', got %q", data)
	}

	// Delete
	err = mgr.DeleteChunk("mock1", "chunk1")
	if err != nil {
		t.Fatalf("DeleteChunk failed: %v", err)
	}
	_, err = mgr.GetChunk("mock1", "chunk1")
	if err == nil {
		t.Errorf("expected error for deleted chunk, got nil")
	}
}

func TestManager_SearchStorageID(t *testing.T) {
	mock1 := newMockCloudStorage("id1")
	mock2 := newMockCloudStorage("id2")
	mgr := manager.NewStorageManager([]cloud.CloudStorage{mock1, mock2})
	if mgr.SearchStorageID("id2") != mock2 {
		t.Errorf("SearchStorageID did not return correct storage")
	}
	if mgr.SearchStorageID("notfound") != nil {
		t.Errorf("SearchStorageID should return nil for missing id")
	}
}
