package errors

import (
	"errors"
	"fmt"
)

var (
	ErrDropboxUpload   = errors.New("dropbox: upload failed")
	ErrDropboxDownload = errors.New("dropbox: download failed")
	ErrDropboxDelete   = errors.New("dropbox: delete failed")

	ErrDriveUpload     = errors.New("gdrive: upload failed")
	ErrDriveDownload   = errors.New("gdrive: download failed")
	ErrDriveDelete     = errors.New("gdrive: delete failed")
	ErrStorageNotFound = errors.New("storage: storage system not found")

	// Add more unified errors for other providers as needed
)

var (
	ErrFileAlreadyExists   = errors.New("file already exists in metadata")
	ErrChunkAlreadyExists  = errors.New("chunk already exists in metadata")
	ErrFileInfoFetchFailed = errors.New("failed to get file info")
	ErrFileNotFound        = errors.New("file not found in metadata")
	ErrChunkNotFound       = errors.New("chunk not found in metadata")
	ErrFileDeleteFailed    = errors.New("failed to delete file metadata")
	ErrChunkDeleteFailed   = errors.New("failed to delete chunk metadata")
)

// Metadata-related errors
var (
	ErrMetadataDBOpenFailed     = errors.New("metadata: failed to open database")
	ErrMetadataSchemaInitFailed = errors.New("metadata: failed to initialize schema")
	ErrChunkInsertFailed        = errors.New("metadata: failed to insert chunk")
	ErrFileInsertFailed         = errors.New("metadata: failed to insert file")
	ErrFileUpdateFailed         = errors.New("metadata: failed to update file")
	ErrDBQueryFailed            = errors.New("metadata: database query failed")
	ErrDBScanFailed             = errors.New("metadata: failed to scan database rows")
)

// Chunker errors
var (
	ErrConfigNotLoaded = errors.New("chunker: app config not loaded")
	ErrChunkReadFailed = errors.New("chunker: failed to read chunk from file")
)

// App / Service initialization errors
var (
	ErrConfigLoadFailed         = errors.New("app: failed to load config")
	ErrMetadataInitFailed       = errors.New("app: failed to initialize metadata service")
	ErrNoCloudStorageConfigured = errors.New("app: no cloud storage configured")
)

// Wrappers to add context
func Wrap(base error, err error) error {
	if err == nil {
		return base
	}
	return fmt.Errorf("%w: %v", base, err)
}

func WrapWithDetails(base error, details string) error {
	return fmt.Errorf("%w: %s", base, details)
}

func WrapDropboxError(base error, err error) error {
	return fmt.Errorf("%w: %v", base, err)
}

func WrapDriveError(base error, err error) error {
	return fmt.Errorf("%w: %v", base, err)
}
