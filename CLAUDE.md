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
- **Frontend**: Svelte 3 (JavaScript components, TypeScript for stores/types), Tailwind CSS v4, Vite
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
| `internal/app` | Central orchestrator (`App` struct), lifecycle management |
| `internal/config` | YAML configuration loading/validation |
| `internal/monitor` | Active window detection (platform-specific via build tags) |
| `internal/blocker` | Regex-based blocklist matching, `StateManager` for focus state |
| `internal/player` | `PlayerController` state machine (StateLofi/StateFBI) |
| `internal/stats` | `Tracker` with background goroutine for time accumulation |
| `internal/youtube` | YouTube embed URL proxy |

### Wails App Structure

**`main.go`** defines `VanillaApp` - a wrapper that exposes methods to frontend:
- Delegates to `internal/app.App` orchestrator for business logic
- Public methods auto-bound to TypeScript via Wails

**Key bindings:**
- `GetStats()`, `GetPlayerState()`, `GetFocusState()`, `GetConfig()`
- `GetBlocklist()`, `AddBlocklistEntry()`, `RemoveBlocklistEntry()`
- `GetLofiEmbedURL()`, `SetVolume()`, `ManualTrigger()`

**Events emitted via `runtime.EventsEmit()`:**
- `player:state-change` - player switched lofi/fbi
- `stats:update` - stats updated (every second)
- `focus:state-change` - working/distracted transition
- `window:focus-change` - active window changed
- `trigger:confirmed` - FBI triggered after grace period
- `app:error` - backend error occurred

### Platform-Specific Code
Files use build tags for platform isolation:
- `monitor_darwin.go` - macOS (CGO + Cocoa) - `//go:build darwin`
- `monitor_windows.go` - Windows (Win32 API) - `//go:build windows`

### Thread Safety
- All shared state protected by `sync.RWMutex`
- Channel-based communication between goroutines
- Context for lifecycle management and cancellation

## Frontend Structure

```
frontend/src/
├── App.svelte                    # Main layout, event listener init
├── lib/components/
│   ├── LofiPlayer.svelte        # YouTube iframe embed
│   ├── FBIVideo.svelte          # Bundled FBI meme player
│   ├── StatsDisplay.svelte      # Glassmorphism stats overlay
│   ├── SettingsPanel.svelte     # Slide-in settings panel
│   └── BlocklistEditor.svelte   # Add/remove blocklist entries
├── stores/app.ts                # Svelte stores + event listeners
└── types/index.ts               # TypeScript interfaces
```

Frontend uses `EventsOn()` from `wailsjs/runtime` to subscribe to backend events.
Wails bindings auto-generated in `frontend/wailsjs/go/main/VanillaApp`.

## Configuration

Config file: `~/.vigilant/config.yaml` (user) or `config/default.yaml` (bundled)

Key sections:
- `blocklist.patterns` - Regex patterns (case-insensitive, match title AND process)
- `exceptions` - Whitelist patterns that bypass blocklist
- `monitor.poll_interval` / `monitor.grace_period` - Timing (default 100ms/500ms)
- `player.lofi_playlist` / `player.default_volume` - Player settings

**Blocklist matching strategy:**
1. Exception patterns checked first (whitelist) - if match, not blocked
2. Block patterns checked on both window title AND process name
3. All patterns use `(?i)` flag for case-insensitive matching

## Common Tasks

### Adding a blocking rule
1. Edit `config/default.yaml` under `blocklist.patterns`
2. Test: `go test ./internal/blocker -v`

### Exposing a new Go function to frontend
1. Add public method to `VanillaApp` in `main.go`
2. Delegate to orchestrator (`a.app`) if needed
3. Bindings auto-regenerate on `wails dev`
4. Import from `wailsjs/go/main/VanillaApp` in frontend

### Adding a new backend event
1. Emit from `internal/app/app.go`: `runtime.EventsEmit(a.wailsCtx, "event:name", payload)`
2. Listen in `frontend/src/stores/app.ts` using `EventsOn("event:name", handler)`

### Adding a UI component
1. Create `.svelte` file in `frontend/src/lib/components/`
2. Import in parent component
3. Note: Svelte components use JavaScript, not TypeScript

### Debugging
```bash
VIGILANT_LOG=debug make dev              # backend logging
# F12 in app window opens DevTools       # frontend debugging
go test ./... -race                      # race condition detection
```

## Troubleshooting

- **Wails dev errors**: `cd frontend && rm -rf node_modules dist && npm install`
- **Port conflicts**: `pkill vigilant` or check `lsof -i :5173`
- **Build fails**: Ensure `go mod download` completed, check Go version 1.23+
- **macOS permissions**: System Preferences → Security & Privacy → Accessibility

## Further Documentation

For deeper understanding, see the `docs/` folder:
- `docs/architecture.md` - Detailed system architecture with Mermaid diagrams
- `docs/onboarding.md` - Developer setup guide and code walkthrough
- `docs/contributing.md` - Contribution guidelines
- `docs/deployment.md` - Build and release process
