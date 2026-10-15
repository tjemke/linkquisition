//go:build darwin

package darwin

import (
	"os"
	"path/filepath"

	"github.com/strobotti/linkquisition"
)

// NewPaths resolves the folders linkquisition uses on macOS.
func NewPaths() linkquisition.Paths {
	return linkquisition.Paths{
		Config:  configFolderPath(),
		Log:     logFolderPath(),
		Cache:   cacheFolderPath(),
		Plugins: pluginFolderPath(),
	}
}

func configFolderPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "Library", "Application Support", "linkquisition")
	}
	return filepath.Join(configDir, "linkquisition")
}

func logFolderPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "linkquisition")
	}
	return filepath.Join(homeDir, "Library", "Logs", "linkquisition")
}

func cacheFolderPath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "linkquisition")
	}
	return filepath.Join(cacheDir, "linkquisition")
}

func pluginFolderPath() string {
	execPath, err := os.Executable()
	if err == nil {
		bundlePath := filepath.Join(filepath.Dir(execPath), "..", "Resources", "plugins")
		if _, statErr := os.Stat(bundlePath); statErr == nil {
			return bundlePath
		}
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "Application Support", "linkquisition", "plugins")
}
