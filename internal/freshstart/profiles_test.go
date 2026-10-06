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
