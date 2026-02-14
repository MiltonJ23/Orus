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

# Run tests
test:
	@go test ./...

# Format code (Go's version of linting tasks)
fmt:
	@go fmt ./...


test-with-coverage:
	@go test -coverprofile=coverage.out ./...


consult-coverage:
	@go tool cover -func=coverage.out
