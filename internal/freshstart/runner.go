package freshstart

import (
	"fmt"
)

// Opener opens a URL in a browser (or test double).
type Opener func(url string) error

// ProgressEvent is emitted while OpenAll runs.
type ProgressEvent struct {
	Index   int // 0-based index into the link list
	Total   int
	URL     string
	Err     error  // non-nil when this URL failed to open
	Message string // human-readable status for this step
}

// ProgressFunc receives progress updates; may be nil.
type ProgressFunc func(ProgressEvent)

// OpenResult summarizes an OpenAll run.
type OpenResult struct {
	Opened int
	Failed int
	Errors []error
}

// OpenAll opens each URL with opener, reporting progress. It continues after
// individual failures (partial success). An empty list returns a clear error
// without calling opener. A nil opener is an error.
func OpenAll(links []string, opener Opener, progress ProgressFunc) OpenResult {
	var result OpenResult
	emit := func(ev ProgressEvent) {
		if progress != nil {
			progress(ev)
		}
	}

	if opener == nil {
		err := fmt.Errorf("opener is nil")
		result.Failed = len(links)
		result.Errors = append(result.Errors, err)
		emit(ProgressEvent{Total: len(links), Err: err, Message: err.Error()})
		return result
	}

	if len(links) == 0 {
		err := fmt.Errorf("no links to open")
		result.Errors = append(result.Errors, err)
		emit(ProgressEvent{Total: 0, Err: err, Message: err.Error()})
		return result
	}

	total := len(links)
	for i, link := range links {
		err := opener(link)
		ev := ProgressEvent{
			Index: i,
			Total: total,
			URL:   link,
			Err:   err,
		}
		if err != nil {
			result.Failed++
			wrapped := fmt.Errorf("open %q: %w", link, err)
			result.Errors = append(result.Errors, wrapped)
			ev.Message = fmt.Sprintf("failed %d/%d: %v", i+1, total, err)
			ev.Err = wrapped
		} else {
			result.Opened++
			ev.Message = fmt.Sprintf("opened %d/%d: %s", i+1, total, link)
		}
		emit(ev)
	}
	return result
}
