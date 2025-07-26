package config

import (
	"encoding/json"
	"os"
	"sync"
)

type AppConfig struct {
	CloudProviders []string `json:"cloud_providers"`
	ChunkSize      int      `json:"chunk_size"`
	InputFile      string   `json:"input_file"`
}

var (
	config     *AppConfig
	configOnce sync.Once
)

// LoadConfig loads configuration from the given JSON file path.
func LoadConfig(path string) (*AppConfig, error) {
	var err error
	configOnce.Do(func() {
		f, e := os.Open(path)
		if e != nil {
			err = e
			return
		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		cfg := &AppConfig{}
		if e := decoder.Decode(cfg); e != nil {
			err = e
			return
		}
		config = cfg
	})
	return config, err
}

// GetConfig returns the loaded config (must call LoadConfig first).
func GetConfig() *AppConfig {
	return config
}
