##@ Go

.PHONY: deps fmt lint

## Download and tidy dependencies
deps:
	@go mod download
	@go mod tidy

## Format code
fmt:
	@go fmt ./...

## Run linter checks
lint: $(GOLANGCI)
	@$(GOLANGCI) -v run ./...
