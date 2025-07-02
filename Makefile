.PHONY: run test build clean deps dev-setup

# Quick development run
run:
	@echo "🚀 Running Sage-Bitrix sync test..."
	go run cmd/test/main.go

# Development setup
dev-setup:
	@echo "⚙️  Setting up development environment..."
	cp .env.example .env 2>/dev/null || echo "Using existing .env"
	go mod tidy
	go mod download
	@echo "✅ Setup complete! Update .env with your database credentials"

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build executable
build:
	mkdir -p bin
	go build -o bin/test cmd/test/main.go

# Test with verbose output
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Clean build files
clean:
	rm -rf bin/

# Show environment variables (for debugging)
env:
	@echo "Current environment configuration:"
	@echo "SAGE_DB_HOST: $(shell grep SAGE_DB_HOST .env | cut -d '=' -f2)"
	@echo "SAGE_DB_NAME: $(shell grep SAGE_DB_NAME .env | cut -d '=' -f2)"
	@echo "SAGE_DB_USER: $(shell grep SAGE_DB_USER .env | cut -d '=' -f2)"

# Help
help:
	@echo "Available commands:"
	@echo "  run        - Run the test program"
	@echo "  dev-setup  - Initial development setup"
	@echo "  deps       - Install dependencies"
	@echo "  build      - Build executable"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  clean      - Clean build files"
	@echo "  env        - Show current config"