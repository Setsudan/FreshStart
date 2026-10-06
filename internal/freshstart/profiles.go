package freshstart

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultProfilesDir is the usual relative path for checked-in profiles.
const DefaultProfilesDir = "profiles"

// ListProfiles returns the names of profile files in dir (non-directories only),
// sorted alphabetically. Missing dir returns an error.
func ListProfiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

// LoadProfile reads and parses the profile named name under dir.
// name may be a bare file name (joined with dir) or an absolute/relative path;
// if name contains a path separator or is absolute, dir is ignored.
func LoadProfile(dir, name string) (ParseResult, string, error) {
	path := profilePath(dir, name)
	result, err := ParseLinksFile(path)
	if err != nil {
		return ParseResult{}, path, err
	}
	return result, path, nil
}

// SaveProfile writes content to the profile named name under dir.
// Content is stored as-is (typically a link list text). Parent directories
// are created if needed when name includes subfolders.
func SaveProfile(dir, name, content string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("profile name is empty")
	}
	path := profilePath(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return path, fmt.Errorf("create profile dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return path, fmt.Errorf("save profile: %w", err)
	}
	return path, nil
}

func profilePath(dir, name string) string {
	if filepath.IsAbs(name) || strings.ContainsRune(name, os.PathSeparator) || strings.Contains(name, "/") {
		return filepath.Clean(name)
	}
	return filepath.Join(dir, name)
}
