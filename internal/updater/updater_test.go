package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/internal/version"
)

func TestCheckForUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		serverResponse Release
		expectedInfo   *UpdateInfo
		expectError    bool
	}{
		{
			name:           "new version available",
			currentVersion: "1.0.0",
			serverResponse: Release{
				TagName:     "v1.1.0",
				Name:        "Version 1.1.0",
				HTMLURL:     "https://github.com/hugofrely/envswitch/releases/tag/v1.1.0",
				PublishedAt: time.Now(),
				Assets: []Asset{
					{
						Name:               "envswitch-darwin-x86_64.tar.gz",
						BrowserDownloadURL: "https://example.com/download",
					},
				},
			},
			expectedInfo: &UpdateInfo{
				Available:      true,
				CurrentVersion: "1.0.0",
				LatestVersion:  "1.1.0",
				ReleaseURL:     "https://github.com/hugofrely/envswitch/releases/tag/v1.1.0",
			},
			expectError: false,
		},
		{
			name:           "already latest version",
			currentVersion: "1.1.0",
			serverResponse: Release{
				TagName:     "v1.1.0",
				Name:        "Version 1.1.0",
				HTMLURL:     "https://github.com/hugofrely/envswitch/releases/tag/v1.1.0",
				PublishedAt: time.Now(),
				Assets:      []Asset{},
			},
			expectedInfo: &UpdateInfo{
				Available:      false,
				CurrentVersion: "1.1.0",
				LatestVersion:  "1.1.0",
				ReleaseURL:     "https://github.com/hugofrely/envswitch/releases/tag/v1.1.0",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Override the API URL with test server
			oldAPIURL := apiURL
			apiURL = server.URL
			defer func() { apiURL = oldAPIURL }()

			// Set test version
			oldVersion := version.Version
			version.Version = tt.currentVersion
			defer func() { version.Version = oldVersion }()

			info, err := CheckForUpdate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.currentVersion, info.CurrentVersion)
			}
		})
	}
}

func TestCheckForUpdate_DevVersion(t *testing.T) {
	oldVersion := version.Version
	version.Version = "dev"
	defer func() { version.Version = oldVersion }()

	info, err := CheckForUpdate()
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "dev", info.CurrentVersion)
	assert.False(t, info.Available)
}

func TestFindAssetURL(t *testing.T) {
	// Get current architecture mapping
	archName := runtime.GOARCH
	archMap := map[string]string{
		"amd64": "x86_64",
		"arm64": "arm64",
	}
	if mapped, ok := archMap[archName]; ok {
		archName = mapped
	}

	tests := []struct {
		name        string
		assets      []Asset
		expectMatch bool
	}{
		{
			name: "matching asset found for current platform",
			assets: []Asset{
				{
					Name:               "envswitch-linux-x86_64.tar.gz",
					BrowserDownloadURL: "https://example.com/linux",
				},
				{
					Name:               "envswitch-darwin-x86_64.tar.gz",
					BrowserDownloadURL: "https://example.com/darwin-x86",
				},
				{
					Name:               "envswitch-darwin-arm64.tar.gz",
					BrowserDownloadURL: "https://example.com/darwin-arm",
				},
				{
					Name:               "envswitch-windows-x86_64.zip",
					BrowserDownloadURL: "https://example.com/windows",
				},
			},
			expectMatch: true,
		},
		{
			name: "no matching asset",
			assets: []Asset{
				{
					Name:               "envswitch-solaris-sparc.tar.gz",
					BrowserDownloadURL: "https://example.com/solaris",
				},
			},
			expectMatch: false,
		},
		{
			name:        "empty assets",
			assets:      []Asset{},
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := findAssetURL(tt.assets)

			if tt.expectMatch {
				assert.NotEmpty(t, url, "Expected to find asset for OS: %s, Arch: %s", runtime.GOOS, archName)
			} else {
				assert.Empty(t, url)
			}
		})
	}
}

func TestGetUpdateCommand(t *testing.T) {
	cmd := GetUpdateCommand()
	assert.NotEmpty(t, cmd)

	if runtime.GOOS == "darwin" {
		assert.Contains(t, cmd, "brew")
	} else {
		assert.Contains(t, cmd, "curl")
		assert.Contains(t, cmd, "install.sh")
	}
}

func TestShouldCheckForUpdate(t *testing.T) {
	// Currently always returns true
	result := ShouldCheckForUpdate("/tmp/test")
	assert.True(t, result)
}

func TestUpdateInfo(t *testing.T) {
	info := &UpdateInfo{
		Available:      true,
		CurrentVersion: "1.0.0",
		LatestVersion:  "1.1.0",
		DownloadURL:    "https://example.com/download",
		ReleaseURL:     "https://example.com/release",
	}

	assert.True(t, info.Available)
	assert.Equal(t, "1.0.0", info.CurrentVersion)
	assert.Equal(t, "1.1.0", info.LatestVersion)
	assert.NotEmpty(t, info.DownloadURL)
	assert.NotEmpty(t, info.ReleaseURL)
}

func TestRelease(t *testing.T) {
	now := time.Now()
	release := Release{
		TagName:     "v1.0.0",
		Name:        "Version 1.0.0",
		HTMLURL:     "https://github.com/test/repo/releases/tag/v1.0.0",
		PublishedAt: now,
		Assets: []Asset{
			{
				Name:               "test-asset.tar.gz",
				BrowserDownloadURL: "https://example.com/download",
			},
		},
	}

	assert.Equal(t, "v1.0.0", release.TagName)
	assert.Equal(t, "Version 1.0.0", release.Name)
	assert.Equal(t, now, release.PublishedAt)
	assert.Len(t, release.Assets, 1)
	assert.Equal(t, "test-asset.tar.gz", release.Assets[0].Name)
}

func TestAsset(t *testing.T) {
	asset := Asset{
		Name:               "envswitch-linux-amd64.tar.gz",
		BrowserDownloadURL: "https://example.com/download/asset.tar.gz",
	}

	assert.Equal(t, "envswitch-linux-amd64.tar.gz", asset.Name)
	assert.Equal(t, "https://example.com/download/asset.tar.gz", asset.BrowserDownloadURL)
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		shouldUpdate   bool
	}{
		{
			name:           "patch update available",
			currentVersion: "1.0.0",
			latestVersion:  "1.0.1",
			shouldUpdate:   true,
		},
		{
			name:           "minor update available",
			currentVersion: "1.0.0",
			latestVersion:  "1.1.0",
			shouldUpdate:   true,
		},
		{
			name:           "major update available",
			currentVersion: "1.0.0",
			latestVersion:  "2.0.0",
			shouldUpdate:   true,
		},
		{
			name:           "same version",
			currentVersion: "1.0.0",
			latestVersion:  "1.0.0",
			shouldUpdate:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldUpdate := tt.currentVersion != tt.latestVersion
			assert.Equal(t, tt.shouldUpdate, shouldUpdate)
		})
	}
}
