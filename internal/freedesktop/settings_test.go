//go:build linux

package freedesktop

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPluginFolderPath_NextToExecutable(t *testing.T) {
	prefix := filepath.Join(t.TempDir(), "app")
	executable := writeTestFile(t, filepath.Join(prefix, "bin", "linkquisition"), "")
	plugins := filepath.Join(prefix, "lib", "linkquisition", "plugins")
	require.NoError(t, os.MkdirAll(plugins, 0755))

	assert.Equal(t, plugins, pluginFolderPath(executable))
}

// On NixOS the executable is reached through profile symlinks, while the
// plugins are next to the real executable in the Nix store.
func TestPluginFolderPath_ResolvesSymlinkedExecutable(t *testing.T) {
	dir := t.TempDir()
	store := filepath.Join(dir, "nix", "store", "abc-linkquisition")
	executable := writeTestFile(t, filepath.Join(store, "bin", "linkquisition"), "")
	plugins := filepath.Join(store, "lib", "linkquisition", "plugins")
	require.NoError(t, os.MkdirAll(plugins, 0755))
	link := filepath.Join(dir, "run", "current-system", "sw", "bin", "linkquisition")
	require.NoError(t, os.MkdirAll(filepath.Dir(link), 0755))
	require.NoError(t, os.Symlink(executable, link))

	assert.Equal(t, plugins, pluginFolderPath(link))
}

func TestPluginFolderPath_FallsBackToSystemFolder(t *testing.T) {
	executable := writeTestFile(t, filepath.Join(t.TempDir(), "bin", "linkquisition-linux-amd64"), "")

	assert.Equal(t, "/usr/lib/linkquisition/plugins", pluginFolderPath(executable))
}

// os.Executable can fail, leaving an empty path. filepath.EvalSymlinks("")
// returns ".", which must not make the plugin folder relative to the working
// directory.
func TestPluginFolderPath_UnknownExecutable(t *testing.T) {
	cwd := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(cwd, "..", "lib", "linkquisition", "plugins"), 0755))
	t.Chdir(cwd)

	assert.Equal(t, "/usr/lib/linkquisition/plugins", pluginFolderPath(""))
}
