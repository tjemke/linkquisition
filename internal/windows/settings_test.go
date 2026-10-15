//go:build windows

package windows

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaths_Config(t *testing.T) {
	path := NewPaths().Config

	assert.NotEmpty(t, path)
	assert.True(t, strings.HasSuffix(path, filepath.Join("linkquisition")))
	// Should be under APPDATA
	appData := os.Getenv("APPDATA")
	if appData != "" {
		assert.True(t, strings.HasPrefix(path, appData))
	}
}

func TestNewPaths_Log(t *testing.T) {
	path := NewPaths().Log

	assert.NotEmpty(t, path)
	assert.True(t, strings.HasSuffix(path, filepath.Join("linkquisition", "logs")))
}

func TestNewPaths_Plugins(t *testing.T) {
	path := NewPaths().Plugins

	assert.Empty(t, path, "plugins are not supported on Windows")
}

func TestNewPaths_Cache(t *testing.T) {
	path := NewPaths().Cache

	assert.NotEmpty(t, path)
	assert.True(t, strings.HasSuffix(path, filepath.Join("linkquisition", "cache")))
}
