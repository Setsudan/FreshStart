package freshstart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLinks_CommentsAndBlanks(t *testing.T) {
	input := `
# header comment

https://example.com/a

  # indented comment
https://example.com/b
`
	r := ParseLinks(strings.NewReader(input))
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	if len(r.Links) != 2 {
		t.Fatalf("got %d links, want 2: %v", len(r.Links), r.Links)
	}
	if r.Links[0] != "https://example.com/a" || r.Links[1] != "https://example.com/b" {
		t.Fatalf("unexpected links: %v", r.Links)
	}
}

func TestParseLinks_InvalidURLs(t *testing.T) {
	input := `https://ok.example
not-a-url
ftp://wrong.scheme/x
http://
`
	r := ParseLinks(strings.NewReader(input))
	if len(r.Links) != 1 || r.Links[0] != "https://ok.example" {
		t.Fatalf("links=%v", r.Links)
	}
	if len(r.Errors) != 3 {
		t.Fatalf("want 3 errors, got %d: %v", len(r.Errors), r.Errors)
	}
	if r.Errors[0].Line != 2 {
		t.Errorf("first bad line want 2 got %d", r.Errors[0].Line)
	}
}

func TestValidateURL(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"https://example.com", false},
		{"http://localhost:8080/path", false},
		{"", true},
		{"example.com", true},
		{"ftp://example.com", true},
		{"://nohost", true},
	}
	for _, tc := range cases {
		err := ValidateURL(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateURL(%q) err=%v wantErr=%v", tc.in, err, tc.wantErr)
		}
	}
}

func TestParseLinksFile_Missing(t *testing.T) {
	_, err := ParseLinksFile(filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseLinksFile_OK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "links.txt")
	content := "# c\nhttps://a.example\n\nbad\nhttps://b.example\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := ParseLinksFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Links) != 2 {
		t.Fatalf("links=%v", r.Links)
	}
	if len(r.Errors) != 1 {
		t.Fatalf("errors=%v", r.Errors)
	}
}

func TestFormatLinks(t *testing.T) {
	if FormatLinks(nil) != "" {
		t.Fatal("empty should format to empty")
	}
	got := FormatLinks([]string{"https://a", "https://b"})
	want := "https://a\nhttps://b\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExtractLinkLines_ClassicCLI(t *testing.T) {
	// Classic CLI: no TrimSpace — leading-space comment is kept; URL spaces kept.
	input := "# c\n\nhttps://ok\nnot-validated\n  # not a comment\n  https://spaced\n"
	links, err := ExtractLinkLines(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://ok", "not-validated", "  # not a comment", "  https://spaced"}
	if len(links) != len(want) {
		t.Fatalf("links=%v want %v", links, want)
	}
	for i := range want {
		if links[i] != want[i] {
			t.Fatalf("links=%v want %v", links, want)
		}
	}
}

func TestExtractLinkLines_CRLF(t *testing.T) {
	input := "https://a.example\r\n# comment\r\nhttps://b.example\r\n"
	links, err := ExtractLinkLines(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Fatalf("links=%v", links)
	}
	for _, l := range links {
		if strings.Contains(l, "\r") {
			t.Fatalf("CRLF left \\r in URL: %q", l)
		}
	}
}

const bom = "\uFEFF"

func TestExtractLinkLines_StripsLeadingBOM(t *testing.T) {
	links, err := ExtractLinkLines(strings.NewReader(bom + "https://a.example\n  https://kept\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://a.example", "  https://kept"}
	if len(links) != len(want) || links[0] != want[0] || links[1] != want[1] {
		t.Fatalf("links=%q want %q", links, want)
	}
}

func TestExtractLinkLines_BOMBeforeComment(t *testing.T) {
	links, err := ExtractLinkLines(strings.NewReader(bom + "# header\nhttps://a.example\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0] != "https://a.example" {
		t.Fatalf("links=%q", links)
	}
}

func TestParseLinks_StripsLeadingBOM(t *testing.T) {
	r := ParseLinks(strings.NewReader(bom + "https://a.example\r\nhttps://b.example\r\n"))
	if len(r.Errors) != 0 {
		t.Fatalf("errors=%v", r.Errors)
	}
	if len(r.Links) != 2 || r.Links[0] != "https://a.example" {
		t.Fatalf("links=%q", r.Links)
	}
}

func TestBOMMidFileLeftAlone(t *testing.T) {
	input := "https://a.example\n" + bom + "https://b.example\n"
	links, err := ExtractLinkLines(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[1] != bom+"https://b.example" {
		t.Fatalf("CLI links=%q", links)
	}
	r := ParseLinks(strings.NewReader(input))
	if len(r.Links) != 1 || len(r.Errors) != 1 || r.Errors[0].Line != 2 {
		t.Fatalf("GUI links=%q errors=%v", r.Links, r.Errors)
	}
}

func TestShortInputWithoutBOM(t *testing.T) {
	links, err := ExtractLinkLines(strings.NewReader("ab"))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0] != "ab" {
		t.Fatalf("links=%q", links)
	}
}
