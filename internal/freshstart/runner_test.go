package freshstart

import (
	"errors"
	"testing"
)

func TestOpenAll_EmptyList(t *testing.T) {
	called := false
	res := OpenAll(nil, func(string) error {
		called = true
		return nil
	}, nil)
	if called {
		t.Fatal("opener should not be called")
	}
	if res.Opened != 0 || len(res.Errors) != 1 {
		t.Fatalf("result=%+v", res)
	}
}

func TestOpenAll_NilOpener(t *testing.T) {
	res := OpenAll([]string{"https://a"}, nil, nil)
	if res.Failed != 1 || len(res.Errors) != 1 {
		t.Fatalf("result=%+v", res)
	}
}

func TestOpenAll_ProgressAndPartialFailure(t *testing.T) {
	var events []ProgressEvent
	boom := errors.New("browser busy")
	res := OpenAll(
		[]string{"https://ok.example", "https://bad.example", "https://ok2.example"},
		func(u string) error {
			if u == "https://bad.example" {
				return boom
			}
			return nil
		},
		func(ev ProgressEvent) { events = append(events, ev) },
	)
	if res.Opened != 2 || res.Failed != 1 {
		t.Fatalf("result=%+v", res)
	}
	if len(events) != 3 {
		t.Fatalf("events=%d", len(events))
	}
	if events[0].Err != nil || events[1].Err == nil || events[2].Err != nil {
		t.Fatalf("event errs: %v %v %v", events[0].Err, events[1].Err, events[2].Err)
	}
	if events[1].Index != 1 || events[1].Total != 3 {
		t.Fatalf("bad event: %+v", events[1])
	}
}

func TestOpenAll_AllSuccess(t *testing.T) {
	var opened []string
	res := OpenAll([]string{"https://a", "https://b"}, func(u string) error {
		opened = append(opened, u)
		return nil
	}, nil)
	if res.Opened != 2 || res.Failed != 0 || len(res.Errors) != 0 {
		t.Fatalf("result=%+v", res)
	}
	if len(opened) != 2 {
		t.Fatalf("opened=%v", opened)
	}
}
