.PHONY: deps build install clean run help

# Default target
help:
	@echo "Osiris Lite CLI - Pure Go Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  deps     - Download and install Go dependencies"
	@echo "  build    - Build the osiris-lite binary to build/"
	@echo "  install  - Install osiris-lite globally (requires sudo)"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Run osiris-lite directly"
	@echo "  help     - Show this help"
	@echo ""
	@echo "Quick start:"
	@echo "  make deps build"
	@echo "  ./build/osiris-lite --help"
	@echo ""
	@echo "For Go package manager releases:"
	@echo "  git tag v1.0.0 && git push origin v1.0.0"
	@echo "  make build"
	@echo "  go install github.com/Enigma-Dark/osiris-lite@v1.0.0"

# Download dependencies
deps:
	@echo "Downloading Go dependencies..."
	go mod tidy
	go mod download

# Build the binary
build: deps
	@echo "Building Osiris Lite CLI..."
	mkdir -p build
	go build -o build/osiris-lite ./cmd
	@echo "Build complete! Binary: ./build/osiris-lite"

# Install globally  
install: build
	@echo "Installing osiris-lite globally..."
	sudo cp build/osiris-lite /usr/local/bin/
	@echo "Installed! You can now run 'osiris-lite' from anywhere."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/
	@echo "Clean complete!"

# Run directly without building
run: deps
	@echo "Running osiris-lite..."
	go run ./cmd $(ARGS)