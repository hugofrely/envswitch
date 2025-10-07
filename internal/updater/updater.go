package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/hugofrely/envswitch/internal/version"
)

const (
	defaultGitHubAPIURL = "https://api.github.com/repos/hugofrely/envswitch/releases/latest"
	updateCheckFile     = ".last_update_check"
	checkInterval       = 24 * time.Hour // Check once per day
)

// apiURL is the GitHub API URL used for fetching releases
// Can be overridden for testing
var apiURL = defaultGitHubAPIURL

// Release represents a GitHub release
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	DownloadURL    string
	ReleaseURL     string
}

// CheckForUpdate checks if a new version is available
func CheckForUpdate() (*UpdateInfo, error) {
	info := &UpdateInfo{
		CurrentVersion: version.Version,
		Available:      false,
	}

	// Don't check for updates in dev mode
	if version.Version == version.DevVersion {
		return info, nil
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	info.LatestVersion = strings.TrimPrefix(release.TagName, "v")
	info.ReleaseURL = release.HTMLURL

	// Compare versions (normalize both by removing 'v' prefix)
	currentVersion := strings.TrimPrefix(info.CurrentVersion, "v")
	if info.LatestVersion != currentVersion {
		info.Available = true
		info.DownloadURL = findAssetURL(release.Assets)
	}

	return info, nil
}

// findAssetURL finds the appropriate download URL for the current platform
func findAssetURL(assets []Asset) string {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Common architecture mappings
	archMap := map[string]string{
		"amd64": "x86_64",
		"arm64": "arm64",
	}

	if mapped, ok := archMap[archName]; ok {
		archName = mapped
	}

	// Try to find matching asset
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) && strings.Contains(name, archName) {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

// GetUpdateCommand returns the command to update envswitch
func GetUpdateCommand() string {
	// Use install script for all platforms (macOS, Linux)
	return "curl -fsSL https://raw.githubusercontent.com/hugofrely/envswitch/main/install.sh | bash"
}

// ShouldCheckForUpdate determines if we should check for updates based on last check time
func ShouldCheckForUpdate(configDir string) bool {
	// For now, always return true. You can implement caching logic here
	// by storing the last check time in configDir/.last_update_check
	return true
}
