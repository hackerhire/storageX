package e2e

import (
	"bytes"
	"os"
	"testing"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
	"github.com/sayuyere/storageX/internal/manager"
	"github.com/sayuyere/storageX/internal/metadata"
	"github.com/sayuyere/storageX/internal/storage"
)

// NOTE: This test requires a real Dropbox token in the environment and a working config.
func TestEndToEnd_UploadDownloadDelete(t *testing.T) {
	dropboxToken := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if dropboxToken == "" {
		t.Skip("DROPBOX_ACCESS_TOKEN not set; skipping e2e test")
	}

	auth := cloud.AuthConfig{DropboxAccessToken: dropboxToken}
	dropbox := cloud.NewDropboxStorageWithAuth(auth)
	mgr := manager.NewStorageManager([]cloud.CloudStorage{dropbox})
	meta, err := metadata.NewMetadataService("e2e_test.db")
	if err != nil {
		t.Fatalf("failed to create metadata: %v", err)
	}
	defer os.Remove("e2e_test.db")
	ch := chunker.NewFileChunker(128 * 1024)
	svc := storage.NewStorageService(mgr, meta, ch)

	// Create a temp file to upload
	f, err := os.CreateTemp("", "e2e-upload-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	data := []byte("end-to-end test data for storageX")
	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek: %v", err)
	}

	// Upload
	if err := svc.UploadFile(f.Name()); err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Download
	var buf bytes.Buffer
	if err := svc.GetFile(f.Name(), &buf); err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("Downloaded data mismatch: got %q, want %q", buf.Bytes(), data)
	}

	// Delete
	if err := svc.DeleteFile(f.Name()); err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
	if ok, _ := meta.FileExists(f.Name()); ok {
		t.Error("file metadata still exists after delete")
	}
}
