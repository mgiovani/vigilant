# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Vigilant is a cross-platform desktop application (Windows 10+, macOS 12+) that helps users stay focused by monitoring their active window and triggering a humorous FBI meme video when accessing blocked apps/websites. Built with Go + Wails v2 + Svelte 3.

## Development Commands

```bash
# Development with hot reload
make dev

# Run tests
make test
go test ./internal/blocker -v      # specific package
go test ./... -race                 # check race conditions
go test ./... -cover                # coverage report

# Build
make build                          # auto-detect OS
make build-darwin                   # macOS Intel
make build-darwin-universal         # macOS Universal (Intel + ARM)
make build-windows                  # Windows (requires Windows env)

# Install to system
make install                        # /Applications (macOS) or Program Files (Windows)

# Clean
make clean
```

## Technology Stack

- **Backend**: Go 1.23+, Wails v2
- **Frontend**: Svelte 3, Tailwind CSS v4, Vite
- **Config**: YAML (gopkg.in/yaml.v3)
- **Platform APIs**: Win32 API (Windows), CGO + Cocoa (macOS)

## Architecture

### Event Flow
```
Monitor (100ms poll) → Detector (grace period) → Blocker → StateManager → PlayerController → Frontend
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `internal/app` | Central orchestrator, Wails bindings |
| `internal/config` | YAML configuration loading/validation |
| `internal/monitor` | Active window detection (platform-specific) |
| `internal/blocker` | Blocklist matching, state management |
| `internal/player` | Lofi/FBI player switching |
| `internal/stats` | Productivity metrics tracking |
| `internal/youtube` | YouTube API integration, embed proxy |

### Platform-Specific Code
Files use build tags for platform isolation:
- `monitor_darwin.go` - macOS (CGO + Cocoa)
- `monitor_windows.go` - Windows (Win32 API)

### Wails Bindings
Public methods on the `App` struct are exposed to TypeScript:
- `GetStats()`, `GetPlayerState()`, `GetFocusState()`
- `SetVolume()`, `ManualTrigger()`

Events emitted via `runtime.EventsEmit()`:
- `player:state-change`, `stats:update`, `focus:state-change`, `app:error`

## Frontend Structure

```
frontend/src/
├── App.svelte                    # Main layout
├── lib/components/
│   ├── LofiPlayer.svelte        # YouTube embed player
│   ├── FBIVideo.svelte          # FBI meme player
│   ├── StatsDisplay.svelte      # Stats overlay
│   ├── SettingsPanel.svelte     # Settings UI
│   └── BlocklistEditor.svelte   # Blocklist config
├── stores/app.ts                # Svelte stores (state)
└── types/index.ts               # TypeScript types
```

Frontend listens to backend events via `EventsOn()` from `wailsjs/runtime`.

## Configuration

Config file: `~/.vigilant/config.yaml` (user) or `config/default.yaml` (bundled)

Key sections:
- `blocklist.processes` - App names to block (case-insensitive)
- `blocklist.websites` - Domains matched in window titles
- `blocklist.patterns` - Regex patterns
- `exceptions` - Allow specific windows despite matching blocklist
- `monitor.poll_interval` / `monitor.grace_period` - Timing settings
- `player.lofi_playlist` / `player.default_volume` - Player config

## Common Tasks

### Adding a blocking rule
1. Edit `config/default.yaml`
2. Test: `go test ./internal/blocker -v`

### Adding a UI component
1. Create `.svelte` file in `frontend/src/lib/components/`
2. Import in parent component
3. Add Wails binding if needs Go data

### Debugging
```bash
VIGILANT_LOG=debug make dev              # backend logging
# Use F12 DevTools in app window         # frontend debugging
```

## Troubleshooting

- **Wails dev errors**: `cd frontend && rm -rf node_modules dist && npm install`
- **Port conflicts**: Kill existing processes: `pkill vigilant`
- **Build fails**: Ensure `go mod download` completed, check Go version
- **macOS permissions**: Grant Accessibility access in System Preferences
