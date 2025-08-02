package cloud_test

import (
	"os"
	"testing"

	"github.com/sayuyere/storageX/internal/cloud"
)

func TestDropboxStorageLifecycle(t *testing.T) {
	auth := cloud.AuthConfig{
		DropboxAccessToken: os.Getenv("DROPBOX_ACCESS_TOKEN"),
	}
	if auth.DropboxAccessToken == "" {
		t.Fatal("DROPBOX_ACCESS_TOKEN environment variable not set")
	}
	dropbox := cloud.NewDropboxStorageWithAuth(auth)
	chunkName := "test-chunk.txt"
	chunkData := []byte("hello, dropbox!")

	// Upload
	err := dropbox.UploadChunk(chunkName, chunkData)
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}

	// Get
	data, err := dropbox.GetChunk(chunkName)
	if err != nil {
		t.Fatalf("GetChunk failed: %v", err)
	}
	if string(data) != string(chunkData) {
		t.Errorf("GetChunk data mismatch: got %q, want %q", data, chunkData)
	}

	// Delete
	err = dropbox.DeleteChunk(chunkName)
	if err != nil {
		t.Fatalf("DeleteChunk failed: %v", err)
	}
}

func TestDropboxStorageUnitSize(t *testing.T) {
	auth := cloud.AuthConfig{
		DropboxAccessToken: os.Getenv("DROPBOX_ACCESS_TOKEN"),
	}
	if auth.DropboxAccessToken == "" {
		t.Fatal("DROPBOX_ACCESS_TOKEN environment variable not set")
	}
	dropbox := cloud.NewDropboxStorageWithAuth(auth)
	size, err := dropbox.GetRemainingSize()
	if err != nil {
		t.Fatalf("GetRemainingSize failed: %v", err)
	}
	if size <= 0 {
		t.Errorf("Expected positive storage unit size, got %d", size)
	}
}

func TestDropboxStorageSystemID(t *testing.T) {
	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if token == "" {
		t.Skip("DROPBOX_ACCESS_TOKEN not set")
	}
	auth := cloud.AuthConfig{
		DropboxAccessToken: token,
	}
	dropbox := cloud.NewDropboxStorageWithAuth(auth)
	id := dropbox.StorageSystemID()
	if id == "dropbox:unknown" || id == "" {
		t.Errorf("Expected unique Dropbox storage system ID, got %q", id)
	}
	if len(id) < 10 {
		t.Errorf("StorageSystemID too short: %q", id)
	}
	if id[:8] != "dropbox:" {
		t.Errorf("StorageSystemID should start with 'dropbox:', got %q", id)
	}
}
