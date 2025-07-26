package cloud

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type DropboxStorage struct {
	client files.Client
}

func NewDropboxStorageWithAuth(auth AuthConfig) *DropboxStorage {
	config := dropbox.Config{
		Token:    auth.DropboxAccessToken,
		LogLevel: dropbox.LogInfo, // or dropbox.LogOff
	}
	client := files.New(config)
	return &DropboxStorage{client: client}
}

func (d *DropboxStorage) UploadChunk(name string, data []byte) error {
	uploadArg := files.NewUploadArg("/" + name)
	uploadArg.Mode.Tag = "overwrite"
	_, err := d.client.Upload(uploadArg, ioutil.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("Dropbox upload failed: %w", err)
	}
	return nil
}

func (d *DropboxStorage) GetChunk(name string) ([]byte, error) {
	downloadArg := files.NewDownloadArg("/" + name)
	_, content, err := d.client.Download(downloadArg)
	if err != nil {
		return nil, fmt.Errorf("Dropbox download failed: %w", err)
	}
	defer content.Close()
	return ioutil.ReadAll(content)
}

func (d *DropboxStorage) DeleteChunk(name string) error {
	deleteArg := files.NewDeleteArg("/" + name)
	_, err := d.client.DeleteV2(deleteArg)
	if err != nil {
		return fmt.Errorf("Dropbox delete failed: %w", err)
	}
	return nil
}
