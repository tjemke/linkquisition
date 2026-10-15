//go:build linux

package freedesktop

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaths_UsesBaseDirs(t *testing.T) {
	dir := t.TempDir()
	dirs := &BaseDirs{
		ConfigHome: filepath.Join(dir, "config"),
		StateHome:  filepath.Join(dir, "state"),
	}

	paths := NewPaths(dirs)

	assert.Equal(t, filepath.Join(dir, "config", "linkquisition"), paths.Config)
	assert.Equal(t, filepath.Join(dir, "state", "linkquisition"), paths.Log)
}

// An empty XDG_STATE_HOME used to put the log folder at "linkquisition",
// relative to the working directory.
func TestNewPaths_EmptyStateHomeFallsBackToDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_STATE_HOME", "")

	paths := NewPaths(NewBaseDirs())

	assert.Equal(t, filepath.Join(home, ".local", "state", "linkquisition"), paths.Log)
}
