.PHONY: build run clean test lint deps

# Build the application
build:
	go build -o bin/spacebridge ./cmd/spacebridge

# Run the application
run: build
	./bin/spacebridge

# Run discovery commands
discover-spaces: build
	./bin/spacebridge discover spaces

discover-stacks: build
	./bin/spacebridge discover stacks

discover-contexts: build
	./bin/spacebridge discover contexts

discover-policies: build
	./bin/spacebridge discover policies

discover-all: build
	./bin/spacebridge discover all

# Export manifest
export: build
	./bin/spacebridge export -o manifest.json

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f manifest.json

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Verify the module
verify:
	go mod verify
