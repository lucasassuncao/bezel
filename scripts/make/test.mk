##@ Test

.PHONY: test test-watch test-coverage bench

## Run tests with gotestsum (testdox format)
test: $(GOTESTSUM)
	@$(GOTESTSUM) --format testdox -- -race ./...

## Run tests in watch mode (reruns on file changes)
test-watch: $(GOTESTSUM)
	@$(GOTESTSUM) --format testdox --watch -- -race ./...

## Run tests with coverage (HTML + Cobertura XML)
test-coverage: $(GOTESTSUM) $(GOCOBERTURA)
	@$(call script,test-coverage) $(GOTESTSUM) $(GOCOBERTURA) $(COVERAGE_DIR)

# Double quotes: cmd.exe keeps single quotes as part of the argument.
## Run benchmarks (tree, list, draw, table)
bench:
	@go test -run "^$$" -bench . -benchmem ./...
