# Vigilant Build Guide

## Quick Start

### Build the Application

```bash
# Clean previous builds
make clean

# Build for your platform
make build-darwin-universal    # macOS (Intel & Apple Silicon)
# or
make build-darwin              # macOS Intel only
# or
make build-windows             # Windows (requires Windows)
```

### Run the Application

```bash
# Open the built app
open build/bin/vigilant.app

# Or launch directly
./build/bin/vigilant.app/Contents/MacOS/vigilant
```

## Available Make Targets

```bash
make help                      # Show all available commands
make dev                       # Run development server with hot reload
make build-darwin              # Build macOS Intel binary
make build-darwin-universal    # Build macOS universal binary (Intel + ARM64)
make build-windows             # Build Windows executable
make clean                     # Remove build artifacts
make test                      # Run all Go tests
```

## Build Specifications

### macOS Binaries

#### Universal Binary (Recommended)
- **Target**: `make build-darwin-universal`
- **Size**: 16 MB
- **Architecture**: x86_64 + arm64 (Intel & Apple Silicon)
- **Build Time**: ~33 seconds
- **Location**: `build/bin/vigilant.app`

#### Intel Binary
- **Target**: `make build-darwin`
- **Size**: 8.5 MB
- **Architecture**: x86_64 only
- **Build Time**: ~44 seconds
- **Location**: `build/bin/vigilant.app`

### Windows Binary
- **Target**: `make build-windows`
- **Size**: ~12 MB (estimated)
- **Architecture**: amd64
- **Requires**: Windows environment or cross-compilation setup
- **Location**: `build/bin/vigilant.exe`

## Development Workflow

### Development Mode (with Hot Reload)
```bash
make dev
```
This runs the Wails development server with:
- Frontend hot reload
- Debug console access
- Rapid development iteration

### Running Tests
```bash
make test
```
All tests should pass:
- vigilant/internal/app
- vigilant/internal/blocker
- vigilant/internal/config
- vigilant/internal/monitor
- vigilant/internal/player
- vigilant/internal/stats

## Build Verification

After building, verify the binary:

```bash
# Check file type
file build/bin/vigilant.app/Contents/MacOS/vigilant

# Check size
du -sh build/bin/vigilant.app

# List contents
ls -lh build/bin/vigilant.app/Contents/

# Run the app
./build/bin/vigilant.app/Contents/MacOS/vigilant
```

## Build Configuration

### Makefile
Located at: `./Makefile`
- Defines all build targets
- Manages clean builds
- Configures platform-specific options

### Wails Configuration
Located at: `./wails.json`
- Window settings (1200x800)
- Asset directory (./assets)
- Build icons

### Application Configuration
Located at: `./config/default.yaml`
- Blocklist settings
- Player configuration
- Monitor settings
- UI preferences

## System Requirements

### To Build
- Go 1.21+
- Wails CLI 2.11+
- Node.js 16+ (for frontend)
- macOS 10.13+ (for building on macOS)

### To Run
- macOS 10.13+ (Intel or Apple Silicon)
- OR Windows 10+
- 16 MB disk space
- 50-150 MB RAM

## Troubleshooting

### Build Fails
1. Update Wails: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
2. Clear node modules: `rm -rf frontend/node_modules && npm install`
3. Run tests to verify setup: `make test`

### App Won't Launch
1. Check permissions: `ls -l build/bin/vigilant.app/Contents/MacOS/vigilant`
2. Make executable: `chmod +x build/bin/vigilant.app/Contents/MacOS/vigilant`
3. Check logs: Run app from terminal to see error messages

### Tests Fail
1. Check Go version: `go version` (should be 1.21+)
2. Clear cache: `go clean -testcache`
3. Run tests with verbose: `go test -v ./...`

## Distribution

### macOS App Bundle
The built `.app` is a standard macOS application bundle:
```
vigilant.app/
├── Contents/
│   ├── MacOS/
│   │   └── vigilant (executable)
│   ├── Resources/
│   │   └── appicon.icns
│   ├── PkgInfo
│   └── Info.plist
```

### Code Signing (Optional)
To sign for distribution:
```bash
codesign -s - build/bin/vigilant.app
```

### Notarization (Optional)
For distribution on macOS 10.15+, notarize the app:
```bash
xcrun altool --notarize-app -f build/bin/vigilant.app ...
```

## Performance Notes

- **Build Time**: Intel 44s, Universal 33s
- **Binary Size**: Intel 8.5 MB, Universal 16 MB
- **Memory Usage**: ~50-100 MB idle, ~150 MB active
- **Startup Time**: <2 seconds
- **CPU Usage**: <5% idle

## Next Steps

After building:
1. Test the application thoroughly
2. Check configuration loads properly
3. Test blocklist functionality
4. Verify FBI meme video plays
5. Test stats tracking
6. For distribution, follow Phase 9 documentation

---

For detailed test results, see: `TEST_RESULTS.md`
For project status, see: `PHASE_8_COMPLETE.md`
