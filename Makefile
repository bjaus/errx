.PHONY: test lint cover build ci clean generate help

## test: Run tests
test:
	go test -race ./...

## lint: Run golangci-lint
lint:
	golangci-lint run

## cover: Run tests with coverage
cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

## cover-html: Generate HTML coverage report
cover-html: cover
	go tool cover -html=coverage.out -o coverage.html

## build: Build the package
build:
	go build ./...

## generate: Run go generate
generate:
	go generate ./...

## ci: Run all CI checks
ci: lint test build

## clean: Remove generated files
clean:
	rm -f coverage.out coverage.html

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
