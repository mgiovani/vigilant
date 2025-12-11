.PHONY: dev build build-windows build-darwin build-darwin-universal install clean test help

# Detect OS
UNAME_S := $(shell uname -s)

help:
	@echo "Vigilant Desktop Application - Available Commands"
	@echo ""
	@echo "  make dev                      Run development server with hot reload"
	@echo "  make build                    Build for current OS (auto-detected)"
	@echo "  make build-windows            Build Windows executable (amd64)"
	@echo "  make build-darwin             Build macOS executable (amd64)"
	@echo "  make build-darwin-universal   Build macOS universal executable (amd64 + arm64)"
	@echo "  make install                  Build and install app (macOS: copy to /Applications)"
	@echo "  make clean                    Remove build artifacts"
	@echo "  make test                     Run Go tests"
	@echo "  make help                     Show this help message"

dev:
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

# Install app to system (macOS only for now)
install: build
ifeq ($(UNAME_S),Darwin)
	@echo "Installing Vigilant.app to /Applications..."
	@rm -rf /Applications/Vigilant.app
	@cp -r build/bin/Vigilant.app /Applications/
	@echo "✓ Vigilant installed to /Applications/Vigilant.app"
else
	@echo "Install command currently only supported on macOS"
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
