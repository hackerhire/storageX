package app

import (
	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
	"github.com/sayuyere/storageX/internal/config"
	errorx "github.com/sayuyere/storageX/internal/errors" // unified error constants
	"github.com/sayuyere/storageX/internal/log"
	"github.com/sayuyere/storageX/internal/manager"
	"github.com/sayuyere/storageX/internal/metadata"
	"github.com/sayuyere/storageX/internal/storage"
)

// ServiceBundle aggregates all core services for the CLI
type ServiceBundle struct {
	Config   *config.AppConfig
	Chunker  *chunker.FileChunker
	Metadata *metadata.MetadataService
	Manager  *manager.StorageManager
	Storage  *storage.StorageService
}

// NewServiceBundle initializes all services and returns a bundle
func NewServiceBundle(configPath string) (*ServiceBundle, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrConfigLoadFailed, err)
	}

	log.InitLogger(cfg.Log.Debug)

	ch := chunker.GetChunkerFromConfig()

	meta, err := metadata.NewMetadataServiceFromConfig()
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrMetadataInitFailed, err)
	}

	// Setup cloud providers
	var cloudSvcs []cloud.CloudStorage
	authConfigs := cloud.AuthConfigFromCloudConfig(&cfg.Cloud)

	for _, auth := range authConfigs {
		if auth.DropboxAccessToken != "" {
			cloudSvcs = append(cloudSvcs, cloud.NewDropboxStorageWithAuth(auth))
		}
	}

	if len(cloudSvcs) == 0 {
		return nil, errorx.WrapWithDetails(errorx.ErrNoCloudStorageConfigured, configPath)
	}

	mgr := manager.NewStorageManager(cloudSvcs)
	for _, svc := range cloudSvcs {
		mgr.AddCloudStorage(svc)
	}

	stor := storage.NewStorageService(mgr, meta, ch)

	return &ServiceBundle{
		Config:   cfg,
		Chunker:  ch,
		Metadata: meta,
		Manager:  mgr,
		Storage:  stor,
	}, nil
}
