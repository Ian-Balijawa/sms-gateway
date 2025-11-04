.PHONY: build run test clean deps install

# Build the application
build:
	@echo "Building SMS Gateway..."
	@go build -o sms-gateway main.go

# Run the application
run:
	@echo "Running SMS Gateway..."
	@go run main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f sms-gateway
	@rm -f *.db
	@go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Install dependencies
install: deps
	@echo "Installing dependencies..."

# Run with development mode
dev:
	@echo "Running in development mode..."
	@GIN_MODE=debug go run main.go

# Build for production
prod: build
	@echo "Build complete. Run with: ./sms-gateway"

