package cloud

import (
	"errors"
	"fmt"
)

var (
	ErrDropboxUpload   = errors.New("dropbox: upload failed")
	ErrDropboxDownload = errors.New("dropbox: download failed")
	ErrDropboxDelete   = errors.New("dropbox: delete failed")

	ErrDriveUpload   = errors.New("gdrive: upload failed")
	ErrDriveDownload = errors.New("gdrive: download failed")
	ErrDriveDelete   = errors.New("gdrive: delete failed")

	// Add more unified errors for other providers as needed
)

func WrapDropboxError(base error, err error) error {
	return fmt.Errorf("%w: %v", base, err)
}

func WrapDriveError(base error, err error) error {
	return fmt.Errorf("%w: %v", base, err)
}
