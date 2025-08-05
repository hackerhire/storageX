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

	t.Log("[e2e] Using Dropbox token of length:", len(dropboxToken))
	auth := cloud.AuthConfig{DropboxAccessToken: dropboxToken}
	t.Log("[e2e] Initializing Dropbox storage...")
	dropbox := cloud.NewDropboxStorageWithAuth(auth)
	mgr := manager.NewStorageManager([]cloud.CloudStorage{dropbox})
	t.Log("[e2e] Initializing metadata service...")
	meta, err := metadata.NewMetadataService("e2e_test.db")
	if err != nil {
		t.Fatalf("failed to create metadata: %v", err)
	}
	defer func() {
		os.Remove("e2e_test.db")
		t.Log("[e2e] Cleaned up metadata DB")
	}()
	ch := chunker.NewFileChunker(128 * 1024)
	svc := storage.NewStorageService(mgr, meta, ch)

	// Create a temp file to upload
	t.Log("[e2e] Creating temp file for upload...")
	f, err := os.CreateTemp("", "e2e-upload-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		os.Remove(f.Name())
		t.Logf("[e2e] Cleaned up temp file: %s", f.Name())
	}()
	data := []byte("end-to-end test data for storageX")
	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek: %v", err)
	}
	t.Logf("[e2e] Temp file created: %s (%d bytes)", f.Name(), len(data))

	// Upload
	t.Log("[e2e] Uploading file...")
	if err := svc.UploadFile(f.Name()); err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
	t.Log("[e2e] Upload successful.")

	// Download
	var buf bytes.Buffer
	t.Log("[e2e] Downloading file...")
	if err := svc.GetFile(f.Name(), &buf); err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}
	t.Logf("[e2e] Downloaded %d bytes.", buf.Len())
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("Downloaded data mismatch: got %q, want %q", buf.Bytes(), data)
	} else {
		t.Log("[e2e] Downloaded data matches original.")
	}

	// Delete
	t.Log("[e2e] Deleting file and metadata...")
	if err := svc.DeleteFile(f.Name()); err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
	if ok, _ := meta.FileExists(f.Name()); ok {
		t.Error("file metadata still exists after delete")
	} else {
		t.Log("[e2e] File and metadata deleted successfully.")
	}
}
