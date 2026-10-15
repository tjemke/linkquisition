//go:build linux

package freedesktop

import (
	"os"
	"path/filepath"

	"github.com/strobotti/linkquisition"
)

const (
	// appDirName is the name of linkquisition's folder inside each XDG base directory.
	appDirName = "linkquisition"

	// systemPluginFolder is where the .deb and .rpm packages install plugins.
	systemPluginFolder = "/usr/lib/linkquisition/plugins"
)

// NewPaths resolves the folders linkquisition uses on Linux from the XDG base
// directories: config in $XDG_CONFIG_HOME, logs in $XDG_STATE_HOME and caches
// in $XDG_CACHE_HOME. Plugins are installed alongside the executable, see
// pluginFolderPath.
func NewPaths(dirs *BaseDirs) linkquisition.Paths {
	executable, _ := os.Executable()

	return linkquisition.Paths{
		Config:  filepath.Join(dirs.ConfigHome, appDirName),
		Log:     filepath.Join(dirs.StateHome, appDirName),
		Cache:   filepath.Join(dirs.CacheHome, appDirName),
		Plugins: pluginFolderPath(executable),
	}
}

// pluginFolderPath returns the plugin folder installed with the given
// executable: <prefix>/lib/linkquisition/plugins for <prefix>/bin/linkquisition,
// after resolving symlinks. That covers /usr, /usr/local, the Nix store,
// Flatpak's /app and AppImage mounts. It falls back to systemPluginFolder when
// there is no such folder, e.g. for development builds, or the executable's path
// is unknown.
func pluginFolderPath(executable string) string {
	if executable == "" {
		return systemPluginFolder
	}

	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		dir := filepath.Join(filepath.Dir(resolved), "..", "lib", appDirName, "plugins")
		if info, statErr := os.Stat(dir); statErr == nil && info.IsDir() {
			return dir
		}
	}

	return systemPluginFolder
}
