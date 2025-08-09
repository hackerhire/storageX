package app

import (
	"fmt"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
	"github.com/sayuyere/storageX/internal/config"
	"github.com/sayuyere/storageX/internal/log"
	"github.com/sayuyere/storageX/internal/manager"
	"github.com/sayuyere/storageX/internal/metadata"
	"github.com/sayuyere/storageX/internal/storage"
)

// ServiceBundle aggregates all core services for the CLI
// Add more fields as you add more services
// This struct can be passed to CLI command handlers

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
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	log.InitLogger(cfg.Log.Debug)
	ch := chunker.GetChunkerFromConfig()
	meta, err := metadata.NewMetadataServiceFromConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to init metadata: %w", err)
	}
	// Setup cloud providers (currently only Dropbox, extend as needed)
	var cloudSvcs []cloud.CloudStorage
	authConfigs := cloud.AuthConfigFromCloudConfig(&cfg.Cloud) // Initialize auth config
	// This is where you would initialize cloud storage with auth configs
	// Example for Dropbox, extend as needed for other providers
	for _, auth := range authConfigs {
		if auth.DropboxAccessToken != "" {
			cloudSvcs = append(cloudSvcs, cloud.NewDropboxStorageWithAuth(auth))
		}
	}
	// If no cloud services configured, return an error
	if len(cloudSvcs) == 0 {
		return nil, fmt.Errorf("no cloud storage configured in %s", configPath)
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
