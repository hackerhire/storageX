package cloud

import (
	"bytes"
	"io/ioutil"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"

	errorsx "github.com/sayuyere/storageX/internal/errors"
	"github.com/sayuyere/storageX/internal/log"
)

type DropboxStorage struct {
	client files.Client
	config dropbox.Config
}

func NewDropboxStorageWithAuth(auth AuthConfig) *DropboxStorage {
	config := dropbox.Config{
		Token:    auth.DropboxAccessToken,
		LogLevel: dropbox.LogInfo, // or dropbox.LogOff
	}
	client := files.New(config)
	return &DropboxStorage{client: client, config: config}
}

func (d *DropboxStorage) UploadChunk(name string, data []byte) error {
	uploadArg := files.NewUploadArg("/" + name)
	uploadArg.Mode.Tag = "overwrite"
	_, err := d.client.Upload(uploadArg, ioutil.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return errorsx.WrapDropboxError(errorsx.ErrDropboxUpload, err)
	}
	return nil
}

func (d *DropboxStorage) GetChunk(name string) ([]byte, error) {
	downloadArg := files.NewDownloadArg("/" + name)
	_, content, err := d.client.Download(downloadArg)
	if err != nil {
		return nil, errorsx.WrapDropboxError(errorsx.ErrDropboxDownload, err)
	}
	defer content.Close()
	return ioutil.ReadAll(content)
}

func (d *DropboxStorage) DeleteChunk(name string) error {
	deleteArg := files.NewDeleteArg("/" + name)
	_, err := d.client.DeleteV2(deleteArg)
	if err != nil {
		return errorsx.WrapDropboxError(errorsx.ErrDropboxDelete, err)
	}
	return nil
}

func (d *DropboxStorage) GetRemainingSize() (uint64, error) {
	userClient := users.New(d.config)
	spaceUsage, err := userClient.GetSpaceUsage()
	if err != nil {
		return 0, errorsx.WrapDropboxError(errorsx.ErrDropboxDownload, err)
	}
	var allocated, used uint64
	if spaceUsage.Allocation.Tag == "individual" && spaceUsage.Allocation.Individual != nil {
		allocated = spaceUsage.Allocation.Individual.Allocated
	} else if spaceUsage.Allocation.Tag == "team" && spaceUsage.Allocation.Team != nil {
		allocated = spaceUsage.Allocation.Team.Allocated
	} else {
		return 0, errorsx.WrapDropboxError(errorsx.ErrDropboxDownload, err)
	}
	used = spaceUsage.Used
	if allocated < used {
		return 0, nil // Guard against underflow
	}
	return allocated - used, nil
}

func (d *DropboxStorage) StorageSystemID() string {
	// Use Dropbox account_id as unique tenant/storage node id
	userClient := users.New(d.config)
	acc, err := userClient.GetCurrentAccount()
	log.Info("Dropbox account ID:", acc.AccountId)
	if err == nil && acc.AccountId != "" {
		return "dropbox:" + acc.AccountId
	}
	// fallback to token hash or config
	return "dropbox:unknown"
}
