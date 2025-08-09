package cloud

import "github.com/sayuyere/storageX/internal/config"

// AuthConfig is a unified authentication configuration for all cloud providers
// Each provider can use the relevant fields

type AuthConfig struct {
	DropboxAccessToken string // Dropbox API access token
	// Add more fields as needed for other providers
}

// AuthConfigFromCloudConfig returns a slice of AuthConfig from a CloudConfig
func AuthConfigFromCloudConfig(cloudCfg *config.CloudConfig) []AuthConfig {
	var result []AuthConfig

	for _, token := range cloudCfg.DropboxAccessTokens {
		result = append(result, AuthConfig{DropboxAccessToken: token})
	}

	return result
}

// LinkAuthConfigForProvider initializes AuthConfig for a given provider using the cloud config section.
// Extend this as you add more providers and fields.
func LinkAuthConfigForProvider(provider string, cloudConfig map[string]interface{}) *AuthConfig {
	ac := &AuthConfig{}
	switch provider {
	case "dropbox":
		if token, ok := cloudConfig["dropbox_access_token"].(string); ok {
			ac.DropboxAccessToken = token
		}
		// Add more providers here, e.g.:
		// case "gdrive":
		//   if cred, ok := cloudConfig["gdrive_credentials"].(string); ok {
		//     ac.GDriveCredentials = cred
		//   }
	}
	return ac
}
