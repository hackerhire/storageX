package config

import (
	"encoding/json"
	"fmt"
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
			return
		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		cfg := &AppConfig{}
		if e := decoder.Decode(cfg); e != nil {
			err = e
			config = defaultConfig // Use default config if decoding fails
			return
		}
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
