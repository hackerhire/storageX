package cloud

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sayuyere/storageX/internal/log"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveStorage is a Google Drive implementation

type DriveStorage struct {
	service  *drive.Service
	folderID string // Optional: store in a specific folder
}

func NewDriveStorage() *DriveStorage {
	ctx := context.Background()
	service, err := drive.NewService(ctx, option.WithCredentialsFile("credentials.json"), option.WithScopes(drive.DriveFileScope))
	if err != nil {
		log.Error("Failed to create Drive service: %v", err)
		return &DriveStorage{service: service}
	}
	log.Info("Google Drive service initialized")
	return &DriveStorage{service: service}
}

func NewDriveStorageWithAuth(auth AuthConfig) *DriveStorage {
	ctx := context.Background()
	credsFile := auth.GoogleCredentialsFile
	if credsFile == "" {
		credsFile = "credentials.json"
	}
	service, err := drive.NewService(ctx, option.WithCredentialsFile(credsFile), option.WithScopes(drive.DriveFileScope))
	if err != nil {
		log.Error("Failed to create Drive service: %v", err)
		return &DriveStorage{service: service}
	}
	log.Info("Google Drive service initialized with unified auth")
	return &DriveStorage{service: service}
}

func NewDriveStorageWithAuthAndFolder(auth AuthConfig, folderID string) *DriveStorage {
	ctx := context.Background()
	credsFile := auth.GoogleCredentialsFile
	if credsFile == "" {
		credsFile = "credentials.json"
	}
	service, err := drive.NewService(ctx, option.WithCredentialsFile(credsFile), option.WithScopes(drive.DriveFileScope))
	if err != nil {
		log.Error("Failed to create Drive service: %v", err)
		return &DriveStorage{service: service, folderID: folderID}
	}
	log.Info("Google Drive service initialized with unified auth and folderID=%s", folderID)
	return &DriveStorage{service: service, folderID: folderID}
}

func (d *DriveStorage) UploadChunk(name string, data []byte) error {
	if d.service == nil {
		return fmt.Errorf("drive service not initialized")
	}
	file := &drive.File{Name: name}
	if d.folderID != "" {
		file.Parents = []string{d.folderID}
	}
	tmpfile, err := os.CreateTemp("", "chunk-*")
	if err != nil {
		log.Error("Failed to create temp file: %v", err)
		return err
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(data); err != nil {
		log.Error("Failed to write to temp file: %v", err)
		return err
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Error("Failed to seek temp file: %v", err)
		return err
	}
	_, err = d.service.Files.Create(file).Media(tmpfile).Do()
	if err != nil {
		log.Error("Failed to upload chunk %s: %v", name, err)
		return fmt.Errorf("failed to upload chunk: %w", err)
	}
	log.Info("Uploaded %s to Google Drive (folderID=%s)", name, d.folderID)
	return nil
}

func (d *DriveStorage) GetChunk(name string) ([]byte, error) {
	if d.service == nil {
		return nil, fmt.Errorf("drive service not initialized")
	}
	q := fmt.Sprintf("name='%s' and trashed=false", name)
	files, err := d.service.Files.List().Q(q).Do()
	if err != nil || len(files.Files) == 0 {
		log.Error("File %s not found: %v", name, err)
		return nil, fmt.Errorf("file not found: %w", err)
	}
	fileID := files.Files[0].Id
	resp, err := d.service.Files.Get(fileID).Download()
	if err != nil {
		log.Error("Failed to download chunk %s: %v", name, err)
		return nil, fmt.Errorf("failed to download chunk: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read chunk %s: %v", name, err)
		return nil, err
	}
	log.Info("Downloaded %s from Google Drive", name)
	return data, nil
}

func (d *DriveStorage) DeleteChunk(name string) error {
	if d.service == nil {
		return fmt.Errorf("drive service not initialized")
	}
	q := fmt.Sprintf("name='%s' and trashed=false", name)
	files, err := d.service.Files.List().Q(q).Do()
	if err != nil || len(files.Files) == 0 {
		log.Error("File %s not found for delete: %v", name, err)
		return fmt.Errorf("file not found: %w", err)
	}
	fileID := files.Files[0].Id
	if err := d.service.Files.Delete(fileID).Do(); err != nil {
		log.Error("Failed to delete chunk %s: %v", name, err)
		return fmt.Errorf("failed to delete chunk: %w", err)
	}
	log.Info("Deleted %s from Google Drive", name)
	return nil
}
