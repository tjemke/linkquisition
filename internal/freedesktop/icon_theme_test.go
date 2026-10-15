//go:build linux

package freedesktop

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHicolorIndex = `[Icon Theme]
Name=Hicolor
Comment=Fallback icon theme
Directories=48x48/apps,128x128/apps,128x128@2/apps,256x256/apps,scalable/apps

[48x48/apps]
Size=48
Context=Applications
Type=Threshold

[128x128/apps]
Size=128
Context=Applications
Type=Threshold

[128x128@2/apps]
Size=128
Scale=2
Context=Applications
Type=Threshold

[256x256/apps]
Size=256
Context=Applications
Type=Threshold

[scalable/apps]
MinSize=1
Size=128
MaxSize=256
Context=Applications
Type=Scalable
`

// testIconDirs is a fake XDG layout: a home directory, a user profile without
// an index.theme (like a NixOS per-user profile) and a system data dir with
// hicolor's index.theme.
type testIconDirs struct {
	*BaseDirs
	profile string
	system  string
}

func newTestIconDirs(t *testing.T) testIconDirs {
	t.Helper()
	root := t.TempDir()
	dirs := testIconDirs{
		BaseDirs: &BaseDirs{
			Home:     filepath.Join(root, "home"),
			DataHome: filepath.Join(root, "home", ".local", "share"),
		},
		profile: filepath.Join(root, "profile", "share"),
		system:  filepath.Join(root, "system", "share"),
	}
	dirs.DataDirs = []string{dirs.profile, dirs.system}
	writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "index.theme"), testHicolorIndex)

	return dirs
}

func writeTestFile(t *testing.T, path, content string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}

func TestFindIcon_PrefersExactSize(t *testing.T) {
	dirs := newTestIconDirs(t)
	for _, subdir := range []string{"48x48", "256x256", "scalable"} {
		writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", subdir, "apps", "firefox.png"), "png")
	}
	want := writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "128x128", "apps", "firefox.png"), "png")

	assert.Equal(t, want, dirs.FindIcon("firefox", iconLookupSize))
}

func TestFindIcon_FallsBackToClosestSize(t *testing.T) {
	dirs := newTestIconDirs(t)
	want := writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "48x48", "apps", "firefox.png"), "png")
	writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "256x256", "apps", "firefox.png"), "png")

	// 48 is 78px away from the 126-130 range 128x128 accepts, 256 is 126px away.
	assert.Equal(t, want, dirs.FindIcon("firefox", iconLookupSize))
}

func TestFindIcon_UsesScalableIcons(t *testing.T) {
	dirs := newTestIconDirs(t)
	writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "48x48", "apps", "org.gnome.Epiphany.png"), "png")
	want := writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "scalable", "apps", "org.gnome.Epiphany.svg"), "<svg/>")

	assert.Equal(t, want, dirs.FindIcon("org.gnome.Epiphany", iconLookupSize))
}

// 128x128@2 holds 256px images for 2x displays: it is listed before scalable
// but isn't an exact match for a 128px icon at scale 1.
func TestFindIcon_DoesNotMatchOtherScales(t *testing.T) {
	dirs := newTestIconDirs(t)
	writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "128x128@2", "apps", "firefox.png"), "png")
	want := writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "scalable", "apps", "firefox.svg"), "<svg/>")

	assert.Equal(t, want, dirs.FindIcon("firefox", iconLookupSize))
}

// On NixOS, a profile's icons/hicolor usually has no index.theme of its own:
// the layout comes from the first index.theme found in any base directory.
func TestFindIcon_ThemeLayoutAppliesToAllBaseDirs(t *testing.T) {
	dirs := newTestIconDirs(t)
	want := writeTestFile(t, filepath.Join(dirs.profile, "icons", "hicolor", "128x128", "apps", "librewolf.png"), "png")

	assert.Equal(t, want, dirs.FindIcon("librewolf", iconLookupSize))
}

// NixOS profiles are trees of symlinks into /nix/store, including symlinked
// directories, which the previous directory walk did not descend into.
func TestFindIcon_FollowsSymlinkedDirectories(t *testing.T) {
	dirs := newTestIconDirs(t)
	store := filepath.Join(t.TempDir(), "nix", "store", "abc-chromium", "share", "icons", "hicolor", "128x128", "apps")
	writeTestFile(t, filepath.Join(store, "chromium.png"), "png")
	link := filepath.Join(dirs.profile, "icons", "hicolor", "128x128", "apps")
	require.NoError(t, os.MkdirAll(filepath.Dir(link), 0755))
	require.NoError(t, os.Symlink(store, link))

	assert.Equal(t, filepath.Join(link, "chromium.png"), dirs.FindIcon("chromium", iconLookupSize))
}

func TestFindIcon_HomeIconsComeFirst(t *testing.T) {
	dirs := newTestIconDirs(t)
	writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "128x128", "apps", "firefox.png"), "png")
	want := writeTestFile(t, filepath.Join(dirs.Home, ".icons", "hicolor", "128x128", "apps", "firefox.png"), "png")

	assert.Equal(t, want, dirs.FindIcon("firefox", iconLookupSize))
}

// Per the spec's LookupFallbackIcon, unthemed icons are also looked for
// directly in the base directories, e.g. ~/.local/share/icons/foo.png.
func TestFindIcon_FallsBackToBaseDirectories(t *testing.T) {
	dirs := newTestIconDirs(t)
	writeTestFile(t, filepath.Join(dirs.system, "pixmaps", "mybrowser.png"), "png")
	want := writeTestFile(t, filepath.Join(dirs.DataHome, "icons", "mybrowser.png"), "png")
	homeIcon := writeTestFile(t, filepath.Join(dirs.Home, ".icons", "other.svg"), "<svg/>")

	assert.Equal(t, want, dirs.FindIcon("mybrowser", iconLookupSize))
	assert.Equal(t, homeIcon, dirs.FindIcon("other", iconLookupSize))
}

func TestFindIcon_FallsBackToPixmaps(t *testing.T) {
	dirs := newTestIconDirs(t)
	writeTestFile(t, filepath.Join(dirs.profile, "pixmaps", "dillo.xpm"), "xpm")
	want := writeTestFile(t, filepath.Join(dirs.system, "pixmaps", "dillo.png"), "png")

	assert.Equal(t, want, dirs.FindIcon("dillo", iconLookupSize), "xpm can't be displayed and is skipped")
}

func TestFindIcon_FallsBackToOtherThemes(t *testing.T) {
	dirs := newTestIconDirs(t)
	writeTestFile(t, filepath.Join(dirs.system, "icons", "Papirus", "index.theme"),
		"[Icon Theme]\nName=Papirus\nDirectories=64x64/apps\n\n[64x64/apps]\nSize=64\nType=Fixed\n")
	want := writeTestFile(t, filepath.Join(dirs.system, "icons", "Papirus", "64x64", "apps", "web-browser.svg"), "<svg/>")
	writeTestFile(t, filepath.Join(dirs.system, "icons", "Papirus", "64x64", "apps", "firefox.svg"), "<svg/>")
	hicolorFirefox := writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "48x48", "apps", "firefox.png"), "png")

	assert.Equal(t, want, dirs.FindIcon("web-browser", iconLookupSize))
	assert.Equal(t, hicolorFirefox, dirs.FindIcon("firefox", iconLookupSize), "hicolor comes before other themes")
}

func TestFindIcon_NotFound(t *testing.T) {
	dirs := newTestIconDirs(t)

	assert.Empty(t, dirs.FindIcon("no-such-icon", iconLookupSize))
}

// Without the hicolor-icon-theme package there is no index.theme for hicolor,
// but applications still install their icons into it.
func TestFindIcon_WithoutHicolorIndex(t *testing.T) {
	dirs := newTestIconDirs(t)
	require.NoError(t, os.Remove(filepath.Join(dirs.system, "icons", "hicolor", "index.theme")))
	writeTestFile(t, filepath.Join(dirs.system, "pixmaps", "firefox.png"), "png")
	writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "48x48", "apps", "firefox.png"), "png")
	want := writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "128x128", "apps", "firefox.png"), "png")
	epiphany := writeTestFile(t, filepath.Join(dirs.profile, "icons", "hicolor", "scalable", "apps", "org.gnome.Epiphany.svg"), "<svg/>")

	assert.Equal(t, want, dirs.FindIcon("firefox", iconLookupSize))
	assert.Equal(t, epiphany, dirs.FindIcon("org.gnome.Epiphany", iconLookupSize))
}

func TestLoadIconTheme_DottedFolderNamesAreNotChildSections(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "Theme", "index.theme"),
		"[Icon Theme]\nDirectories=apps,apps.hidpi\n\n[apps]\nSize=48\n\n[apps.hidpi]\nContext=Applications\n")

	theme := loadIconTheme([]string{dir}, "Theme")

	require.NotNil(t, theme)
	require.Len(t, theme.subdirs, 1, "apps.hidpi has no Size and must not inherit the one of apps")
	assert.Equal(t, "apps", theme.subdirs[0].path)
}

func TestBaseDirs_IconDirs(t *testing.T) {
	dirs := &BaseDirs{Home: "/home/alice", DataHome: "/home/alice/.local/share", DataDirs: []string{"/run/current-system/sw/share"}}

	assert.Equal(t, []string{
		"/home/alice/.icons",
		"/home/alice/.local/share/icons",
		"/run/current-system/sw/share/icons",
	}, dirs.IconDirs())
	assert.Equal(t, []string{
		"/home/alice/.local/share/pixmaps",
		"/run/current-system/sw/share/pixmaps",
		"/usr/share/pixmaps",
	}, dirs.PixmapDirs())
}

func TestIconSubdir_SizeMatching(t *testing.T) {
	fixed := iconSubdir{size: 48, scale: 1, iconType: iconTypeFixed}
	threshold := iconSubdir{size: 48, scale: 1, iconType: "Threshold", threshold: 2}
	scalable := iconSubdir{size: 128, scale: 1, iconType: iconTypeScalable, minSize: 16, maxSize: 256}
	scaled := iconSubdir{size: 64, scale: 2, iconType: iconTypeFixed}

	for _, tt := range []struct {
		name     string
		subdir   iconSubdir
		size     int
		matches  bool
		distance int
	}{
		{"fixed exact", fixed, 48, true, 0},
		{"fixed smaller", fixed, 32, false, 16},
		{"fixed larger", fixed, 64, false, 16},
		{"threshold within", threshold, 50, true, 0},
		{"threshold below", threshold, 40, false, 6},
		{"threshold above", threshold, 128, false, 78},
		{"scalable within", scalable, 200, true, 0},
		{"scalable below", scalable, 8, false, 8},
		{"scalable above", scalable, 512, false, 256},
		{"other scale", scaled, 64, false, 64},
		{"other scale same pixels", scaled, 128, false, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.matches, tt.subdir.matchesSize(tt.size))
			assert.Equal(t, tt.distance, tt.subdir.sizeDistance(tt.size))
		})
	}
}

func TestResolveIconName(t *testing.T) {
	dirs := newTestIconDirs(t)
	themed := writeTestFile(t, filepath.Join(dirs.system, "icons", "hicolor", "128x128", "apps", "firefox.png"), "png")
	absolute := writeTestFile(t, filepath.Join(t.TempDir(), "opt", "browser", "icon.png"), "png")
	loader := &DefaultBrowserIconLoader{BaseDirs: dirs.BaseDirs}

	assert.Equal(t, themed, loader.ResolveIconName("firefox"))
	assert.Equal(t, absolute, loader.ResolveIconName(absolute))
	assert.Empty(t, loader.ResolveIconName(filepath.Join(t.TempDir(), "missing.png")))
	assert.Empty(t, loader.ResolveIconName(""))
}
