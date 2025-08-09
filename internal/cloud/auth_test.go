package cloud

import (
	"reflect"
	"testing"

	"github.com/sayuyere/storageX/internal/config"
)

func TestAuthConfigFromCloudConfig_Struct(t *testing.T) {
	input := config.CloudConfig{DropboxAccessTokens: []string{"token1", "token2"}}
	got := AuthConfigFromCloudConfig(&input)
	want := []AuthConfig{{DropboxAccessToken: "token1"}, {DropboxAccessToken: "token2"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AuthConfigFromCloudConfig(struct) = %v, want %v", got, want)
	}
}

func TestLinkAuthConfigForProvider_Dropbox(t *testing.T) {
	input := map[string]interface{}{"dropbox_access_token": "dbtoken"}
	ac := LinkAuthConfigForProvider("dropbox", input)
	if ac.DropboxAccessToken != "dbtoken" {
		t.Errorf("LinkAuthConfigForProvider(dropbox) = %v, want DropboxAccessToken=dbtoken", ac)
	}
}

func TestLinkAuthConfigForProvider_UnknownProvider(t *testing.T) {
	input := map[string]interface{}{"irrelevant": "value"}
	ac := LinkAuthConfigForProvider("unknown", input)
	if *ac != (AuthConfig{}) {
		t.Errorf("LinkAuthConfigForProvider(unknown) = %v, want zero AuthConfig", ac)
	}
}
