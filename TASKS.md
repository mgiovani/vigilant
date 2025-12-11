# Vigilant - Implementation Tasks

This document breaks down the implementation of Vigilant into discrete, manageable tasks organized by phase.

---

## Phase 1: Project Setup

### Task 1.1: Initialize Go Module and Wails Project
**Priority**: P0 (Blocker)
**Estimated Time**: 30 minutes

**Subtasks**:
- [ ] Navigate to `/Users/giovani.moutinho/projects/vigilant`
- [ ] Run `go mod init github.com/vigilant-app/vigilant`
- [ ] Install Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [ ] Initialize Wails project: `wails init -n vigilant -t svelte`
- [ ] Create project directory structure:
  ```
  mkdir -p internal/{app,config,monitor,blocker,player,stats}
  mkdir -p assets
  mkdir -p config
  mkdir -p build/{windows,darwin}
  ```
- [ ] Verify Wails dev mode works: `wails dev`

**Acceptance Criteria**:
- go.mod exists with correct module name
- Wails project initialized with Svelte template
- Directory structure matches plan
- `wails dev` launches successfully

---

### Task 1.2: Create Makefile and .gitignore
**Priority**: P0
**Estimated Time**: 15 minutes

**Subtasks**:
- [ ] Create `Makefile` with targets:
  ```makefile
  .PHONY: dev build-windows build-darwin build-darwin-universal clean test

  dev:
      wails dev

  build-windows:
      wails build -platform windows/amd64

  build-darwin:
      wails build -platform darwin/amd64

  build-darwin-universal:
      wails build -platform darwin/universal

  clean:
      rm -rf build/bin

  test:
      go test ./...
  ```
- [ ] Create `.gitignore`:
  ```
  # Binaries
  build/bin/
  *.exe
  *.app

  # Frontend
  frontend/node_modules/
  frontend/dist/

  # Go
  *.log

  # OS
  .DS_Store
  Thumbs.db

  # IDE
  .vscode/
  .idea/
  ```

**Acceptance Criteria**:
- `make dev` launches Wails dev mode
- Build targets work (may fail until implementation, that's OK)
- .gitignore covers common artifacts

---

### Task 1.3: Set Up Configuration System
**Priority**: P0
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Create `internal/config/config.go` with structs:
  - Config
  - BlocklistConfig
  - PlayerConfig
  - MonitorConfig
  - UIConfig
- [ ] Add Viper dependency: `go get github.com/spf13/viper`
- [ ] Implement `LoadConfig(path string) (*Config, error)` function
- [ ] Add validation function `Validate() error` on Config
- [ ] Create `config/default.yaml` with example configuration
- [ ] Write unit tests for config loading

**Acceptance Criteria**:
- Config structs defined with YAML tags
- Viper loads YAML successfully
- Validation catches invalid configs
- Default config loads without errors

**Files**:
- `internal/config/config.go`
- `internal/config/loader.go`
- `internal/config/validator.go`
- `config/default.yaml`

---

## Phase 2: Core Monitoring

### Task 2.1: Define Monitor Interface
**Priority**: P0 (Blocker for platform implementations)
**Estimated Time**: 30 minutes

**Subtasks**:
- [ ] Create `internal/monitor/monitor.go`
- [ ] Define `WindowInfo` struct (PID, Title, Process, Handle)
- [ ] Define `FocusEvent` struct (Previous, Current, Time)
- [ ] Define `Monitor` interface with methods:
  - `Start(ctx context.Context) error`
  - `Stop() error`
  - `Events() <-chan FocusEvent`
  - `GetActiveWindow() (*WindowInfo, error)`
- [ ] Add documentation comments for godoc

**Acceptance Criteria**:
- Interface compiles without errors
- Clear godoc comments on all public types
- Interface designed for cross-platform use

**Files**:
- `internal/monitor/monitor.go`

---

### Task 2.2: Windows Monitor Implementation
**Priority**: P1
**Estimated Time**: 2-3 hours

**Subtasks**:
- [ ] Create `internal/monitor/monitor_windows.go` with build tag
- [ ] Add imports for `golang.org/x/sys/windows` and `syscall`
- [ ] Define `WindowsMonitor` struct
- [ ] Implement `GetForegroundWindow` syscall
- [ ] Implement `GetWindowTextW` syscall for title
- [ ] Implement `GetWindowThreadProcessId` for PID
- [ ] Implement `GetActiveWindow()` method
- [ ] Implement `Start()` with polling loop (respects poll_interval config)
- [ ] Implement `Stop()` to gracefully shutdown
- [ ] Implement focus change detection (compare previous vs current)
- [ ] Emit FocusEvents to channel
- [ ] Handle errors gracefully

**Acceptance Criteria**:
- Compiles only on Windows (build tag)
- Detects active window correctly
- Emits FocusEvent when window changes
- Poll interval configurable
- No memory leaks in polling loop

**Files**:
- `internal/monitor/monitor_windows.go`

---

### Task 2.3: macOS Monitor Implementation
**Priority**: P1
**Estimated Time**: 3-4 hours

**Subtasks**:
- [ ] Create `internal/monitor/monitor_darwin.go` with build tag
- [ ] Add CGO imports for `Cocoa` and `ApplicationServices`
- [ ] Define `DarwinMonitor` struct
- [ ] Implement CGO wrapper for `CGWindowListCopyWindowInfo`
- [ ] Implement CGO wrapper for `NSWorkspace.frontmostApplication`
- [ ] Implement `GetActiveWindow()` method
- [ ] Implement `Start()` with polling loop
- [ ] Implement `Stop()` to gracefully shutdown
- [ ] Implement focus change detection
- [ ] Emit FocusEvents to channel
- [ ] Request Accessibility permissions if needed
- [ ] Handle errors gracefully

**Technical Notes**:
- CGO required: `#cgo CFLAGS: -x objective-c`
- Frameworks: `-framework Cocoa -framework ApplicationServices`
- May need to prompt user to enable Accessibility access

**Acceptance Criteria**:
- Compiles only on macOS (build tag)
- Detects active window correctly on both Intel and ARM
- Emits FocusEvent when window changes
- Handles Accessibility permission denial gracefully

**Files**:
- `internal/monitor/monitor_darwin.go`

---

### Task 2.4: Focus Change Detector with Grace Period
**Priority**: P2
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Create `internal/monitor/detector.go`
- [ ] Define `Detector` struct with grace period timer
- [ ] Implement `ProcessEvent(event FocusEvent)` method
- [ ] Add grace period logic:
  - Start timer on focus change to blocked content
  - Cancel timer if user switches away before grace period expires
  - Emit "confirmed distraction" event after grace period
- [ ] Add debouncing to prevent rapid toggling
- [ ] Write unit tests for grace period logic

**Acceptance Criteria**:
- Grace period works as expected (500ms default)
- Rapid window switching doesn't trigger false positives
- Events emitted only after grace period

**Files**:
- `internal/monitor/detector.go`

---

## Phase 3: Blocklist Logic

### Task 3.1: Blocklist Matcher
**Priority**: P1
**Estimated Time**: 2 hours

**Subtasks**:
- [ ] Create `internal/blocker/blocker.go`
- [ ] Define `Blocker` interface
- [ ] Implement `BlocklistMatcher` struct
- [ ] Implement `IsBlocked(window *WindowInfo) bool` method
- [ ] Add process name matching (case-insensitive, handles .exe suffix)
- [ ] Add website matching via window title substring
- [ ] Add regex pattern matching
- [ ] Implement exception checking (bypasses blocklist)
- [ ] Handle errors in regex compilation
- [ ] Write comprehensive unit tests

**Test Cases**:
- Process "Discord.exe" matches "Discord" in blocklist
- Window title "Reddit - Google Chrome" matches "reddit.com" in websites
- Pattern `.*Netflix.*` matches "Watching Netflix - Google Chrome"
- Exception "YouTube Music" bypasses "youtube.com" block

**Acceptance Criteria**:
- All matching types work correctly
- Case-insensitive for processes
- Exceptions properly bypass rules
- Unit tests cover edge cases

**Files**:
- `internal/blocker/blocker.go`
- `internal/blocker/blocker_test.go`

---

### Task 3.2: Focus State Manager
**Priority**: P2
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Create `internal/blocker/state.go`
- [ ] Define `FocusState` type (Working, Distracted)
- [ ] Define `StateManager` struct
- [ ] Implement state tracking with timestamp
- [ ] Implement `OnFocusChange(event FocusEvent)` method
- [ ] Emit state change events (Working → Distracted, Distracted → Working)
- [ ] Track last safe window for reference
- [ ] Write unit tests

**Acceptance Criteria**:
- State transitions correctly based on blocklist
- Events emitted only on actual state changes
- Last safe window tracked

**Files**:
- `internal/blocker/state.go`

---

## Phase 4: Statistics

### Task 4.1: Stats Tracker
**Priority**: P2
**Estimated Time**: 1.5 hours

**Subtasks**:
- [ ] Create `internal/stats/stats.go`
- [ ] Define `Stats` struct (FocusedTime, DistractedTime, TriggerCount)
- [ ] Define `Tracker` struct with mutex for thread safety
- [ ] Implement `Start()` to begin tracking
- [ ] Implement `Stop()` to pause tracking
- [ ] Implement time tracking goroutine (updates every second)
- [ ] Implement `IncrementTrigger()` method
- [ ] Implement `GetStats()` method (returns snapshot)
- [ ] Add `CalculateFocusRate()` helper
- [ ] Write unit tests

**Thread Safety**:
- Use `sync.Mutex` to protect stats
- All public methods must lock

**Acceptance Criteria**:
- Time tracking accurate to the second
- Thread-safe for concurrent access
- FocusRate calculation correct

**Files**:
- `internal/stats/stats.go`
- `internal/stats/stats_test.go`

---

## Phase 5: Frontend (Wails + Svelte)

### Task 5.1: Frontend Project Structure
**Priority**: P1
**Estimated Time**: 30 minutes

**Subtasks**:
- [ ] Install Tailwind CSS in frontend: `npm install -D tailwindcss postcss autoprefixer`
- [ ] Initialize Tailwind: `npx tailwindcss init -p`
- [ ] Configure `tailwind.config.js` to scan `.svelte` files
- [ ] Add Tailwind directives to `app.css`
- [ ] Create component directory structure:
  ```
  frontend/src/lib/
  ├── LofiPlayer.svelte
  ├── FBIVideo.svelte
  └── StatsDisplay.svelte
  ```
- [ ] Create stores directory: `frontend/src/stores/app.ts`

**Acceptance Criteria**:
- Tailwind classes work in Svelte components
- Project structure clean and organized

---

### Task 5.2: Main App Component
**Priority**: P1
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Update `frontend/src/App.svelte`
- [ ] Create layout with two panels:
  - Main player area (lofi or FBI)
  - Stats sidebar
- [ ] Add status indicator (working/distracted)
- [ ] Implement conditional rendering (show lofi OR FBI)
- [ ] Add basic styling with Tailwind
- [ ] Make responsive

**Acceptance Criteria**:
- Layout looks clean and modern
- Components render conditionally
- Responsive to window size

**Files**:
- `frontend/src/App.svelte`

---

### Task 5.3: Lofi Player Component
**Priority**: P1
**Estimated Time**: 2 hours

**Subtasks**:
- [ ] Create `frontend/src/lib/LofiPlayer.svelte`
- [ ] Embed YouTube iframe player
- [ ] Load YouTube IFrame API from CDN
- [ ] Initialize player with playlist URL from config
- [ ] Implement volume control (slider)
- [ ] Add play/pause button (optional)
- [ ] Handle player state events (playing, paused, ended)
- [ ] Add error handling for failed embeds
- [ ] Style with Tailwind

**YouTube IFrame API**:
```javascript
const player = new YT.Player('player', {
  videoId: 'jfKfPfyJRdk',
  playerVars: { autoplay: 1, controls: 0 },
  events: {
    onReady: onPlayerReady,
    onStateChange: onPlayerStateChange
  }
});
```

**Acceptance Criteria**:
- YouTube video loads and autoplays
- Volume control works
- No error 153 (Referer issue)

**Files**:
- `frontend/src/lib/LofiPlayer.svelte`

---

### Task 5.4: FBI Meme Component
**Priority**: P1
**Estimated Time**: 1.5 hours

**Subtasks**:
- [ ] Create `frontend/src/lib/FBIVideo.svelte`
- [ ] Use HTML5 `<video>` element
- [ ] Load bundled FBI video via Wails asset path
- [ ] Implement autoplay and loop
- [ ] Add dramatic entrance animation (fade-in, scale-up)
- [ ] Set volume to 100% (or configurable)
- [ ] Handle video load errors
- [ ] Style with Tailwind

**Animation**:
```css
@keyframes fbi-enter {
  from { opacity: 0; transform: scale(0.8); }
  to { opacity: 1; transform: scale(1); }
}
```

**Acceptance Criteria**:
- FBI video plays instantly when triggered
- Loops continuously
- Dramatic entrance animation
- High volume for comedic effect

**Files**:
- `frontend/src/lib/FBIVideo.svelte`

---

### Task 5.5: Stats Display Component
**Priority**: P2
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Create `frontend/src/lib/StatsDisplay.svelte`
- [ ] Display focused time (formatted as HH:MM:SS)
- [ ] Display distracted time
- [ ] Display FBI trigger count
- [ ] Calculate and display focus rate percentage
- [ ] Add fun icons/emojis
- [ ] Update stats in real-time (poll from backend)
- [ ] Style with Tailwind (card layout)

**Display Format**:
```
📊 Session Stats
✅ Focused: 1h 23m 45s
🚨 Distracted: 5m 12s
🎯 FBI Triggers: 7x
💪 Focus Rate: 94.1%
```

**Acceptance Criteria**:
- Stats display correctly formatted
- Real-time updates (every second)
- Focus rate calculation correct

**Files**:
- `frontend/src/lib/StatsDisplay.svelte`

---

## Phase 6: Backend Integration

### Task 6.1: Create Application Orchestrator
**Priority**: P0 (Blocker)
**Estimated Time**: 2 hours

**Subtasks**:
- [ ] Create `internal/app/app.go`
- [ ] Define `App` struct with fields:
  - config *config.Config
  - monitor monitor.Monitor
  - blocker blocker.Blocker
  - stats *stats.Tracker
  - playerState (lofi/fbi)
- [ ] Implement `New(configPath string) (*App, error)` constructor
- [ ] Implement `Start(ctx context.Context) error` method
- [ ] Wire up event flow:
  - Monitor events → Blocker → State change → Update player
- [ ] Implement graceful shutdown in `Stop()` method
- [ ] Add logging with zerolog

**Event Flow**:
```
Monitor (FocusEvent) → Blocker (IsBlocked?) → StateManager (Working/Distracted)
                                                    ↓
                                            Update PlayerState
                                                    ↓
                                            Notify Frontend
```

**Acceptance Criteria**:
- All components initialized correctly
- Event flow works end-to-end
- Graceful shutdown on SIGINT

**Files**:
- `internal/app/app.go`

---

### Task 6.2: Wails Bindings
**Priority**: P1
**Estimated Time**: 2 hours

**Subtasks**:
- [ ] Update `cmd/vigilant/main.go`
- [ ] Initialize Wails app with App struct
- [ ] Bind methods to frontend:
  - `GetConfig() *config.Config`
  - `GetStats() *stats.Stats`
  - `GetPlayerState() string` (lofi/fbi)
  - `SetVolume(level float64) error`
- [ ] Implement runtime event emitting (Go → JS):
  - `player:state-change` (lofi/fbi)
  - `stats:update` (new stats)
- [ ] Set up Wails app options (window size, title, etc.)

**Wails Runtime Events**:
```go
runtime.EventsEmit(ctx, "player:state-change", "fbi")
runtime.EventsEmit(ctx, "stats:update", app.stats.GetStats())
```

**Acceptance Criteria**:
- Frontend can call Go methods
- Go can emit events to frontend
- Bidirectional communication works

**Files**:
- `cmd/vigilant/main.go`

---

### Task 6.3: Player Controller
**Priority**: P1
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Create `internal/player/player.go`
- [ ] Define `PlayerState` enum (Lofi, FBI)
- [ ] Define `Controller` struct
- [ ] Implement `SwitchToLofi()` method
- [ ] Implement `SwitchToFBI()` method
- [ ] Emit events to frontend on state change
- [ ] Track current state

**Acceptance Criteria**:
- State changes reflected in frontend
- Events emitted correctly

**Files**:
- `internal/player/player.go`

---

### Task 6.4: Application Lifecycle
**Priority**: P2
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Implement window positioning logic for second monitor
- [ ] Add always-on-top option (from config)
- [ ] Implement graceful shutdown:
  - Stop monitor
  - Stop stats tracker
  - Close event channels
- [ ] Add logging for lifecycle events

**Acceptance Criteria**:
- App opens on second monitor if available
- Always-on-top works if configured
- Clean shutdown on exit

**Files**:
- `cmd/vigilant/main.go`
- `internal/app/app.go`

---

## Phase 7: Asset Bundling

### Task 7.1: Embed FBI Video
**Priority**: P1
**Estimated Time**: 30 minutes

**Subtasks**:
- [ ] Download/create FBI meme video (< 5MB recommended)
- [ ] Add video to `assets/fbi-meme.mp4`
- [ ] Configure Wails to embed assets in `wails.json`:
  ```json
  {
    "assetdir": "./assets",
    "embedAssets": true
  }
  ```
- [ ] Update frontend to load video from embedded path
- [ ] Test video playback in dev and production builds

**Acceptance Criteria**:
- FBI video plays in embedded player
- Video bundled in final executable
- No external dependencies

**Files**:
- `assets/fbi-meme.mp4`
- `wails.json`

---

### Task 7.2: App Icons
**Priority**: P3 (Nice to have)
**Estimated Time**: 30 minutes

**Subtasks**:
- [ ] Create or source app icon (1024x1024 PNG)
- [ ] Convert to Windows .ico format (use icon converter)
- [ ] Convert to macOS .icns format (use icon converter)
- [ ] Add icons to `build/windows/icon.ico` and `build/darwin/icon.icns`
- [ ] Configure Wails to use icons in `wails.json`:
  ```json
  {
    "icon": {
      "windows": "build/windows/icon.ico",
      "darwin": "build/darwin/icon.icns"
    }
  }
  ```

**Icon Ideas**:
- Eye symbol (vigilance)
- FBI logo parody
- Lofi aesthetic

**Acceptance Criteria**:
- Icons display in app window and taskbar
- Works on both platforms

**Files**:
- `build/windows/icon.ico`
- `build/darwin/icon.icns`

---

## Phase 8: Build & Distribution

### Task 8.1: Build Scripts
**Priority**: P1
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Test `make build-windows` on Windows or with cross-compilation
- [ ] Test `make build-darwin` on macOS
- [ ] Test `make build-darwin-universal` for Intel + ARM
- [ ] Verify binary sizes (should be 10-15MB)
- [ ] Add version info to builds
- [ ] Create build output structure:
  ```
  build/bin/
  ├── vigilant-windows-amd64.exe
  ├── vigilant-darwin-amd64
  └── vigilant-darwin-arm64
  ```

**Acceptance Criteria**:
- Builds work for all platforms
- Binaries are portable (no dependencies)
- Binary size reasonable

---

### Task 8.2: Testing on Platforms
**Priority**: P0
**Estimated Time**: 2-3 hours

**Subtasks**:
- [ ] Test on Windows 10/11:
  - Focus monitoring works
  - Blocklist correctly identifies apps
  - FBI meme triggers
  - Lofi player works
- [ ] Test on macOS Intel:
  - Same test cases
  - Accessibility permissions prompt
- [ ] Test on macOS Apple Silicon (if available):
  - Universal binary works
- [ ] Test edge cases:
  - No second monitor
  - Invalid config file
  - Blocked app not installed

**Acceptance Criteria**:
- All features work on all platforms
- No crashes or critical bugs
- Graceful error handling

---

## Phase 9: Documentation

### Task 9.1: Code Documentation
**Priority**: P2
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Add package-level godoc comments to all packages
- [ ] Document all public interfaces and methods
- [ ] Add inline comments for complex logic (e.g., CGO, syscalls)
- [ ] Run `go doc` to verify documentation

**Acceptance Criteria**:
- All public APIs documented
- godoc generates readable documentation
- Code easy to follow for demonstration

---

### Task 9.2: User Documentation (README.md)
**Priority**: P1
**Estimated Time**: 1 hour

**Subtasks**:
- [ ] Create comprehensive `README.md` with:
  - Project description
  - Features list
  - Screenshots/GIFs (optional)
  - Installation instructions (download and run)
  - Configuration guide (editing config.yaml)
  - Usage instructions
  - Building from source
  - FAQ (e.g., Accessibility permissions on macOS)
  - License
- [ ] Add badges (build status, version, etc.)

**Acceptance Criteria**:
- README clear and comprehensive
- New users can get started quickly

**Files**:
- `README.md`

---

### Task 9.3: AI Assistant Documentation (CLAUDE.md)
**Priority**: P3
**Estimated Time**: 30 minutes

**Subtasks**:
- [ ] Create `CLAUDE.md` for Claude Code guidance:
  - Project overview
  - Technology stack
  - Development commands (make dev, make build)
  - Architecture overview
  - Key directories and files
  - Testing approach
  - Contribution guidelines

**Acceptance Criteria**:
- CLAUDE.md helpful for AI assistants
- Covers key development workflows

**Files**:
- `CLAUDE.md`

---

## Summary

### Total Estimated Time
- Phase 1: 2 hours
- Phase 2: 7 hours
- Phase 3: 3 hours
- Phase 4: 1.5 hours
- Phase 5: 6 hours
- Phase 6: 6 hours
- Phase 7: 1 hour
- Phase 8: 4 hours
- Phase 9: 2.5 hours

**Total: ~33 hours (~4-5 days of focused work)**

### Critical Path
1. Project Setup (1.1, 1.3) → 2 hours
2. Monitor Interface + Windows Impl (2.1, 2.2) → 3 hours
3. Blocker Logic (3.1) → 2 hours
4. Frontend Components (5.1-5.4) → 5 hours
5. Backend Integration (6.1, 6.2) → 4 hours
6. Testing (8.2) → 3 hours

**Critical Path: ~19 hours**

### Priority Levels
- **P0**: Must complete for MVP (blocking other tasks)
- **P1**: Core features required for MVP
- **P2**: Important features, can be deferred slightly
- **P3**: Nice to have, not critical for MVP

---

**Document Version**: 1.0
**Last Updated**: 2025-12-10
**Status**: Ready for Implementation
