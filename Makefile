.PHONY: dev build build-windows build-darwin build-darwin-universal install clean test help setup

# Detect OS
UNAME_S := $(shell uname -s)

help:
	@echo "Vigilant Desktop Application - Available Commands"
	@echo ""
	@echo "  make setup                    Install all dependencies (Go + frontend)"
	@echo "  make dev                      Run development server with hot reload"
	@echo "  make build                    Build for current OS (auto-detected)"
	@echo "  make install                  Setup, build, and install app to system"
	@echo "  make test                     Run Go tests"
	@echo "  make clean                    Remove build artifacts"
	@echo ""
	@echo "  Platform-specific builds:"
	@echo "  make build-darwin             Build macOS executable (amd64)"
	@echo "  make build-darwin-universal   Build macOS universal executable (amd64 + arm64)"
	@echo "  make build-windows            Build Windows executable (amd64)"

setup:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "✓ Setup complete"

dev: setup
	wails dev

# Auto-detect OS and build accordingly
build:
ifeq ($(UNAME_S),Darwin)
	@echo "Detected macOS - building universal binary..."
	wails build -platform darwin/universal
else ifeq ($(UNAME_S),Linux)
	@echo "Detected Linux - building linux/amd64..."
	wails build -platform linux/amd64
else
	@echo "Detected Windows - building windows/amd64..."
	wails build -platform windows/amd64
endif

# Install app to system (setup + build + install)
install: setup build
ifeq ($(UNAME_S),Darwin)
	@echo "Installing vigilant.app to /Applications..."
	@rm -rf /Applications/vigilant.app
	@cp -r build/bin/vigilant.app /Applications/
	@xattr -cr /Applications/vigilant.app
	@echo "✓ Vigilant installed to /Applications/vigilant.app"
else ifeq ($(OS),Windows_NT)
	@echo "Installing Vigilant to Program Files..."
	@if not exist "C:\Program Files\Vigilant" mkdir "C:\Program Files\Vigilant"
	@copy /Y "build\bin\vigilant.exe" "C:\Program Files\Vigilant\vigilant.exe"
	@echo "Creating Start Menu shortcut..."
	@powershell -Command "$$WshShell = New-Object -ComObject WScript.Shell; $$Shortcut = $$WshShell.CreateShortcut(\"$$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Vigilant.lnk\"); $$Shortcut.TargetPath = 'C:\Program Files\Vigilant\vigilant.exe'; $$Shortcut.Save()"
	@echo "✓ Vigilant installed to C:\Program Files\Vigilant\"
	@echo "✓ Start Menu shortcut created"
else
	@echo "Install command currently only supported on macOS and Windows"
endif

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
