//go:build windows

package windows

import (
	"os"
	"path/filepath"

	"github.com/strobotti/linkquisition"
)

// NewPaths resolves the folders linkquisition uses on Windows.
// Config lives in %APPDATA%\linkquisition, logs in %LOCALAPPDATA%\linkquisition\logs,
// and plugins are not supported on Windows (Plugins is empty).
func NewPaths() linkquisition.Paths {
	return linkquisition.Paths{
		Config: configFolderPath(),
		Log:    logFolderPath(),
	}
}

func configFolderPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "AppData", "Roaming", "linkquisition")
	}
	return filepath.Join(configDir, "linkquisition")
}

func logFolderPath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "AppData", "Local", "linkquisition", "logs")
	}
	return filepath.Join(cacheDir, "linkquisition", "logs")
}
