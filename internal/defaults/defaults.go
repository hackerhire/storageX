package defaults

const (
	DefaultChunkSize              = 1024 * 1024 // 1MB
	DefaultConfigPath             = "config/config.json"
	DefaultDBPath                 = "metadata.db"
	DefaultLogDebug               = false
	DefaultStorageUploadWorkers   = 4 // Default number of upload workers
	DefaultStorageDownloadWorkers = 4 // Default number of download workers
)
