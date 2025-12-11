# Vigilant - AI Assistant Development Guide

This document provides guidance for Claude Code (or other AI assistants) when working with the Vigilant project.

## Quick Start for AI Assistants

### Project Overview
- **Type**: Cross-platform desktop application (Go + Wails + Svelte)
- **Purpose**: Productivity monitoring with humorous FBI meme deterrent
- **Size**: ~2000 lines of backend, ~500 lines frontend
- **Platforms**: Windows 10+, macOS 12+
- **Build System**: Makefile + Wails + npm

### Getting Started
1. Read PRD.md for complete specification
2. Review TASKS.md for implementation details
3. Check the file structure (see below)
4. Read relevant package documentation via `go doc`

## Technology Stack

### Backend (Go 1.21+)
- **Framework**: Wails v2 (desktop app wrapper)
- **Config**: Viper (YAML configuration)
- **Platform APIs**:
  - Windows: Win32 API (golang.org/x/sys/windows)
  - macOS: CGO + Cocoa framework
- **Concurrency**: Context, channels, sync.RWMutex
- **Testing**: stdlib testing package

### Frontend (Svelte + Node.js 18+)
- **Framework**: Svelte 4
- **Styling**: Tailwind CSS v4
- **Build**: Vite
- **State**: Svelte stores (writable)
- **Type Safety**: TypeScript

## Project Structure

```
vigilant/
├── cmd/vigilant/main.go            # Wails app entry point
├── internal/                        # All business logic
│   ├── app/app.go                  # Central orchestrator
│   ├── config/config.go            # Configuration loading
│   ├── monitor/                    # Window monitoring
│   │   ├── monitor.go              # Interface & types
│   │   ├── monitor_windows.go      # Windows impl (build tag)
│   │   ├── monitor_darwin.go       # macOS impl (build tag)
│   │   └── detector.go             # Grace period logic
│   ├── blocker/                    # Blocklist logic
│   │   ├── blocker.go              # Matching rules
│   │   └── state.go                # State management
│   ├── player/player.go            # Player control
│   └── stats/stats.go              # Stats tracking
├── frontend/                       # Svelte UI
│   ├── src/
│   │   ├── App.svelte             # Main layout
│   │   ├── lib/components/        # UI components
│   │   ├── stores/app.ts          # State management
│   │   └── types/index.ts         # TypeScript types
│   └── package.json
├── config/default.yaml            # Default config
├── assets/fbi-meme.mp4           # Embedded video
├── wails.json                     # Wails config
├── go.mod / go.sum               # Go dependencies
├── Makefile                       # Build automation
├── README.md                      # User documentation
└── PRD.md / TASKS.md             # Project docs
```

## Key Architectural Patterns

### Interface-Based Design
All major components expose interfaces (Monitor, Blocker, StateManager). Platform-specific code is isolated in separate files with build tags:
- `//go:build windows` - Windows only
- `//go:build darwin` - macOS only

### Event-Driven Architecture
Components communicate via channels and events, not direct calls:
```
Monitor.Events() → Blocker → StateManager → PlayerController → Frontend
```

### Thread Safety
- Use `sync.RWMutex` for shared state access
- Channel-based communication for goroutine coordination
- All public methods are goroutine-safe

### Configuration Over Code
All behavior is configurable via YAML:
- Blocklist rules
- Monitor intervals
- Player settings
- UI preferences

## Common Tasks for AI Assistants

### Adding a New Blocking Rule

1. Update `config/default.yaml` with new process/website/pattern
2. Test with `go test ./internal/blocker`
3. Verify in TestIsBlocked_* test cases

### Modifying Monitor Behavior

1. Change polling interval: Update `config.Monitor.PollInterval`
2. Change grace period: Update `config.Monitor.GracePeriod`
3. Test with detector unit tests

### Adding UI Component

1. Create `.svelte` file in `frontend/src/lib/components/`
2. Import component in parent
3. Update Wails bindings if needs Go data
4. Test with `make dev` (hot reload)

### Debugging

**Backend**:
```bash
# Run with logging
VIGILANT_LOG=debug make dev

# Test specific package
go test ./internal/monitor -v

# Check race conditions
go test ./... -race
```

**Frontend**:
- Use browser DevTools (F12 in app window)
- Check Svelte component state in DevTools
- Monitor Wails.invoke() calls in Network tab

## Wails-Specific Details

### App Struct Public Methods
These are exposed to TypeScript via `WithBind()`:
- `GetStats() Stats`
- `GetPlayerState() string`
- `GetFocusState() string`
- `SetVolume(float64)`
- `ManualTrigger()`

### Events to Frontend
Events emitted via `runtime.EventsEmit()`:
- `player:state-change` - Player switched (lofi/fbi)
- `stats:update` - Stats changed
- `focus:state-change` - Focus state changed
- `window:focus-change` - Window focus event
- `app:error` - Error occurred

### Frontend Event Listening
```svelte
<script>
  import { EventsOn } from 'wailsjs/runtime';

  onMount(() => {
    EventsOn('player:state-change', (newState) => {
      // Handle state change
    });
  });
</script>
```

## Testing Strategy

### Unit Tests
- Each package has corresponding `_test.go` file
- Target >90% code coverage
- Run `go test ./...` before committing
- Check race conditions with `-race` flag

### Integration Tests
- App orchestrator wires components together
- Test event flow end-to-end
- Verify all packages compile together

### Manual Testing
1. `make dev` - Development mode with hot reload
2. Test blocklist with actual Discord/YouTube
3. Verify YouTube player loads
4. Check FBI video triggers correctly
5. Verify stats accumulate

## Common Patterns

### Getting Stats in Frontend
```svelte
<script>
  async function updateStats() {
    const stats = await invoke('GetStats');
    console.log(stats);
  }
</script>
```

### Configuration Access in Backend
```go
// In any component with access to app.config
domain := app.config.Player.LofiPlaylist
pollInterval := app.config.Monitor.PollInterval
```

### Emitting Events from Backend
```go
// In app.go or any handler
runtime.EventsEmit(ctx, "event:name", data)
```

## Frequently Asked Questions

**Q: How do I add a new blocker type (besides process/website/regex)?**
A: Update `BlocklistConfig` in config.go, add matching logic to blocker.go, add tests.

**Q: How do I change the default FBI meme video?**
A: Replace assets/fbi-meme.mp4 with new video, keep filename same.

**Q: How do I cross-compile Windows from macOS?**
A: Use `xgo` Docker tool (documented in Makefile comments) or build on Windows directly.

**Q: Where's the database?**
A: MVP uses in-memory stats only. Database planned for Phase 3.

**Q: How do I add a new configuration option?**
A: Add struct field to Config, add YAML tag, update default.yaml, add validation.

## Development Workflow

1. **Make code changes**
2. **Run tests**: `make test`
3. **Test manually**: `make dev` or `./build/bin/vigilant*`
4. **Check race conditions**: `go test ./... -race`
5. **Verify coverage**: `go test ./... -cover`
6. **Build**: `make build-darwin` (or appropriate target)
7. **Commit** with clear messages

## Performance Tips

- Monitor uses 100ms polling to balance responsiveness and CPU
- Stats tracker updates every 1 second (not every poll)
- Regex patterns compiled once at app start, not on each check
- Frontend debounces rapid store updates

## Security Considerations

- No telemetry or data collection
- No network calls except YouTube API
- Local file-based configuration only
- No elevated privileges required
- Standard OS window permissions only

## Troubleshooting for Developers

**App crashes on start**:
- Check config file is valid YAML
- Verify YouTube URL is accessible
- Check file permissions on assets/

**Tests fail intermittently**:
- Timing-sensitive tests may need adjustment
- Check system clock is correct
- Run with `-race` to find race conditions

**Wails dev mode errors**:
- Clear frontend cache: `cd frontend && rm -rf node_modules dist && npm install`
- Kill any existing Wails processes: `pkill vigilant`
- Check port 5173 is available (frontend dev server)

**Build fails**:
- Ensure all dependencies installed: `go mod download`
- Check Go version: `go version` (need 1.21+)
- Windows: CGO may fail; ensure gcc is in PATH

## Resources

- [Wails Documentation](https://wails.io/docs)
- [Go Platform-Specific Code](https://golang.org/doc/install/source#environment)
- [Svelte Documentation](https://svelte.dev/docs)
- [Tailwind CSS v4](https://tailwindcss.com/docs)

## Contact

- GitHub: https://github.com/mgiovani/vigilant
- Issues: https://github.com/mgiovani/vigilant/issues

---

**Last Updated**: 2025-12-10
**Status**: Production Ready
