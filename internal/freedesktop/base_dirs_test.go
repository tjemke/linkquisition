//go:build linux

package freedesktop

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseDirs_ReadsEnvironment(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", filepath.Join(dir, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_CONFIG_DIRS", filepath.Join(dir, "etc1")+":"+filepath.Join(dir, "etc2"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	t.Setenv("XDG_DATA_DIRS", filepath.Join(dir, "share1")+":"+filepath.Join(dir, "share2"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "state"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))

	dirs := NewBaseDirs()

	assert.Equal(t, filepath.Join(dir, "home"), dirs.Home)
	assert.Equal(t, filepath.Join(dir, "config"), dirs.ConfigHome)
	assert.Equal(t, []string{filepath.Join(dir, "etc1"), filepath.Join(dir, "etc2")}, dirs.ConfigDirs)
	assert.Equal(t, filepath.Join(dir, "data"), dirs.DataHome)
	assert.Equal(t, []string{filepath.Join(dir, "share1"), filepath.Join(dir, "share2")}, dirs.DataDirs)
	assert.Equal(t, filepath.Join(dir, "state"), dirs.StateHome)
	assert.Equal(t, filepath.Join(dir, "cache"), dirs.CacheHome)
}

func TestNewBaseDirs_DefaultsWhenUnset(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, name := range []string{
		"XDG_CONFIG_HOME", "XDG_CONFIG_DIRS", "XDG_DATA_HOME", "XDG_DATA_DIRS", "XDG_STATE_HOME", "XDG_CACHE_HOME",
	} {
		t.Setenv(name, "")
	}

	dirs := NewBaseDirs()

	assert.Equal(t, filepath.Join(home, ".config"), dirs.ConfigHome)
	assert.Equal(t, []string{"/etc/xdg"}, dirs.ConfigDirs)
	assert.Equal(t, filepath.Join(home, ".local", "share"), dirs.DataHome)
	assert.Equal(t, []string{"/usr/local/share", "/usr/share"}, dirs.DataDirs)
	assert.Equal(t, filepath.Join(home, ".local", "state"), dirs.StateHome)
	assert.Equal(t, filepath.Join(home, ".cache"), dirs.CacheHome)
}

// The spec says relative paths in XDG_* variables are invalid and must be
// ignored. Previously a relative XDG_CONFIG_HOME made the config folder ".config"
// in the working directory.
func TestNewBaseDirs_IgnoresRelativePaths(t *testing.T) {
	home := t.TempDir()
	share := filepath.Join(home, "share")
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "relative/config")
	t.Setenv("XDG_STATE_HOME", "relative/state")
	t.Setenv("XDG_DATA_DIRS", "relative/share:"+share+"::"+share)

	dirs := NewBaseDirs()

	assert.Equal(t, filepath.Join(home, ".config"), dirs.ConfigHome)
	assert.Equal(t, filepath.Join(home, ".local", "state"), dirs.StateHome)
	assert.Equal(t, []string{share}, dirs.DataDirs, "relative, empty and duplicate entries are dropped")
}
