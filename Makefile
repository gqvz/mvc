# Get swag binary location from GOPATH
SWAG=$(shell go env GOPATH)/bin/swag

.PHONY: swagger build run test lint clean

# Generate Swagger docs
swagger:
	$(SWAG) init -g cmd/main.go

# Build the Go project
build:
	go build -o bin/app cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run tests
test:
	go test ./...

# Run linter (requires golangci-lint installed)
lint:
	golangci-lint run

# Remove build artifacts and generated docs
clean:
	rm -rf bin docs/swagger