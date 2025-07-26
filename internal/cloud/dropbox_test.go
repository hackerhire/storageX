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
