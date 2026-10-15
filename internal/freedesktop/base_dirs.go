//go:build linux

package freedesktop

import (
	"slices"

	"github.com/adrg/xdg"
)

// BaseDirs is a snapshot of the XDG base directories, see
// https://specifications.freedesktop.org/basedir-spec/latest/
//
// The values come from github.com/adrg/xdg, which applies the spec's rules:
// variables that are unset, empty or not absolute fall back to the defaults,
// and the *Dirs lists keep only unique absolute entries, in order.
//
// Code in this package that looks for files in XDG locations receives a
// BaseDirs instead of reading the environment itself, so a single snapshot is
// shared and tests can point it at temporary directories.
type BaseDirs struct {
	// Home is the user's home directory.
	Home string

	// ConfigHome is $XDG_CONFIG_HOME, default ~/.config.
	ConfigHome string

	// ConfigDirs is $XDG_CONFIG_DIRS, default /etc/xdg.
	ConfigDirs []string

	// DataHome is $XDG_DATA_HOME, default ~/.local/share.
	DataHome string

	// DataDirs is $XDG_DATA_DIRS, default /usr/local/share:/usr/share.
	DataDirs []string

	// StateHome is $XDG_STATE_HOME, default ~/.local/state.
	StateHome string

	// CacheHome is $XDG_CACHE_HOME, default ~/.cache.
	CacheHome string
}

// NewBaseDirs reads the XDG base directories from the current environment.
func NewBaseDirs() *BaseDirs {
	xdg.Reload()

	return &BaseDirs{
		Home:       xdg.Home,
		ConfigHome: xdg.ConfigHome,
		ConfigDirs: slices.Clone(xdg.ConfigDirs),
		DataHome:   xdg.DataHome,
		DataDirs:   slices.Clone(xdg.DataDirs),
		StateHome:  xdg.StateHome,
		CacheHome:  xdg.CacheHome,
	}
}

// DataSearchPath returns the directories to search for data files, most
// important first: DataHome followed by DataDirs.
func (d *BaseDirs) DataSearchPath() []string {
	return uniquePaths(append([]string{d.DataHome}, d.DataDirs...))
}

// uniquePaths returns paths without empty and repeated entries, keeping the
// first occurrence of each.
func uniquePaths(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		unique = append(unique, path)
	}

	return unique
}
