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
// name must be a bare file name (no path separators or ".." traversal).
func LoadProfile(dir, name string) (ParseResult, string, error) {
	path, err := profilePath(dir, name)
	if err != nil {
		return ParseResult{}, "", err
	}
	result, err := ParseLinksFile(path)
	if err != nil {
		return ParseResult{}, path, err
	}
	return result, path, nil
}

// SaveProfile writes content to the profile named name under dir.
// Content is stored as-is (typically a link list text). Writes are atomic
// (temp file + rename) so a failed write does not truncate an existing profile.
func SaveProfile(dir, name, content string) (string, error) {
	path, err := profilePath(dir, name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return path, fmt.Errorf("create profile dir: %w", err)
	}
	if err := writeFileAtomic(path, []byte(content), 0o644); err != nil {
		return path, fmt.Errorf("save profile: %w", err)
	}
	return path, nil
}

// profilePath joins dir with a bare profile file name. Names with path
// separators, absolute paths, or ".." are rejected to prevent traversal.
func profilePath(dir, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("profile name is empty")
	}
	if name == "." || name == ".." || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name {
		return "", fmt.Errorf("profile name must be a bare file name without path separators")
	}
	return filepath.Join(dir, name), nil
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".freshstart-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		// Windows refuses rename over an existing file; remove then retry.
		if remErr := os.Remove(path); remErr != nil && !os.IsNotExist(remErr) {
			return err
		}
		if err2 := os.Rename(tmpName, path); err2 != nil {
			return err2
		}
	}
	cleanup = false
	return nil
}
