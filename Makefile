.PHONY: all build clean test fmt vet lint install help

# Variables
BINARY_CLIENT=ensemble
BINARY_SERVER=ensemble-server
BIN_DIR=bin
GO=go
GOFLAGS=-v

# Build all binaries
all: build

# Build both client and server
build: build-client build-server

# Build client binary
build-client:
	@echo "Building $(BINARY_CLIENT)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_CLIENT) ./cmd/ensemble

# Build server binary
build-server:
	@echo "Building $(BINARY_SERVER)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_SERVER) ./cmd/ensemble-server

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f $(BINARY_CLIENT) $(BINARY_SERVER)
	@$(GO) clean

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# Install binaries to $GOPATH/bin
install:
	@echo "Installing binaries..."
	$(GO) install ./cmd/ensemble
	$(GO) install ./cmd/ensemble-server

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GO) mod verify

# Run the server (development)
run-server: build-server
	@echo "Running server..."
	./$(BIN_DIR)/$(BINARY_SERVER)

# Run the client (development)
run-client: build-client
	@echo "Running client..."
	./$(BIN_DIR)/$(BINARY_CLIENT)

# Help
help:
	@echo "Ensemble - Multi-Agent Coordination Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build both client and server binaries"
	@echo "  make build-client - Build client binary only"
	@echo "  make build-server - Build server binary only"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Run go vet"
	@echo "  make lint         - Run golangci-lint"
	@echo "  make install      - Install binaries to GOPATH/bin"
	@echo "  make tidy         - Tidy dependencies"
	@echo "  make deps         - Download dependencies"
	@echo "  make verify       - Verify dependencies"
	@echo "  make run-server   - Build and run server"
	@echo "  make run-client   - Build and run client"
	@echo "  make help         - Show this help message"
