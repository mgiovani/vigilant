# Vigilant - Deployment Guide

## Overview

Vigilant is distributed as a standalone desktop application. This guide covers building, packaging, and distributing the application for Windows and macOS.

## Build Requirements

### All Platforms
- Go 1.23+
- Node.js 18+
- Wails CLI v2.x

### macOS
- Xcode Command Line Tools
- macOS 12+ for building

### Windows
- Visual Studio Build Tools (for CGO)
- WebView2 SDK (auto-installed by Wails)

## Build Commands

### Development Build

```bash
# Start development server with hot reload
make dev
```

### Production Builds

```bash
# Auto-detect OS and build
make build

# Platform-specific builds
make build-darwin           # macOS Intel (amd64)
make build-darwin-universal # macOS Universal (Intel + Apple Silicon)
make build-windows          # Windows (amd64)
```

### Build Outputs

```
build/
├── bin/
│   ├── vigilant.app/          # macOS application bundle
│   │   └── Contents/
│   │       ├── MacOS/
│   │       │   └── vigilant   # Main executable
│   │       ├── Resources/
│   │       │   └── iconfile.icns  # App icon
│   │       └── Info.plist     # macOS metadata
│   └── vigilant.exe           # Windows executable
```

## Installation

### macOS Installation

```bash
# Build and install to /Applications
make install

# Manual installation
cp -r build/bin/vigilant.app /Applications/

# Clear quarantine attribute (if needed)
xattr -cr /Applications/vigilant.app
```

### Windows Installation

```bash
# Build and install to Program Files
make install

# This creates:
# - C:\Program Files\Vigilant\vigilant.exe
# - Start Menu shortcut
```

### Manual Windows Installation

1. Build the application: `make build-windows`
2. Copy `build\bin\vigilant.exe` to desired location
3. Optionally create a desktop shortcut

## Wails Configuration

The `wails.json` file configures the build:

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "vigilant",
  "outputfilename": "vigilant",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "assetdir": "./assets",
  "author": {
    "name": "mgiovani",
    "email": "e@giovani.dev"
  },
  "build": {
    "windows": {
      "icon": "build/windows/icon.ico"
    },
    "darwin": {
      "icon": "build/darwin/icon.icns"
    }
  }
}
```

## Asset Bundling

### Bundled Assets
Assets in the `assets/` directory are embedded in the binary:
- `assets/fbi-meme.mp4` - FBI meme video

### Configuration Files
Default configuration is bundled:
- `config/default.yaml` - Default blocklist and settings

## Release Process

### 1. Version Bump

Update version in relevant files:
- `wails.json` (if using version field)
- `README.md` (badges, if any)

### 2. Build All Platforms

```bash
# On macOS
make build-darwin-universal

# On Windows (or cross-compile)
make build-windows
```

### 3. Create Release Artifacts

```bash
# macOS: Create DMG (optional)
hdiutil create -volname "Vigilant" -srcfolder build/bin/vigilant.app -ov -format UDZO Vigilant-macos.dmg

# Windows: Zip the executable
zip -j Vigilant-windows.zip build/bin/vigilant.exe
```

### 4. GitHub Release

1. Create a new release tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. Create GitHub release with:
   - Release notes
   - `Vigilant-macos.dmg` or `vigilant.app.zip`
   - `Vigilant-windows.zip`

## Cross-Compilation

### Building Windows from macOS

Cross-compilation requires Docker with xgo:

```bash
# Install xgo (one-time)
docker pull crazymax/xgo

# Cross-compile for Windows
xgo -targets=windows/amd64 -out vigilant .
```

### Building macOS from Linux/Windows

Not recommended - macOS builds should be done on macOS for proper code signing and notarization.

## Code Signing (Optional)

### macOS Code Signing

```bash
# Sign the application
codesign --force --deep --sign "Developer ID Application: Your Name" build/bin/vigilant.app

# Verify signature
codesign --verify --verbose build/bin/vigilant.app
```

### macOS Notarization

```bash
# Submit for notarization
xcrun notarytool submit Vigilant-macos.dmg --apple-id "your@email.com" --password "@keychain:AC_PASSWORD" --team-id "TEAM_ID" --wait

# Staple the notarization ticket
xcrun stapler staple Vigilant-macos.dmg
```

### Windows Code Signing

Use signtool with a code signing certificate:

```bash
signtool sign /a /t http://timestamp.digicert.com /fd SHA256 build\bin\vigilant.exe
```

## Troubleshooting

### macOS: "App is damaged" Error

```bash
# Clear quarantine attribute
xattr -cr /Applications/vigilant.app
```

### macOS: Accessibility Permissions

Vigilant needs accessibility permissions to monitor windows. macOS will prompt on first launch.

Manual setup:
1. Open System Preferences → Security & Privacy → Privacy
2. Select "Accessibility"
3. Add vigilant.app

### Windows: WebView2 Missing

WebView2 is required and should auto-install. If missing:
1. Download from [Microsoft WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)
2. Run the installer
3. Restart Vigilant

### Build Fails: CGO Issues

Ensure CGO is enabled and compiler is available:

```bash
# Check CGO status
go env CGO_ENABLED

# macOS: Install Xcode CLI tools
xcode-select --install

# Windows: Install Visual Studio Build Tools
```

## Performance Optimization

### Binary Size Reduction

The Wails build includes size optimizations:
```bash
wails build -ldflags="-s -w"
```

Current binary sizes:
- macOS Universal: ~15MB
- Windows amd64: ~12MB

### Startup Optimization

- Configuration is loaded synchronously at startup
- YouTube player loads asynchronously
- Monitor starts after UI is ready

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `YOUTUBE_API_KEY` | YouTube Data API key for dynamic lofi stream search | (optional, falls back to hardcoded video ID) |

## Monitoring and Logging

Logs are written to stderr with timestamps:
```
[vigilant] 2025/12/11 12:00:00 App initialized successfully
[vigilant] 2025/12/11 12:00:00 Monitor started
```

In production, logs can be captured:
```bash
# macOS
./vigilant.app/Contents/MacOS/vigilant 2>&1 | tee vigilant.log

# Windows
vigilant.exe 2>&1 > vigilant.log
```

---

**Last Updated**: 2025-12-11
