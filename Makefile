.PHONY: build run test test-unit test-e2e test-all lint swagger tidy clean

# Build
build:
	go build -o bin/url-shortener ./cmd/api/...

# Run
run:
	go run ./cmd/api/...

# Unit tests (all packages, race detector)
test-unit:
	go test -race -count=1 ./...

# E2E / integration tests
test-e2e:
	go test -race -count=1 -tags=e2e ./internal/integration/ -v

# All tests
test-all:
	go test -race -count=1 ./...

# Lint
lint:
	go vet ./...

# Swagger
swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Dependency management
tidy:
	go mod tidy

# Coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html to view report"

# Generate swagger + run tests
check: swagger test-all lint
	@echo "All checks passed"

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html docs/
