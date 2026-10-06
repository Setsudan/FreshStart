//go:build windows

// FreshStart desktop GUI (Windows beta).
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/Setsudan/FreshStart/internal/freshstart"
	"github.com/pkg/browser"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("FreshStart"), app.Size(unit.Dp(720), unit.Dp(560)))
		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type ui struct {
	theme *material.Theme

	profilesDir widget.Editor
	profileName widget.Editor
	links       widget.Editor
	status      widget.Editor

	btnRefresh widget.Clickable
	btnLoad    widget.Clickable
	btnSave    widget.Clickable
	btnOpen    widget.Clickable

	profileList layout.List
	profileBtns []*widget.Clickable
	profiles    []string

	mu           sync.Mutex
	busy         bool
	statusDirty  bool
	statusLines  []string
	invalidate   func()
}

func newUI() *ui {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	u := &ui{
		theme:       th,
		profileList: layout.List{Axis: layout.Vertical},
	}
	u.profilesDir.SetText(freshstart.DefaultProfilesDir)
	u.profilesDir.SingleLine = true
	u.profileName.SingleLine = true
	u.links.SingleLine = false
	u.status.SingleLine = false
	u.status.ReadOnly = true
	u.appendStatus("Ready. Pick a profile or paste links, then Open All.")
	u.refreshProfiles()
	return u
}

func (u *ui) appendStatus(line string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.statusLines = append(u.statusLines, line)
	if len(u.statusLines) > 200 {
		u.statusLines = u.statusLines[len(u.statusLines)-200:]
	}
	u.statusDirty = true
}

func (u *ui) flushStatus() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if !u.statusDirty {
		return
	}
	u.status.SetText(strings.Join(u.statusLines, "\n"))
	u.statusDirty = false
}

func (u *ui) setBusy(v bool) {
	u.mu.Lock()
	u.busy = v
	u.mu.Unlock()
}

func (u *ui) isBusy() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.busy
}

func (u *ui) refreshProfiles() {
	dir := strings.TrimSpace(u.profilesDir.Text())
	if dir == "" {
		dir = freshstart.DefaultProfilesDir
		u.profilesDir.SetText(dir)
	}
	names, err := freshstart.ListProfiles(dir)
	if err != nil {
		u.profiles = nil
		u.profileBtns = nil
		u.appendStatus(fmt.Sprintf("Profiles: %v", err))
		return
	}
	u.profiles = names
	u.profileBtns = make([]*widget.Clickable, len(names))
	for i := range names {
		u.profileBtns[i] = new(widget.Clickable)
	}
	u.appendStatus(fmt.Sprintf("Loaded %d profile(s) from %s", len(names), dir))
}

func (u *ui) loadProfile(name string) {
	dir := strings.TrimSpace(u.profilesDir.Text())
	result, path, err := freshstart.LoadProfile(dir, name)
	if err != nil {
		u.appendStatus(fmt.Sprintf("Load failed: %v", err))
		return
	}
	u.profileName.SetText(name)
	// Show raw file text so users can edit comments too.
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		u.appendStatus(fmt.Sprintf("Load failed reading %s: %v", path, readErr))
		return
	}
	u.links.SetText(string(raw))
	for _, le := range result.Errors {
		u.appendStatus(fmt.Sprintf("Parse warning: %v", le))
	}
	u.appendStatus(fmt.Sprintf("Loaded %s (%d valid link(s))", path, len(result.Links)))
}

func (u *ui) saveProfile() {
	dir := strings.TrimSpace(u.profilesDir.Text())
	name := strings.TrimSpace(u.profileName.Text())
	if name == "" {
		u.appendStatus("Save failed: profile name is empty")
		return
	}
	path, err := freshstart.SaveProfile(dir, name, u.links.Text())
	if err != nil {
		u.appendStatus(fmt.Sprintf("Save failed: %v", err))
		return
	}
	u.appendStatus(fmt.Sprintf("Saved %s", path))
	u.refreshProfiles()
}

func (u *ui) openAll() {
	if u.isBusy() {
		u.appendStatus("Already opening links…")
		return
	}
	parsed := freshstart.ParseLinks(strings.NewReader(u.links.Text()))
	for _, le := range parsed.Errors {
		u.appendStatus(fmt.Sprintf("Skipped: %v", le))
	}
	if len(parsed.Links) == 0 {
		u.appendStatus("Open failed: no valid links (empty list or only comments/invalid lines)")
		return
	}

	u.setBusy(true)
	u.appendStatus(fmt.Sprintf("Opening %d link(s)…", len(parsed.Links)))
	links := append([]string(nil), parsed.Links...)
	go func() {
		defer u.setBusy(false)
		res := freshstart.OpenAll(links, func(url string) error {
			return browser.OpenURL(url)
		}, func(ev freshstart.ProgressEvent) {
			u.appendStatus(ev.Message)
			if u.invalidate != nil {
				u.invalidate()
			}
		})
		u.appendStatus(fmt.Sprintf("Done: opened %d, failed %d", res.Opened, res.Failed))
		if u.invalidate != nil {
			u.invalidate()
		}
	}()
}

func run(w *app.Window) error {
	u := newUI()
	u.invalidate = w.Invalidate
	var ops op.Ops

	for {
		e := w.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			u.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (u *ui) layout(gtx layout.Context) layout.Dimensions {
	u.flushStatus()
	if u.btnRefresh.Clicked(gtx) {
		u.refreshProfiles()
	}
	if u.btnLoad.Clicked(gtx) {
		name := strings.TrimSpace(u.profileName.Text())
		if name == "" {
			u.appendStatus("Load failed: enter or select a profile name")
		} else {
			u.loadProfile(name)
		}
	}
	if u.btnSave.Clicked(gtx) {
		u.saveProfile()
	}
	if u.btnOpen.Clicked(gtx) {
		u.openAll()
	}
	for i, btn := range u.profileBtns {
		if btn.Clicked(gtx) {
			u.profileName.SetText(u.profiles[i])
			u.loadProfile(u.profiles[i])
		}
	}

	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.H5(u.theme, "FreshStart").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Body1(u.theme, "Profiles dir:").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return material.Editor(u.theme, &u.profilesDir, "profiles").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Button(u.theme, &u.btnRefresh, "Refresh").Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Body1(u.theme, "Profile:").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return material.Editor(u.theme, &u.profileName, "name.txt").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Button(u.theme, &u.btnLoad, "Load").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Button(u.theme, &u.btnSave, "Save").Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(100))
				return u.profileList.Layout(gtx, len(u.profiles), func(gtx layout.Context, i int) layout.Dimensions {
					return material.Button(u.theme, u.profileBtns[i], u.profiles[i]).Layout(gtx)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(u.theme, "Link list (one URL per line; # comments allowed):").Layout(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return material.Editor(u.theme, &u.links, "https://example.com").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(u.theme, &u.btnOpen, "Open All")
				if u.isBusy() {
					btn.Background = u.theme.Fg
				}
				return btn.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(u.theme, "Status:").Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(120))
				gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(140))
				return material.Editor(u.theme, &u.status, "").Layout(gtx)
			}),
		)
	})
}
