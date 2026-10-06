# Go FreshStart

This is a Go tool that opens all links from a text file (for example `links.txt`) in your browser. Useful on a fresh Windows install.

The CLI waits for user input after the first link, then opens the rest.

There is a **beta desktop GUI** on the `feat/desktop-gui` branch (Windows).

## links.txt format

```txt
https://www.mozilla.org/fr/firefox/download/thanks/
https://www.google.fr/chrome/
https://discord.com/api/downloads/distributions/app/installers/latest?channel=stable&platform=win&arch=x86
https://github.com/git-for-windows/git/releases/download/v2.35.0.windows.1/Git-2.35.0-64-bit.exe
https://cdn.akamai.steamstatic.com/client/installer/SteamSetup.exe
```

Blank lines and lines starting with `#` are ignored. Profiles under `profiles/` use the same format.

## Building the CLI

Requires Go 1.24.4 or newer (see `go.mod`).

```bash
# Linux / macOS / Windows (console)
go build -o freshstart .

# Cross-compile Windows CLI from Linux
GOOS=windows GOARCH=amd64 go build -o freshstart.exe .
```

Run the CLI and enter the path to a `.txt` link list (for example `profiles/dev.txt`).

## Building the Windows GUI (beta)

The GUI lives in `cmd/freshstart-gui` and is tagged for Windows only (`//go:build windows`). It uses [Gio](https://gioui.org) (pure Go on Windows, no cgo / no MinGW).

From Windows, or cross-compiling from Linux:

```bash
# Recommended: Windows GUI subsystem (no console window)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o freshstart-gui.exe ./cmd/freshstart-gui

# Optional: keep a console for debug logs
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o freshstart-gui.exe ./cmd/freshstart-gui
```

No rsrc/manifest step is required for Gio. Do not commit built `.exe` files; the checked-in `freshstart.exe` is historical and left untouched.

### GUI framework choice

Gio was chosen because it targets Windows with pure Go (no cgo), so `GOOS=windows GOARCH=amd64 go build` works from Linux or macOS without a MinGW toolchain. Fyne was avoided for that reason. The non-UI logic stays in `internal/freshstart` with no GUI imports so Linux `go test` / `go vet` stay headless.

## Testing

```bash
go test ./...
go vet ./...
```

To be honest I prefer using [this](https://github.com/ChrisTitusTech/winutil)
