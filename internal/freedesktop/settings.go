//go:build linux

package freedesktop

import (
	"os"
	"path/filepath"

	"github.com/strobotti/linkquisition"
)

// NewPaths resolves the folders linkquisition uses on Linux.
func NewPaths() linkquisition.Paths {
	return linkquisition.Paths{
		Config:  configFolderPath(),
		Log:     logFolderPath(),
		Plugins: "/usr/lib/linkquisition/plugins",
	}
}

func configFolderPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ".config"
	}

	return filepath.Join(configDir, "linkquisition")
}

func logFolderPath() string {
	stateDir, isset := os.LookupEnv("XDG_STATE_HOME")
	if isset {
		return filepath.Join(stateDir, "linkquisition")
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(homeDir, ".local", "state", "linkquisition")
	}

	return filepath.Join(os.TempDir(), "linkquisition")
}
