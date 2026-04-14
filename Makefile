BINARY_NAME=orus

APP_NAME=Orus
VERSION=0.1.0


# build the application

build:
	@echo "Building $(APP_NAME) version $(VERSION)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/orus/main.go
	@echo "Build completed. Binary: $(BINARY_NAME)"

# run the application
run: build
	@echo "Running $(APP_NAME)..."
	@./bin/$(BINARY_NAME)

# clean the build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/$(BINARY_NAME)
	@echo "Clean completed."

# Run tests (internal only — cmd/orus requires native UI libraries)
test:
	@go test -race -count=1 -timeout 120s ./internal/...

# Format code (Go's version of linting tasks)
fmt:
	@go fmt ./...

# Run go vet (internal only — cmd/orus requires native UI libraries)
vet:
	@go vet ./internal/...

# Run golangci-lint (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@golangci-lint run ./internal/...

test-with-coverage:
	@go test -coverprofile=coverage.out -covermode=atomic ./internal/...


consult-coverage:
	@go tool cover -func=coverage.out

.PHONY: build run clean test fmt vet lint test-with-coverage consult-coverage
