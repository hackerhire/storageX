package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sayuyere/storageX/internal/defaults"
)

type LogConfig struct {
	Debug bool `json:"debug"`
}

type CloudConfig struct {
	DropboxAccessTokens []string `json:"dropbox_access_tokens,omitempty"`
	// Add other provider configs here
}

type MetaDataServiceConfig struct {
	DBPath string `json:"db_path"`
}

type ParallelConfig struct {
	Upload   int `json:"upload_workers"`
	Download int `json:"download_workers"`
}

type AppConfig struct {
	ChunkSize int                   `json:"chunk_size"`
	Cloud     CloudConfig           `json:"cloud"`
	Log       LogConfig             `json:"log"`
	Meta      MetaDataServiceConfig `json:"metadata"`
	Parallel  ParallelConfig        `json:"parallel"`
}

var (
	config     *AppConfig
	configOnce sync.Once
)

func ResetConfigSingleton() {
	// This is not thread-safe, but fine for test use
	configOnce = sync.Once{}
	config = nil // Reset the config to nil
}

func LookupSecrets(cfg *AppConfig) {
	for i := 0; i < len(cfg.Cloud.DropboxAccessTokens); i++ {
		if cfg.Cloud.DropboxAccessTokens[i] == "" {
			log.Default().Printf("[StorageX] Warning: Dropbox access token at index %d is empty", i)
		} else if os.Getenv(cfg.Cloud.DropboxAccessTokens[i]) == "" {
			log.Default().Printf("[StorageX] Warning: Environment variable for Dropbox access token at index %d is not set", i)
		} else {
			log.Default().Printf("[StorageX] Dropbox access token at index %d is set", i)
			cfg.Cloud.DropboxAccessTokens[i] = os.Getenv(cfg.Cloud.DropboxAccessTokens[i])
		}
	}
}
func UpdatePaths(cfg *AppConfig) {
	if cfg.Meta.DBPath == "" {
		cfg.Meta.DBPath = defaults.DefaultDBPath
	}
	// Expand ~ to home directory if present
	if len(cfg.Meta.DBPath) > 0 && cfg.Meta.DBPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.Meta.DBPath = filepath.Join(home, cfg.Meta.DBPath[1:])
		}
	}
	// Convert DBPath to absolute path if it's relative
	if !filepath.IsAbs(cfg.Meta.DBPath) {
		absPath, err := filepath.Abs(cfg.Meta.DBPath)
		log.Default().Printf("[StorageX] DBPath updated to absolute path: %s", absPath)
		if err == nil {
			cfg.Meta.DBPath = absPath
		}
	}
	if cfg.Parallel.Upload <= 0 {
		cfg.Parallel.Upload = defaults.DefaultStorageUploadWorkers // default upload workers
	}
	if cfg.Parallel.Download <= 0 {
		cfg.Parallel.Download = defaults.DefaultStorageDownloadWorkers // default download workers
	}
}

// LoadConfig loads configuration from the given JSON file path.
func LoadConfig(path string) (*AppConfig, error) {
	var err error

	configOnce.Do(func() {
		defaultConfig := &AppConfig{
			ChunkSize: defaults.DefaultChunkSize,
			Cloud: CloudConfig{
				DropboxAccessTokens: []string{},
			},
			Log: LogConfig{
				Debug: defaults.DefaultLogDebug,
			},
			Meta: MetaDataServiceConfig{
				DBPath: defaults.DefaultDBPath,
			},
			Parallel: ParallelConfig{
				Upload:   4,
				Download: 4,
			},
		}
		f, e := os.Open(path)
		if e != nil {
			err = e
			config = defaultConfig // Use default config if file not found
			log.Printf("Failed to open config file %s: %v\n", path, e)
			return
		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		cfg := &AppConfig{}
		if e := decoder.Decode(cfg); e != nil {
			err = e
			config = defaultConfig // Use default config if decoding fails
			log.Printf("Failed to decode config from %s: %v\n", path, e)
			return
		}
		LookupSecrets(cfg)
		UpdatePaths(cfg)
		config = cfg
	})
	return config, err
}

// GetConfig returns the loaded config (must call LoadConfig first).
func GetConfig() *AppConfig {
	if config == nil {
		LoadConfig(defaults.DefaultConfigPath)
		log.Default().Printf("[StorageX] App config not loaded. Using default config.")
	}
	return config
}
