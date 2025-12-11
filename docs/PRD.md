# Vigilant - Product Requirements Document

## Executive Summary

**Vigilant** is a cross-platform (Windows/macOS) desktop application that helps users stay focused by monitoring their active window and triggering a humorous FBI meme video whenever they try to access blocked applications or websites. Built for content creators and productivity enthusiasts who want a fun, meme-worthy approach to focus management.

---

## Product Overview

### Vision
Transform productivity monitoring from boring to entertaining by using humor and social pressure to discourage distractions.

### Core Value Proposition
- **Fun & Meme-Worthy**: FBI meme makes distraction embarrassing and funny (perfect for TikTok/Instagram content)
- **Non-Invasive**: Passive alerts let users stay in control
- **Productive Ambience**: Lofi music creates a pleasant work environment
- **Social Flex**: Statistics to showcase productivity on social media

---

## Target Audience

### Primary Users
1. **Content Creators** (YouTube, TikTok, Instagram)
   - Want to create productivity content
   - Need visually interesting tools for videos
   - Appreciate meme culture

2. **Developers & Tech Professionals**
   - Easily distracted by Discord, Reddit, Twitter
   - Want lightweight, non-intrusive focus tools
   - Appreciate clean code and open-source tools

3. **Students & Remote Workers**
   - Struggle with self-discipline
   - Want gamified productivity
   - Enjoy sharing stats with friends

---

## Key Features

### F1: Focus Monitoring
**Description**: Real-time monitoring of active window/process focus.

**Requirements**:
- Monitor active window on Windows (Win32 API: GetForegroundWindow, GetWindowText)
- Monitor active window on macOS (CGWindowListCopyWindowInfo, NSWorkspace)
- Detect browser tab titles via window title reading
- Poll interval: 100ms (configurable)
- Grace period: 500ms before triggering FBI meme
- Cross-platform unified interface

**Acceptance Criteria**:
- Detects window switches within 100ms
- Works with Chrome, Edge, Brave (Chromium-based browsers)
- Minimal CPU usage (<2% when idle)

---

### F2: Configurable Blocklist
**Description**: User-defined list of processes and websites to block.

**Requirements**:
- Block specific process names (e.g., "Discord.exe", "Steam")
- Block websites via window title matching (e.g., "reddit.com" in "Reddit - Google Chrome")
- Support regex patterns for advanced matching (e.g., `.*Netflix.*`)
- YAML configuration file for easy editing
- Exceptions list to allow specific windows matching blocklist
- Case-insensitive matching for processes
- Hot-reload configuration without restart (future enhancement)

**Configuration Format**:
```yaml
blocklist:
  processes:
    - "Discord"
    - "Steam"
    - "Slack"
  websites:
    - "youtube.com"
    - "twitter.com"
    - "reddit.com"
  patterns:
    - ".*Netflix.*"
    - ".*Prime Video.*"

exceptions:
  - "YouTube Music"
  - "Work - Reddit"
```

**Acceptance Criteria**:
- Matches process names correctly on both platforms
- Matches website URLs in browser window titles
- Regex patterns work as expected
- Exceptions bypass blocklist rules

---

### F3: Lofi Music Player
**Description**: Embedded YouTube player for lo-fi music during work sessions.

**Requirements**:
- Embedded YouTube iframe player in Wails WebView
- Default playlist: Lofi Girl stream (https://www.youtube.com/watch?v=jfKfPfyJRdk)
- Volume control (0-100%)
- Play/pause controls
- Continuous playback during normal work
- Configurable playlist URL via config file

**Technical Notes**:
- Use YouTube IFrame Player API
- Set proper HTTP Referer via Wails baseUrl config to avoid error 153
- Handle playlist/single video URLs

**Acceptance Criteria**:
- YouTube video loads and plays in embedded player
- Volume control works smoothly
- Player continues during window focus changes (non-blocked content)

---

### F4: FBI Meme Trigger
**Description**: Comedic deterrent that plays when user focuses on blocked content.

**Requirements**:
- Detect when blocked content is focused
- Switch from lofi music to FBI meme video
- Play FBI "open up" meme at higher volume (or full volume)
- Loop FBI video until user refocuses on safe content
- Immediate switch back to lofi when user returns to work
- FBI video bundled in executable (no download required)

**User Flow**:
1. User focuses on Discord/blocked website
2. Grace period (500ms) elapses
3. FBI meme video replaces lofi player with dramatic entrance
4. FBI video loops continuously
5. User manually switches to different window (e.g., IDE)
6. FBI meme stops, lofi player resumes

**Acceptance Criteria**:
- FBI meme triggers within 1 second of focusing blocked content
- Video plays smoothly with good quality
- Instant switch back to lofi when refocusing on work
- No lag or audio glitches during transitions

---

### F5: Statistics Tracking
**Description**: Track and display productivity metrics.

**Requirements**:
- Track total focused time (working on non-blocked content)
- Track total distracted time (FBI meme playing)
- Count number of FBI triggers (distraction attempts)
- Display stats in UI with fun formatting
- Session-based tracking (resets on app restart)
- Real-time updates

**Display Format**:
```
📊 Session Stats
✅ Focused: 1h 23m 45s
🚨 Distracted: 5m 12s
🎯 FBI Triggers: 7 times
💪 Focus Rate: 94.1%
```

**Future Enhancements** (not MVP):
- Daily/weekly statistics
- Export stats to JSON
- Historical tracking with SQLite

**Acceptance Criteria**:
- Time tracking accurate to the second
- Stats update in real-time in UI
- Focus rate calculated correctly

---

### F6: Second Monitor Support
**Description**: Display app window on secondary monitor for always-visible monitoring.

**Requirements**:
- Detect available monitors
- Prefer launching on second monitor if available
- Fall back to primary monitor if only one display
- Optional always-on-top mode
- Remember window position (future enhancement)

**Acceptance Criteria**:
- App opens on second monitor when available
- Window can be moved between monitors
- Always-on-top option works correctly

---

## Technical Requirements

### Platform Support
- **Windows**: 10/11 (amd64)
- **macOS**: 12+ Monterey and later (Intel and Apple Silicon/ARM64)

### Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Language** | Go 1.21+ | Cross-platform, fast, single binary |
| **GUI Framework** | Wails v2 | Small binaries (~4MB base), native WebView |
| **Frontend** | Svelte 4+ | Lightweight, reactive, fast build |
| **Styling** | Tailwind CSS | Rapid UI development, small bundle |
| **Configuration** | Viper | YAML/JSON support, hot-reload capable |
| **Logging** | zerolog | Fast structured logging |
| **Windows API** | golang.org/x/sys/windows | Standard library extension |
| **macOS API** | CGO (Cocoa/ApplicationServices) | Native window detection |

### Architecture Principles

**SOLID**:
- **Single Responsibility**: Monitor, Blocker, Player each handle one concern
- **Open/Closed**: Platform implementations extend interfaces without modifying core
- **Liskov Substitution**: Windows/Darwin monitors interchangeable via Monitor interface
- **Interface Segregation**: Small, focused interfaces (Monitor, Blocker, Player)
- **Dependency Inversion**: Core app depends on interfaces, not concrete implementations

**DRY**:
- Platform-specific code isolated in `_windows.go` and `_darwin.go` files
- Shared logic in common interface files
- Configuration types defined once

**YAGNI**:
- No premature optimization
- No database for MVP (in-memory stats)
- No plugins or extensibility until needed

---

## Non-Functional Requirements

### Performance
- **Startup Time**: < 2 seconds from launch to UI displayed
- **Memory Usage**: < 100MB RAM during normal operation
- **CPU Usage**: < 2% when idle, < 5% during monitoring
- **Binary Size**: 10-15MB (with bundled FBI video)

### Reliability
- **Crash Recovery**: Graceful handling of errors, no data corruption
- **Platform Stability**: No kernel panics or system instability
- **Error Handling**: All errors logged, user-friendly error messages

### Usability
- **Zero Installation**: Download and run, no setup required
- **Intuitive UI**: No learning curve, obvious controls
- **Configuration**: Easy-to-edit YAML file with examples
- **Documentation**: Clear README with setup and usage

### Security
- **Privacy**: No telemetry, no data collection, no internet required (except YouTube)
- **Permissions**: Minimal OS permissions (window title reading only)
- **Sandboxing**: No elevated privileges required

---

## Configuration Schema

### Default Configuration File

Location: `config/default.yaml` (bundled) or `~/.vigilant/config.yaml` (user override)

```yaml
# Vigilant Configuration File

blocklist:
  # Process names to block (case-insensitive)
  processes:
    - "Discord"
    - "Discord.exe"
    - "Slack"
    - "Slack.exe"
    - "Steam"
    - "Steam.exe"
    - "Battle.net"

  # Website domains to block (matches window titles)
  websites:
    - "youtube.com"
    - "twitter.com"
    - "x.com"
    - "reddit.com"
    - "instagram.com"
    - "tiktok.com"
    - "facebook.com"
    - "twitch.tv"

  # Regex patterns for advanced matching
  patterns:
    - ".*Netflix.*"
    - ".*Prime Video.*"
    - ".*Disney\\+.*"
    - ".*Hulu.*"

# Exceptions - allow these even if they match blocklist
exceptions:
  - "YouTube Music"        # Allow music streaming
  - "Work - YouTube"       # Allow work-related YouTube
  - "YouTube Studio"       # Allow content creation

player:
  # YouTube playlist or video URL for lofi music
  lofi_playlist: "https://www.youtube.com/watch?v=jfKfPfyJRdk"

  # Path to FBI meme video (embedded:// for bundled assets)
  fbi_meme_source: "embedded://fbi-meme.mp4"

  # Default volume (0.0 - 1.0)
  default_volume: 0.5

monitor:
  # How often to check active window (duration)
  poll_interval: 100ms

  # Grace period before triggering FBI meme (duration)
  grace_period: 500ms

ui:
  # UI mode: "gui" or "tui" (future)
  mode: "gui"

  # Prefer second monitor for app window
  second_monitor: true

  # Keep window always on top
  always_on_top: false
```

---

## Success Criteria

### MVP Success Metrics
1. ✅ App compiles and runs on Windows and macOS without errors
2. ✅ Focus monitoring detects window changes within 100ms
3. ✅ Blocklist correctly identifies blocked processes/websites
4. ✅ FBI meme triggers within 1 second of focusing blocked content
5. ✅ Lofi player works with embedded YouTube
6. ✅ Statistics track time accurately
7. ✅ Single portable executable < 20MB
8. ✅ Clean, well-documented code suitable for demonstration

### User Validation
- 5+ beta testers successfully use app for 1 week
- Positive feedback on UX and fun factor
- No critical bugs or crashes reported

### Content Creation
- Successfully record TikTok/Instagram Reel demonstrating app
- Code walkthrough suitable for technical audience
- Clean architecture visible in code demonstration

---

## Out of Scope (Future Enhancements)

### Phase 2
- Hot-reload configuration without restart
- System tray integration with quick controls
- Multiple work profiles (e.g., "Deep Work", "Light Work", "Meeting Mode")
- Custom FBI meme video upload

### Phase 3
- Pomodoro timer integration
- Terminal TUI mode (Bubble Tea)
- Historical statistics with database
- Export stats to JSON/CSV

### Phase 4
- Mobile companion app (notification sync)
- Team/enterprise features (shared blocklists)
- AI-powered distraction scoring
- Integration with time tracking apps (Toggl, RescueTime)

---

## Dependencies & Risks

### Technical Dependencies
- **Wails v2**: Stable, well-maintained
- **YouTube IFrame API**: Public API, should remain stable
- **Windows/macOS APIs**: Standard OS APIs, low risk

### Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| YouTube changes embed policy | High | Use backup lofi audio files, allow local video URLs |
| Browser title format changes | Medium | Make website matching flexible with regex |
| macOS Accessibility permissions denied | Medium | Clear prompts, documentation on enabling access |
| FBI video copyright issues | Low | Use royalty-free alternative or user-provided video |
| Wails breaking changes | Low | Pin specific Wails version, test before updating |

---

## Timeline & Milestones

### Phase 1: Foundation (Week 1)
- Project setup, go.mod, Wails init
- Configuration system with Viper
- Basic project structure

### Phase 2: Core Monitoring (Week 1-2)
- Monitor interface and Windows implementation
- macOS implementation with CGO
- Focus detection and grace period

### Phase 3: Logic & Stats (Week 2)
- Blocklist matching (processes, websites, patterns)
- Statistics tracking
- State management

### Phase 4: Frontend (Week 2-3)
- Svelte app structure
- Lofi player component (YouTube embed)
- FBI video component
- Stats display

### Phase 5: Integration & Testing (Week 3)
- Backend-frontend bindings
- End-to-end testing on Windows
- End-to-end testing on macOS

### Phase 6: Polish & Release (Week 4)
- Asset bundling (FBI video, icons)
- Build scripts and distribution
- Documentation (README, code comments)
- Beta testing and bug fixes

**Total Timeline**: 4 weeks for MVP

---

## Appendix

### Key Files Structure
```
vigilant/
├── cmd/vigilant/main.go              # Entry point
├── internal/
│   ├── app/app.go                    # Orchestrator
│   ├── config/config.go              # Configuration
│   ├── monitor/
│   │   ├── monitor.go                # Interface
│   │   ├── monitor_windows.go        # Windows impl
│   │   └── monitor_darwin.go         # macOS impl
│   ├── blocker/blocker.go            # Blocklist logic
│   ├── player/player.go              # Player control
│   └── stats/stats.go                # Statistics
├── frontend/                         # Svelte app
├── assets/fbi-meme.mp4              # Bundled video
├── config/default.yaml               # Default config
├── PRD.md                            # This document
├── TASKS.md                          # Implementation tasks
└── README.md                         # User documentation
```

### References
- [Wails Documentation](https://wails.io/docs/introduction/)
- [YouTube IFrame Player API](https://developers.google.com/youtube/iframe_api_reference)
- [Windows API GetForegroundWindow](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getforegroundwindow)
- [macOS CGWindowListCopyWindowInfo](https://developer.apple.com/documentation/coregraphics/1455137-cgwindowlistcopywindowinfo)
- [Viper Configuration](https://github.com/spf13/viper)

---

**Document Version**: 1.0
**Last Updated**: 2025-12-10
**Status**: Approved for Implementation
