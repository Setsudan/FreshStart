package freshstart

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// LineError describes a problem on a specific line of a link list.
type LineError struct {
	Line int
	Text string
	Err  error
}

func (e LineError) Error() string {
	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
}

// ParseResult holds valid links and any line-level problems found while parsing.
type ParseResult struct {
	Links  []string
	Errors []LineError
}

// ParseLinks reads a link list from r. Blank lines and lines starting with '#'
// (after trimming leading space) are skipped. Other lines are trimmed; each must
// be an absolute http or https URL to be accepted as a link.
func ParseLinks(r io.Reader) ParseResult {
	var result ParseResult
	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if err := ValidateURL(trimmed); err != nil {
			result.Errors = append(result.Errors, LineError{
				Line: lineNo,
				Text: trimmed,
				Err:  err,
			})
			continue
		}
		result.Links = append(result.Links, trimmed)
	}
	if err := scanner.Err(); err != nil {
		result.Errors = append(result.Errors, LineError{
			Line: lineNo,
			Err:  fmt.Errorf("read error: %w", err),
		})
	}
	return result
}

// ParseLinksFile opens path and parses it with ParseLinks.
// A missing or unreadable file returns an error and an empty ParseResult.
func ParseLinksFile(path string) (ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return ParseResult{}, fmt.Errorf("open link file: %w", err)
	}
	defer f.Close()
	return ParseLinks(f), nil
}

// ValidateURL reports whether s is an absolute http or https URL.
func ValidateURL(s string) error {
	if s == "" {
		return fmt.Errorf("empty URL")
	}
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	if u.Host == "" {
		return fmt.Errorf("URL missing host")
	}
	return nil
}

// FormatLinks writes links as one URL per line, suitable for saving a profile.
func FormatLinks(links []string) string {
	if len(links) == 0 {
		return ""
	}
	var b strings.Builder
	for _, link := range links {
		b.WriteString(link)
		b.WriteByte('\n')
	}
	return b.String()
}

// ExtractLinkLines skips blank lines and '#' comments (after trim) and returns
// the remaining trimmed lines without URL validation. Matches classic CLI behavior.
func ExtractLinkLines(r io.Reader) ([]string, error) {
	var links []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		links = append(links, trimmed)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return links, nil
}
