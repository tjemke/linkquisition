//go:build linux

package freedesktop

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/ini.v1"
)

// Icon lookup following the freedesktop.org Icon Theme Specification, see
// https://specifications.freedesktop.org/icon-theme-spec/latest/
//
// Icons are found by checking the paths a theme's index.theme says they can be
// at, never by walking directory trees. Walking skips symlinked directories,
// and NixOS profiles are made of them.

const (
	// iconLookupSize is the nominal size, in pixels, of the icons looked up.
	// The picker shows icons at up to 96px and browsers almost always ship 128px.
	iconLookupSize = 128

	// fallbackIconTheme is the theme every icon theme falls back to, and where
	// applications install their icons.
	fallbackIconTheme = "hicolor"

	iconTypeFixed        = "Fixed"
	iconTypeScalable     = "Scalable"
	iconTypeThreshold    = "Threshold"
	defaultIconThreshold = 2
)

// iconExtensions lists the icon file types to look for, in order of preference.
// The spec also allows .xpm, which the UI can't display.
var iconExtensions = []string{".png", ".svg"}

// indexThemeLoadOptions parses index.theme files as plain key files. By default
// ini.v1 treats "a.b" as a child section of "a" inheriting its keys, but folder
// names may contain dots.
var indexThemeLoadOptions = ini.LoadOptions{IgnoreInlineComment: true, ChildSectionDelimiter: "\x00"}

// IconDirs returns the base directories of icon themes, most important first:
// ~/.icons (kept for backwards compatibility), then the "icons" folder of each
// XDG data directory.
func (d *BaseDirs) IconDirs() []string {
	dirs := []string{filepath.Join(d.Home, ".icons")}
	for _, dir := range d.DataSearchPath() {
		dirs = append(dirs, filepath.Join(dir, "icons"))
	}

	return uniquePaths(dirs)
}

// PixmapDirs returns the folders searched for icons that aren't part of any
// theme: the "pixmaps" folder of each XDG data directory, and /usr/share/pixmaps
// as named by the spec.
func (d *BaseDirs) PixmapDirs() []string {
	var dirs []string
	for _, dir := range d.DataSearchPath() {
		dirs = append(dirs, filepath.Join(dir, "pixmaps"))
	}

	return uniquePaths(append(dirs, "/usr/share/pixmaps"))
}

// FindIcon returns the path of the icon with the given name at, or as close as
// possible to, the given size, or "" if there's none. It looks in the hicolor
// theme, then for an unthemed icon directly in the base directories and the
// pixmaps folders, then in any other installed theme.
//
// Users' theme selection is desktop-specific and not covered by the spec, so
// it isn't consulted. Browsers install their icons into hicolor.
func (d *BaseDirs) FindIcon(name string, size int) string {
	baseDirs := d.IconDirs()

	if theme := loadIconTheme(baseDirs, fallbackIconTheme); theme != nil {
		if len(theme.subdirs) == 0 {
			// hicolor's index.theme comes with the hicolor-icon-theme package,
			// which minimal systems may lack while apps still install icons there.
			theme.subdirs = hicolorAppsLayout()
		}
		if path := theme.lookup(name, size); path != "" {
			return path
		}
	}

	// The spec's LookupFallbackIcon
	for _, dir := range append(slices.Clone(baseDirs), d.PixmapDirs()...) {
		if path := findIconFile(dir, name); path != "" {
			return path
		}
	}

	// Not part of the spec: also look through the other installed themes, so
	// an icon that only ships in a specific theme is still found.
	for _, themeName := range installedIconThemes(baseDirs) {
		if themeName == fallbackIconTheme {
			continue
		}
		if theme := loadIconTheme(baseDirs, themeName); theme != nil {
			if path := theme.lookup(name, size); path != "" {
				return path
			}
		}
	}

	return ""
}

// iconTheme is an installed icon theme and its layout, as described by its
// index.theme.
type iconTheme struct {
	// dirs are the theme's folders in the base directories that have one.
	dirs    []string
	subdirs []iconSubdir
}

// iconSubdir describes one of an icon theme's folders, e.g. "48x48/apps".
type iconSubdir struct {
	path      string
	size      int
	scale     int
	iconType  string
	minSize   int
	maxSize   int
	threshold int
}

// loadIconTheme finds the named theme's folders in the base directories and
// reads its layout from the first index.theme among them. It returns nil if the
// theme isn't installed, and a theme without subdirs if it has no index.theme.
func loadIconTheme(baseDirs []string, name string) *iconTheme {
	theme := &iconTheme{}
	for _, dir := range baseDirs {
		if info, err := os.Stat(filepath.Join(dir, name)); err == nil && info.IsDir() {
			theme.dirs = append(theme.dirs, filepath.Join(dir, name))
		}
	}
	if len(theme.dirs) == 0 {
		return nil
	}

	for _, dir := range theme.dirs {
		index, err := ini.LoadSources(indexThemeLoadOptions, filepath.Join(dir, "index.theme"))
		if err != nil {
			continue
		}

		section := index.Section("Icon Theme")
		subdirs := append(section.Key("Directories").Strings(","), section.Key("ScaledDirectories").Strings(",")...)
		for _, subdir := range subdirs {
			keys := index.Section(subdir)
			size := keys.Key("Size").MustInt()
			if size <= 0 {
				continue
			}
			theme.subdirs = append(theme.subdirs, iconSubdir{
				path:      subdir,
				size:      size,
				scale:     keys.Key("Scale").MustInt(1),
				iconType:  keys.Key("Type").String(),
				minSize:   keys.Key("MinSize").MustInt(size),
				maxSize:   keys.Key("MaxSize").MustInt(size),
				threshold: keys.Key("Threshold").MustInt(defaultIconThreshold),
			})
		}

		break
	}

	return theme
}

// hicolorAppsLayout describes hicolor's application icon folders as its
// index.theme does, for when that file is missing.
func hicolorAppsLayout() []iconSubdir {
	var subdirs []iconSubdir
	for _, size := range []int{16, 22, 24, 32, 36, 48, 64, 72, 96, 128, 192, 256, 512} {
		subdirs = append(subdirs, iconSubdir{
			path:      fmt.Sprintf("%dx%d/apps", size, size),
			size:      size,
			scale:     1,
			iconType:  iconTypeThreshold,
			threshold: defaultIconThreshold,
		})
	}

	return append(subdirs, iconSubdir{
		path: "scalable/apps", size: 128, scale: 1, iconType: iconTypeScalable, minSize: 1, maxSize: 256,
	})
}

// lookup implements the spec's LookupIcon for an icon scale of 1: the first
// folder matching the size wins, otherwise the closest size does. The folders
// listed in index.theme are looked for in every base directory having the theme.
func (t *iconTheme) lookup(name string, size int) string {
	for _, subdir := range t.subdirs {
		if !subdir.matchesSize(size) {
			continue
		}
		if path := t.findIconFile(subdir, name); path != "" {
			return path
		}
	}

	closest, minDistance := "", math.MaxInt
	for _, subdir := range t.subdirs {
		distance := subdir.sizeDistance(size)
		if distance >= minDistance {
			continue
		}
		if path := t.findIconFile(subdir, name); path != "" {
			closest, minDistance = path, distance
		}
	}

	return closest
}

func (t *iconTheme) findIconFile(subdir iconSubdir, name string) string {
	for _, dir := range t.dirs {
		if path := findIconFile(filepath.Join(dir, subdir.path), name); path != "" {
			return path
		}
	}

	return ""
}

// matchesSize implements the spec's DirectoryMatchesSize for an icon scale of 1.
func (s iconSubdir) matchesSize(size int) bool {
	if s.scale != 1 {
		return false
	}

	switch s.iconType {
	case iconTypeFixed:
		return s.size == size
	case iconTypeScalable:
		return s.minSize <= size && size <= s.maxSize
	default: // Threshold
		return s.size-s.threshold <= size && size <= s.size+s.threshold
	}
}

// sizeDistance implements the spec's DirectorySizeDistance for an icon scale of
// 1. The Threshold case uses Size±Threshold, like GTK does, instead of the
// spec's MinSize/MaxSize, which is a known erratum.
func (s iconSubdir) sizeDistance(size int) int {
	minSize, maxSize := s.size, s.size
	switch s.iconType {
	case iconTypeFixed:
	case iconTypeScalable:
		minSize, maxSize = s.minSize, s.maxSize
	default: // Threshold
		minSize, maxSize = s.size-s.threshold, s.size+s.threshold
	}

	switch {
	case size < minSize*s.scale:
		return minSize*s.scale - size
	case size > maxSize*s.scale:
		return size - maxSize*s.scale
	default:
		return 0
	}
}

// findIconFile returns the path of the icon file with the given name in dir.
// os.Stat follows symlinks, so icons symlinked into NixOS profiles are found.
func findIconFile(dir, name string) string {
	for _, ext := range iconExtensions {
		path := filepath.Join(dir, name+ext)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

// installedIconThemes returns the names of the themes with an index.theme in
// any of the base directories, sorted.
func installedIconThemes(baseDirs []string) []string {
	var names []string
	for _, dir := range baseDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if _, err := os.Stat(filepath.Join(dir, entry.Name(), "index.theme")); err == nil {
				names = append(names, entry.Name())
			}
		}
	}
	slices.Sort(names)

	return slices.Compact(names)
}
