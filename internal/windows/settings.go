//go:build windows

package windows

import (
	"os"
	"path/filepath"

	"github.com/strobotti/linkquisition"
)

// NewPaths resolves the folders linkquisition uses on Windows.
// Config lives in %APPDATA%\linkquisition, logs in %LOCALAPPDATA%\linkquisition\logs,
// caches in %LOCALAPPDATA%\linkquisition\cache, and plugins are not supported on
// Windows (Plugins is empty).
func NewPaths() linkquisition.Paths {
	return linkquisition.Paths{
		Config: configFolderPath(),
		Log:    localFolderPath("logs"),
		Cache:  localFolderPath("cache"),
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

// localFolderPath returns the named folder in %LOCALAPPDATA%\linkquisition.
func localFolderPath(name string) string {
	localAppData, err := os.UserCacheDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "AppData", "Local", "linkquisition", name)
	}
	return filepath.Join(localAppData, "linkquisition", name)
}
