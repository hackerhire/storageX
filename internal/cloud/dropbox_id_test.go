package cloud

import (
	"os"
	"testing"
)

func TestDropboxStorageSystemID(t *testing.T) {
	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	if token == "" {
		t.Skip("DROPBOX_ACCESS_TOKEN not set")
	}
	dropbox := NewDropboxStorageWithAuth(AuthConfig{DropboxAccessToken: token})
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
