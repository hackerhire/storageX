package cloud

// AuthConfig is a unified authentication configuration for all cloud providers
// Each provider can use the relevant fields

type AuthConfig struct {
	DropboxAccessToken string // Dropbox API access token
	// Add more fields as needed for other providers
}
