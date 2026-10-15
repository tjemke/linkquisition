//go:build linux

package freedesktop

import (
	"path/filepath"

	"github.com/strobotti/linkquisition"
)

// appDirName is the name of linkquisition's folder inside each XDG base directory.
const appDirName = "linkquisition"

// NewPaths resolves the folders linkquisition uses on Linux from the XDG base
// directories: config in $XDG_CONFIG_HOME and logs in $XDG_STATE_HOME.
func NewPaths(dirs *BaseDirs) linkquisition.Paths {
	return linkquisition.Paths{
		Config:  filepath.Join(dirs.ConfigHome, appDirName),
		Log:     filepath.Join(dirs.StateHome, appDirName),
		Plugins: "/usr/lib/linkquisition/plugins",
	}
}
