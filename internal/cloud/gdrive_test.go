package cloud_test

import (
	"os"
	"testing"

	"github.com/yourusername/storageX/internal/log"

	"github.com/yourusername/storageX/internal/cloud"
)

func TestDriveStorageLifecycle(t *testing.T) {
	auth := cloud.AuthConfig{
		GoogleCredentialsFile: os.Getenv("GOOGLE_CREDENTIALS_FILE"),
	}
	folderID := os.Getenv("GDRIVE_FOLDER_ID") // Set this env var to your shared folder ID
	if folderID == "" {
		t.Fatal("GDRIVE_FOLDER_ID environment variable not set")
	}

	log.InitLogger(true) // Initialize logger for testing
	log.Info("Starting DriveStorage lifecycle test")
	log.Info("Auth config:", auth, os.Getenv("GOOGLE_CREDENTIALS_FILE"))

	drive := cloud.NewDriveStorageWithAuthAndFolder(auth, folderID)
	chunkName := "test-chunk.txt"
	chunkData := []byte("hello, world!")

	// Upload
	err := drive.UploadChunk(chunkName, chunkData)
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}

	// Get
	data, err := drive.GetChunk(chunkName)
	if err != nil {
		t.Fatalf("GetChunk failed: %v", err)
	}
	if string(data) != string(chunkData) {
		t.Errorf("GetChunk data mismatch: got %q, want %q", data, chunkData)
	}

	// Delete
	err = drive.DeleteChunk(chunkName)
	if err != nil {
		t.Fatalf("DeleteChunk failed: %v", err)
	}
}
