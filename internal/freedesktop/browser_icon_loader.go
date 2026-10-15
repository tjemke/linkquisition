//go:build linux

package freedesktop

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"

	"github.com/strobotti/linkquisition"
	"github.com/strobotti/linkquisition/resources"
)

// BrowserIconLoader loads browser icons and resolves icon names to file paths.
type BrowserIconLoader interface {
	LoadIcon(browser linkquisition.Browser) ([]byte, error)
	// ResolveIconName resolves an icon name (from a .desktop file's Icon= field) to an
	// absolute file path. If the name is already an absolute path and the file exists,
	// it is returned as-is. Otherwise, it is looked up in the installed icon themes.
	// Returns an empty string if the icon cannot be found.
	ResolveIconName(iconName string) string
}

// DefaultBrowserIconLoader is the standard implementation of BrowserIconLoader.
type DefaultBrowserIconLoader struct {
	XdgService          *XdgService
	DesktopEntryService *DesktopEntryService
	BaseDirs            *BaseDirs
}

func (l *DefaultBrowserIconLoader) LoadIcon(browser linkquisition.Browser) ([]byte, error) {
	// If the browser has a cached icon path, use it directly
	if browser.IconPath != "" {
		if icon, errLoad := fyne.LoadResourceFromPath(browser.IconPath); errLoad == nil {
			return icon.Content(), nil
		}
	}

	// Fall back to resolving from .desktop entry
	iconPath := l.resolveIconPathForCommand(browser.Command)
	if iconPath != "" {
		if icon, errLoad := fyne.LoadResourceFromPath(iconPath); errLoad == nil {
			return icon.Content(), nil
		}
	}

	// As a fallback we'll use the bundled application icon
	return resources.LinkquisitionIcon.Content(), nil
}

// resolveIconPathForCommand finds the icon path for a browser command by looking up
// the .desktop entry and resolving its Icon field to an absolute path.
func (l *DefaultBrowserIconLoader) resolveIconPathForCommand(command string) string {
	dePath, err := l.XdgService.GetDesktopEntryPathForBinary(command)
	if err != nil {
		return ""
	}

	desktopEntry, err := l.DesktopEntryService.CreateFromPath(dePath)
	if err != nil {
		return ""
	}

	return l.ResolveIconName(desktopEntry.Icon)
}

// ResolveIconName resolves an icon name (from a .desktop file's Icon= field) to an
// absolute file path. If the name is already an absolute path and the file exists,
// it is returned as-is. Otherwise, it is looked up as the Icon Theme spec
// describes, see BaseDirs.FindIcon. Returns an empty string if the icon cannot be found.
func (l *DefaultBrowserIconLoader) ResolveIconName(iconName string) string {
	if iconName == "" {
		return ""
	}

	// If the icon field is already a full path, use it directly if it exists
	if filepath.IsAbs(iconName) {
		if _, err := os.Stat(iconName); err == nil {
			return iconName
		}
		return ""
	}

	return l.BaseDirs.FindIcon(iconName, iconLookupSize)
}
