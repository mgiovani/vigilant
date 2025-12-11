# Vigilant - Stay Focused

A cross-platform desktop application that helps developers, students, and content creators stay focused by monitoring their active window and triggering a humorous FBI meme whenever they try to access blocked applications or websites.

## Features

- **Real-Time Window Monitoring**: Detects when you switch to blocked apps/websites
- **FBI Meme Deterrent**: Comedic FBI "Open Up" video plays at full volume (perfect for TikTok!)
- **Lofi Music Player**: Embedded YouTube Lofi Girl stream for productive ambience
- **Live Statistics**: Track your focus rate, distraction time, and FBI triggers in real-time
- **Customizable Blocklist**: Easily configure which processes and websites to block
- **Cross-Platform**: Works on Windows 10/11 and macOS (Intel and Apple Silicon)
- **Lightweight**: Single portable executable, no installation required

## System Requirements

### macOS
- **Version**: macOS 12 Monterey or later
- **Processor**: Intel (x86_64) or Apple Silicon (ARM64)
- **Memory**: 100 MB RAM
- **Storage**: 20 MB disk space

### Windows
- **Version**: Windows 10/11
- **Processor**: x86-64
- **Memory**: 100 MB RAM
- **Storage**: 20 MB disk space

## Installation

### Quick Start

1. **Download** the latest release from [Releases Page](https://github.com/mgiovani/vigilant/releases)
2. **macOS**:
   - Download `vigilant-darwin-universal` (works on Intel and Apple Silicon)
   - Make executable: `chmod +x vigilant-darwin-universal`
   - Run: `./vigilant-darwin-universal`
3. **Windows**:
   - Download `vigilant-windows-amd64.exe`
   - Run the executable
4. **Grant Permissions** (macOS only):
   - On first run, grant Accessibility permissions when prompted
   - System Preferences → Security & Privacy → Accessibility → Add Vigilant

### Building from Source

**Requirements**:
- Go 1.21+
- Node.js 18+ (for frontend)
- Make

**Steps**:
```bash
# Clone repository
git clone https://github.com/mgiovani/vigilant.git
cd vigilant

# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Development mode (with hot reload)
make dev

# Build production binary
make build-darwin          # macOS Intel
make build-darwin-universal # macOS Universal
make build-windows         # Windows (requires Windows environment)
```

## Configuration

### Default Blocklist

By default, Vigilant blocks:
- **Applications**: Discord, Slack, Steam, Battle.net
- **Websites**: YouTube, Twitter/X, Reddit, Instagram, TikTok, Facebook, Twitch
- **Patterns**: Netflix, Prime Video, Disney+, Hulu

### Customizing Your Blocklist

1. **Locate config file**:
   - First launch creates: `~/.vigilant/config.yaml`
   - Or use bundled default: `config/default.yaml`

2. **Edit config.yaml**:
   ```yaml
   blocklist:
     processes:
       - "Discord"
       - "Slack"
     websites:
       - "reddit.com"
       - "twitter.com"
     patterns:
       - ".*Netflix.*"

   exceptions:
     - "YouTube Music"      # YouTube Music won't trigger FBI
     - "Work - YouTube"     # Custom exceptions

   player:
     lofi_playlist: "https://www.youtube.com/watch?v=jfKfPfyJRdk"
     default_volume: 0.5    # 0.0-1.0

   monitor:
     poll_interval: 100ms   # Check window every 100ms
     grace_period: 500ms    # 500ms delay before FBI meme
   ```

3. **Restart** Vigilant to apply changes

### Configuration Format

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `blocklist.processes` | List | See above | Application names to block |
| `blocklist.websites` | List | See above | Domain names to block (matched in window titles) |
| `blocklist.patterns` | List | See above | Regex patterns for advanced matching |
| `exceptions` | List | [] | Exceptions that bypass blocklist |
| `player.lofi_playlist` | URL | Lofi Girl stream | YouTube video/playlist URL |
| `monitor.poll_interval` | Duration | 100ms | How often to check active window |
| `monitor.grace_period` | Duration | 500ms | Delay before triggering FBI meme |

## Usage

### Basic Workflow

1. **Start Vigilant**: Run the application
2. **Work normally**: Your focus time accumulates
3. **Try to distract**: Switch to Discord, Reddit, etc.
4. **Grace period**: 500ms timer starts (you can switch back before it expires)
5. **FBI trigger**: If timer expires, FBI meme plays at full volume
6. **Return to work**: Switch back to IDE/document, lofi resumes

### Reading the UI

```
┌─────────────────────────────────────────┬──────────────────┐
│                                         │  📊 Session Stats│
│          Lofi Music Player              │  ✅ Focused: 2h  │
│        (YouTube Embedded)               │  🚨 Distracted:5m│
│                                         │  🎯 Triggers: 12 │
│          Volume: ●─────→                │  💪 Focus: 95.7% │
│                                         │                  │
│       Status: 🟢 Working                │                  │
└─────────────────────────────────────────┴──────────────────┘
```

## Troubleshooting

### FAQ

**Q: YouTube says "error 153" or video won't load**
A: This is a common Wails issue with YouTube embedding. Try:
   1. Check internet connection
   2. Restart Vigilant
   3. Update to latest version
   4. If persists, YouTube may have changed policies (check releases)

**Q: FBI meme is too loud!**
A: You can:
   - Lower default volume in config: `player.default_volume: 0.3`
   - Use mute button in UI during testing
   - Restart Vigilant after config change

**Q: App says "Accessibility Permission Denied" (macOS)**
A: Grant permission:
   1. System Preferences → Security & Privacy → Accessibility
   2. Click the lock to unlock
   3. Click "+" button and select Vigilant application
   4. Restart Vigilant

**Q: Windows Defender blocks the app**
A: This is a SmartScreen false positive for unsigned apps. You can:
   - Click "More info" → "Run anyway"
   - Sign the executable (future release)
   - Build from source (requires Go compiler)

**Q: App is using too much CPU/memory**
A: This is normal, but check:
   - `poll_interval` in config (higher = less CPU, less responsive)
   - Close other heavy applications
   - Restart Vigilant
   - Report issue on GitHub if excessive

**Q: Can I use Vigilant on Linux?**
A: Not yet. Linux support is planned for Phase 2.

### Getting Help

- **GitHub Issues**: https://github.com/mgiovani/vigilant/issues
- **Discussions**: https://github.com/mgiovani/vigilant/discussions

## Development

### Project Structure

```
vigilant/
├── cmd/vigilant/
│   └── main.go                 # Application entry point
├── internal/
│   ├── app/                    # Orchestrator and bindings
│   ├── config/                 # Configuration loading
│   ├── monitor/                # Window monitoring (Windows/macOS)
│   ├── blocker/                # Blocklist matching and state
│   ├── player/                 # Media player control
│   └── stats/                  # Statistics tracking
├── frontend/                   # Svelte + Tailwind UI
│   ├── src/
│   │   ├── App.svelte         # Main layout
│   │   ├── lib/components/    # UI components
│   │   ├── stores/            # State management
│   │   └── types/             # TypeScript types
│   └── package.json           # Frontend dependencies
├── config/                     # Configuration files
├── assets/                     # Embedded assets (FBI video)
├── build/                      # Build output
├── Makefile                    # Build automation
└── README.md                   # This file
```

### Making Changes

1. **Backend (Go)**:
   ```bash
   # Run development server with hot reload
   make dev

   # Run tests
   make test

   # Build for your platform
   make build-darwin
   ```

2. **Frontend (Svelte)**:
   - Changes in `frontend/src/` auto-reload when using `make dev`
   - Tailwind CSS compiles on save
   - TypeScript errors show in console

### Testing

```bash
# Run all tests
make test

# Run specific package
go test ./internal/blocker -v

# Check for race conditions
go test ./... -race

# Code coverage
go test ./... -cover
```

### Building for Distribution

```bash
# macOS (Intel)
make build-darwin

# macOS (Universal - Intel + Apple Silicon)
make build-darwin-universal

# Windows (requires Windows environment)
make build-windows

# Clean build artifacts
make clean
```

## Architecture

### Event Flow

```
Monitor (Poll every 100ms)
  ↓
  FocusEvent (Window changed?)
  ↓
  Detector (Grace period)
  ↓
  Blocker (Is it blocked?)
  ↓
  StateManager (Working or Distracted?)
  ↓
  PlayerController (Switch player)
  ↓
  StatsTracker (Accumulate time)
  ↓
  Frontend (Update UI via Wails events)
```

### Key Components

- **Monitor**: Detects active window (platform-specific: Win32/Cocoa)
- **Detector**: Applies grace period debouncing
- **Blocker**: Matches window against blocklist rules
- **StateManager**: Tracks focus state transitions
- **StatsTracker**: Accumulates productivity metrics
- **PlayerController**: Manages lofi/FBI player switching
- **App Orchestrator**: Coordinates all components

## Performance

- **Startup Time**: < 2 seconds
- **Memory (Idle)**: < 100 MB
- **CPU (Monitoring)**: < 5%
- **Binary Size**: 8-16 MB (depending on platform)

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Commit changes with clear messages
4. Push to your fork
5. Open a Pull Request

## Roadmap

### Phase 2 (Upcoming)
- Hot-reload configuration without restart
- System tray integration
- Multiple work profiles (Deep Work, Meeting Mode, etc.)
- Custom FBI meme video upload
- Linux support

### Phase 3
- Pomodoro timer
- Historical statistics with database
- Export stats to CSV/JSON
- Terminal UI mode (TUI)

### Phase 4
- Mobile companion app
- Team productivity analytics
- Cloud synchronization
- Integration with Toggl, RescueTime

## Acknowledgments

- [Wails](https://wails.io/) - Go desktop application framework
- [Svelte](https://svelte.dev/) - Reactive component framework
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS
- Lofi Girl - For the relaxing background music

## Contact

- GitHub: [@mgiovani](https://github.com/mgiovani/vigilant)

---

**Stay focused. Stay productive. Stay vigilant.**
