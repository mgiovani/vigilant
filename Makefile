.PHONY: dev build-windows build-darwin build-darwin-universal clean test help

help:
	@echo "Vigilant Desktop Application - Available Commands"
	@echo ""
	@echo "  make dev                      Run development server with hot reload"
	@echo "  make build-windows            Build Windows executable (amd64)"
	@echo "  make build-darwin             Build macOS executable (amd64)"
	@echo "  make build-darwin-universal   Build macOS universal executable (amd64 + arm64)"
	@echo "  make clean                    Remove build artifacts"
	@echo "  make test                     Run Go tests"
	@echo "  make help                     Show this help message"

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
