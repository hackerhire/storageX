package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

type AppConfig struct {
	ChunkSize int                   `json:"chunk_size"`
	Cloud     CloudConfig           `json:"cloud"`
	Log       LogConfig             `json:"log"`
	Meta      MetaDataServiceConfig `json:"metadata"`
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
			log.Default().Printf("Warning: Dropbox access token at index %d is empty", i)
		} else if os.Getenv(cfg.Cloud.DropboxAccessTokens[i]) == "" {
			log.Default().Printf("Warning: Environment variable for Dropbox access token at index %d is not set", i)
		} else {
			log.Default().Printf("Dropbox access token at index %d is set", i)
			cfg.Cloud.DropboxAccessTokens[i] = os.Getenv(cfg.Cloud.DropboxAccessTokens[i])
		}
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
		config = cfg
	})
	return config, err
}

// GetConfig returns the loaded config (must call LoadConfig first).
func GetConfig() *AppConfig {
	if config == nil {
		LoadConfig(defaults.DefaultConfigPath)
		fmt.Println("App config not loaded. Using default config.")
	}
	return config
}
