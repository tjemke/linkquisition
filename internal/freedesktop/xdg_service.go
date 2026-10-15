//go:build linux

package freedesktop

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type XdgService struct {
	BaseDirs *BaseDirs
}

func (x *XdgService) GetDesktopEntryPathForFilename(name string) (string, error) {
	paths := x.GetApplicationPaths()

	if len(paths) == 0 {
		return "", fmt.Errorf("no valid desktop entry paths found in $XDG_DATA_DIRS")
	}

	for _, path := range paths {
		desktopEntryPath := filepath.Join(path, name)
		if _, err := os.Stat(desktopEntryPath); err == nil {
			return desktopEntryPath, nil
		}
	}

	return "", fmt.Errorf("no .desktop entry found for %s", name)
}

func (x *XdgService) GetDesktopEntryPathForBinary(command string) (string, error) {
	paths := x.GetApplicationPaths()

	if len(paths) == 0 {
		return "", fmt.Errorf("no valid desktop entry paths found in $XDG_DATA_DIRS")
	}

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			entryPath := filepath.Join(dir, entry.Name())
			if desktopEntryExecMatches(entryPath, command) {
				return entryPath, nil
			}
		}
	}

	return "", fmt.Errorf("no .desktop entry found for %s", command)
}

// GetApplicationPaths returns the existing "applications" folders under the XDG
// data directories, most important first: $XDG_DATA_HOME, then each entry of
// $XDG_DATA_DIRS. This is where desktop entries are installed, e.g. under
// /run/current-system/sw/share and the user's profiles on NixOS, or
// /var/lib/flatpak/exports/share for Flatpak apps.
func (x *XdgService) GetApplicationPaths() []string {
	var paths []string

	for _, dir := range x.BaseDirs.DataSearchPath() {
		appsDir := filepath.Join(dir, "applications")
		if _, err := os.Stat(appsDir); err == nil {
			paths = append(paths, appsDir)
		}
	}

	return paths
}

func (x *XdgService) SettingsCheck(property, subProperty string) (string, error) {
	cmd := exec.Command("xdg-settings", "check", property, subProperty)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to call xdg-settings check: %v", err)
	}

	// TODO not exatcly sure if this is The Way™
	return strings.Trim(out.String(), "\n"), nil
}

func (x *XdgService) SettingsGet(property string) (string, error) {
	cmd := exec.Command("xdg-settings", "get", property)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to call xdg-settings get: %v", err)
	}
	return out.String(), nil
}

func (x *XdgService) SettingsSet(property, value string) error {
	cmd := exec.Command("xdg-settings", "set", property, value)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to call xdg-settings set: %v", err)
	}
	return nil
}

// extractExecutable parses a command string and returns the executable path.
// It handles environment variable expansion ($HOME) and tilde expansion (~),
// and strips any arguments or field codes (%u, %U, etc.).
func extractExecutable(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}

	// Split into tokens and take the first one (the executable)
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	exe := fields[0]

	// Expand environment variables (e.g. $HOME)
	exe = os.ExpandEnv(exe)

	// Expand tilde
	if strings.HasPrefix(exe, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			exe = filepath.Join(home, exe[2:])
		}
	}

	return exe
}

// desktopEntryExecMatches checks if a .desktop file's Exec= line uses
// the same executable as the given command. It extracts the binary path
// from both the command and the Exec= line and compares them by full path
// or by basename.
func desktopEntryExecMatches(path, command string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	commandExe := extractExecutable(command)
	if commandExe == "" {
		return false
	}

	prefix := "Exec="
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			execValue := strings.TrimPrefix(line, prefix)
			entryExe := extractExecutable(execValue)
			if entryExe == "" {
				return false
			}

			// Compare full paths if both are absolute
			if filepath.IsAbs(commandExe) && filepath.IsAbs(entryExe) {
				return commandExe == entryExe
			}

			// Otherwise compare basenames
			return filepath.Base(commandExe) == filepath.Base(entryExe)
		}
	}
	return false
}
