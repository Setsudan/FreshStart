package freshstart

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListProfiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"z.txt", "a.txt", "mid"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("https://x.example\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	names, err := ListProfiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.txt", "mid", "z.txt"}
	if len(names) != len(want) {
		t.Fatalf("got %v want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("got %v want %v", names, want)
		}
	}
}

func TestListProfiles_MissingDir(t *testing.T) {
	_, err := ListProfiles(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadAndSaveProfile(t *testing.T) {
	dir := t.TempDir()
	content := "# p\nhttps://save.example\n"
	path, err := SaveProfile(dir, "mine.txt", content)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "mine.txt" {
		t.Fatalf("path=%s", path)
	}
	result, loadedPath, err := LoadProfile(dir, "mine.txt")
	if err != nil {
		t.Fatal(err)
	}
	if loadedPath != path {
		t.Fatalf("loaded path %s want %s", loadedPath, path)
	}
	if len(result.Links) != 1 || result.Links[0] != "https://save.example" {
		t.Fatalf("result=%+v", result)
	}
	// Comments preserved on disk (save stores content as-is).
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != content {
		t.Fatalf("raw=%q want %q", raw, content)
	}
}

func TestSaveProfile_EmptyName(t *testing.T) {
	_, err := SaveProfile(t.TempDir(), "  ", "x")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadProfile_Missing(t *testing.T) {
	_, _, err := LoadProfile(t.TempDir(), "gone.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProfilePath_RejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"../evil.txt", "..", "foo/bar", "foo\\bar", "/tmp/x"} {
		if _, err := SaveProfile(dir, name, "x"); err == nil {
			t.Fatalf("SaveProfile(%q) should reject traversal", name)
		}
		if _, _, err := LoadProfile(dir, name); err == nil {
			t.Fatalf("LoadProfile(%q) should reject traversal", name)
		}
	}
	// Ensure nothing escaped the temp dir.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty dir after rejected saves, got %v", entries)
	}
}

func TestSaveProfile_OverwritePreservesOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path, err := SaveProfile(dir, "p.txt", "first\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SaveProfile(dir, "p.txt", "second\n"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "second\n" {
		t.Fatalf("got %q", raw)
	}
	// No leftover temp files.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "p.txt" {
		t.Fatalf("entries=%v", entries)
	}
}
