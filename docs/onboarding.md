# Vigilant - Developer Onboarding Guide

Welcome to Vigilant! This guide will help you get up and running with the codebase.

## Prerequisites

### Required Tools

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| Go | 1.23+ | Backend language | [golang.org/dl](https://golang.org/dl/) |
| Node.js | 18+ | Frontend build | [nodejs.org](https://nodejs.org/) |
| Wails CLI | v2.x | Desktop app framework | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| npm | 9+ | Package manager | Included with Node.js |

### Platform-Specific Requirements

#### macOS
- Xcode Command Line Tools: `xcode-select --install`
- macOS 12 Monterey or later

#### Windows
- Windows 10 or later
- Visual Studio Build Tools (for CGO if needed)
- WebView2 Runtime (usually pre-installed on Windows 10+)

## Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/mgiovani/vigilant.git
cd vigilant

# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend && npm install && cd ..
```

### 2. Environment Configuration

```bash
# Copy environment template (if exists)
cp .env.example .env

# Or create minimal .env
echo "YOUTUBE_API_KEY=your_api_key_here" > .env
```

### 3. Run Development Server

```bash
# Start Wails dev server with hot reload
make dev
# or
wails dev
```

This will:
- Compile the Go backend
- Start the Vite dev server for frontend
- Open the application window
- Enable hot reload for frontend changes

### 4. Run Tests

```bash
# Run all Go tests
make test
# or
go test ./...

# Run tests with race detection
go test ./... -race

# Run tests with coverage
go test ./... -cover
```

## Project Structure

```
vigilant/
├── main.go                      # Wails application entry point (root directory)
├── internal/                    # Private application code
│   ├── app/                     # Central orchestrator
│   │   ├── app.go              # Main App struct and lifecycle
│   │   └── app_test.go         # Integration tests
│   ├── config/                  # Configuration loading
│   │   ├── config.go           # Config structs and types
│   │   ├── loader.go           # YAML loading logic
│   │   ├── validator.go        # Config validation
│   │   └── config_test.go      # Config tests
│   ├── monitor/                 # Window monitoring
│   │   ├── monitor.go          # Interface and types
│   │   ├── monitor_darwin.go   # macOS implementation
│   │   ├── monitor_windows.go  # Windows implementation
│   │   ├── detector.go         # Grace period logic
│   │   └── detector_test.go    # Detector tests
│   ├── blocker/                 # Blocklist logic
│   │   ├── blocker.go          # Matching rules
│   │   ├── blocker_test.go     # Blocker tests
│   │   ├── state.go            # State management
│   │   └── state_test.go       # State tests
│   ├── player/                  # Player control
│   │   ├── player.go           # Player state machine
│   │   └── player_test.go      # Player tests
│   ├── stats/                   # Statistics tracking
│   │   ├── stats.go            # Stats accumulation
│   │   └── stats_test.go       # Stats tests
│   └── youtube/                 # YouTube API integration
│       ├── api.go              # YouTube Data API
│       └── embedproxy.go       # Embed URL handling
├── frontend/                    # Svelte frontend
│   ├── src/
│   │   ├── App.svelte          # Main application component
│   │   ├── main.js             # Frontend entry point
│   │   ├── app.css             # Application styles
│   │   ├── style.css           # Global styles
│   │   ├── vite-env.d.ts       # Vite type definitions
│   │   ├── assets/             # Static assets
│   │   ├── lib/
│   │   │   └── components/     # Reusable UI components
│   │   │       ├── BlocklistEditor.svelte
│   │   │       ├── FBIVideo.svelte
│   │   │       ├── LofiPlayer.svelte
│   │   │       ├── SettingsPanel.svelte
│   │   │       └── StatsDisplay.svelte
│   │   ├── stores/             # Svelte stores (TypeScript)
│   │   │   └── app.ts          # Main application store
│   │   └── types/              # TypeScript type definitions
│   │       └── index.ts        # Shared types
│   ├── wailsjs/                # Auto-generated Wails bindings
│   └── package.json            # Frontend dependencies
├── config/
│   └── default.yaml            # Default configuration
├── assets/
│   └── fbi-meme.mp4            # Bundled FBI video
├── build/                       # Build outputs
├── Makefile                     # Build automation
├── wails.json                   # Wails configuration
├── go.mod                       # Go module definition
├── PRD.md                       # Product requirements
├── TASKS.md                     # Implementation tasks
└── CLAUDE.md                    # AI assistant guide
```

## Key Concepts

### 1. Wails Architecture

Vigilant uses Wails v2, which provides:
- Go backend with native performance
- WebView2 (Windows) / WebKit (macOS) frontend
- Automatic TypeScript bindings generation
- Event system for Go ↔ Frontend communication

**Public Methods Exposed to Frontend:**

The `VanillaApp` struct in `main.go` wraps the orchestrator and exposes methods to the frontend. All public methods are automatically bound via Wails:

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetStats` | `() interface{}` | Returns current statistics snapshot |
| `GetPlayerState` | `() string` | Returns current player state ("lofi" or "fbi") |
| `GetFocusState` | `() string` | Returns current focus state ("working" or "distracted") |
| `GetConfig` | `() interface{}` | Returns application configuration |
| `GetBlocklist` | `() interface{}` | Returns current blocklist configuration |
| `GetLofiVideoID` | `() string` | Returns YouTube video ID for lofi music |
| `GetLofiEmbedURL` | `() string` | Returns YouTube embed URL for lofi player |
| `SetVolume` | `(level float64)` | Sets player volume (0.0-1.0) |
| `ManualTrigger` | `()` | Manually triggers FBI meme (for testing) |
| `AddBlocklistEntry` | `(category, value string)` | Adds a new blocklist entry |
| `RemoveBlocklistEntry` | `(category, value string)` | Removes a blocklist entry |
| `SaveConfig` | `()` | Saves current configuration to disk |

**Frontend Usage:**
```typescript
// Import auto-generated bindings
import { GetStats, GetPlayerState, ManualTrigger } from '../wailsjs/go/main/VanillaApp';

// Call backend methods
const stats = await GetStats();
const playerState = await GetPlayerState();
await ManualTrigger();
```

### 2. Event-Driven Communication

Backend emits events to frontend:
```go
runtime.EventsEmit(wailsCtx, "player:state-change", map[string]interface{}{
    "state":     state,
    "timestamp": time.Now().Unix(),
})
```

Frontend subscribes to events:
```svelte
<script>
  import { EventsOn } from '../../wailsjs/runtime/runtime';

  onMount(() => {
    EventsOn('player:state-change', (data) => {
      console.log(data.state); // "lofi" or "fbi"
    });
  });
</script>
```

**Available Events:**

| Event Name | Payload | Description |
|------------|---------|-------------|
| `player:state-change` | `{ state: string, timestamp: number }` | Player switched between lofi/fbi |
| `focus:state-change` | `{ state: string, window: string, process: string, timestamp: number }` | Focus state changed (working/distracted) |
| `stats:update` | `{ focusedTime: string, distractedTime: string, triggerCount: number, focusRate: number, sessionStart: number, lastUpdate: number }` | Stats updated (every second) |
| `window:focus-change` | `{ title: string, process: string, pid: number, blocked: boolean, timestamp: number }` | Active window changed |
| `trigger:confirmed` | `{ window: string, process: string, timestamp: number }` | FBI meme triggered after grace period |
| `trigger:manual` | `{ timestamp: number }` | Manual trigger activated |
| `player:volume-change` | `number` | Volume level changed (0.0-1.0) |
| `app:error` | `{ message: string, timestamp: number }` | Error occurred |

### 3. Platform Abstraction

Platform-specific code uses build tags:

```go
// monitor_darwin.go
//go:build darwin

package monitor

// macOS-specific implementation
```

```go
// monitor_windows.go
//go:build windows

package monitor

// Windows-specific implementation
```

### 4. Configuration System

Configuration is loaded from YAML:
- Default: `config/default.yaml` (bundled)
- User override: `~/.vigilant/config.yaml`

```yaml
blocklist:
  processes:
    - "Discord"
    - "Steam"
  websites:
    - "reddit.com"
    - "twitter.com"
```

## Common Development Tasks

### Adding a New Blocked Process

1. Edit `config/default.yaml`:
```yaml
blocklist:
  processes:
    - "NewApp"  # Add here
```

2. Test the change:
```bash
go test ./internal/blocker -v
```

### Adding a New Frontend Component

1. Create component in `frontend/src/lib/components/`:
```svelte
<!-- MyComponent.svelte -->
<script>
  export let prop = '';
</script>

<div class="my-component">
  {prop}
</div>
```

2. Import in parent component:
```svelte
<script>
  import MyComponent from './lib/components/MyComponent.svelte';
</script>

<MyComponent prop="value" />
```

**Note:** The project uses JavaScript for Svelte components. TypeScript is used only in stores (`app.ts`) and type definitions (`types/index.ts`).

### Exposing a New Go Function to Frontend

1. Add public method to `VanillaApp` struct in `main.go`:
```go
func (a *VanillaApp) MyNewFunction(arg string) string {
    // You can delegate to the orchestrator if needed
    return "result"
}
```

2. Regenerate bindings (happens automatically on `wails dev`)

3. Use in frontend:
```typescript
import { MyNewFunction } from '../wailsjs/go/main/VanillaApp';
const result = await MyNewFunction("argument");
```

**Note:** The `VanillaApp` in `main.go` is a wrapper around the orchestrator `App` from `internal/app/app.go`. Add methods to `VanillaApp` to expose them to the frontend, and delegate to the orchestrator as needed.

### Adding a New Event

1. Emit from backend in `internal/app/app.go`:
```go
runtime.EventsEmit(a.wailsCtx, "my:new-event", map[string]interface{}{
    "key": "value",
})
```

2. Listen in frontend (in `frontend/src/stores/app.ts` or component):
```typescript
import { EventsOn } from '../../wailsjs/runtime/runtime';

EventsOn('my:new-event', (data: { key: string }) => {
    console.log(data.key);
});
```

## Testing Strategy

### Unit Tests
Each package has corresponding `_test.go` files:
```bash
# Test specific package
go test ./internal/blocker -v

# Test with coverage
go test ./internal/blocker -cover
```

### Integration Tests
App orchestrator tests in `internal/app/app_test.go`:
```bash
go test ./internal/app -v
```

### Manual Testing
1. Start dev server: `make dev`
2. Open a blocked app (Discord, YouTube)
3. Verify FBI meme triggers after grace period
4. Switch to safe app (IDE)
5. Verify lofi music resumes

## Debugging Tips

### Backend Logging
Logs are prefixed with `[vigilant]`:
```go
a.logger.Printf("Debug message: %v", value)
```

### Frontend DevTools
Press `F12` or `Cmd+Option+I` in the app window to open DevTools.

### Race Condition Detection
```bash
go test ./... -race
```

### Wails Dev Issues
If dev mode fails:
```bash
# Clear frontend cache
cd frontend && rm -rf node_modules dist && npm install && cd ..

# Kill existing processes
pkill vigilant

# Check port 5173 is available
lsof -i :5173
```

## Build Commands

```bash
# Development (with hot reload)
make dev

# Production build (auto-detect OS)
make build

# Platform-specific builds
make build-darwin           # macOS amd64
make build-darwin-universal # macOS universal (Intel + Apple Silicon)
make build-windows          # Windows amd64

# Install to system
make install

# Run tests
make test

# Clean build artifacts
make clean
```

## Getting Help

- **README.md**: User-facing documentation
- **CLAUDE.md**: AI assistant development guide
- **PRD.md**: Product requirements and specifications
- **TASKS.md**: Detailed implementation tasks
- **docs/architecture.md**: System architecture overview

## Next Steps

1. Read `PRD.md` for full product requirements
2. Review `docs/architecture.md` for system design
3. Check `TASKS.md` for implementation details
4. Run `make dev` and explore the codebase
5. Pick a task and start contributing!

---

**Last Updated**: 2025-12-11
