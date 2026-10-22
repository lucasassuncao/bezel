##@ Security

.PHONY: security govulncheck sbom vuln

## Run security analysis with gosec
security: $(GOSEC)
	@$(GOSEC) -stdout -severity medium ./...

# gosec reads this repository's code; govulncheck reads the dependency tree the
# way the compiler does, so it only reports advisories whose vulnerable symbol
# is actually reachable from this module. For a library that matters twice
# over: every consumer inherits this dependency tree.

## Report vulnerabilities reachable from this code
govulncheck: $(GOVULNCHECK)
	@echo "Checking for reachable vulnerabilities..."
	@$(GOVULNCHECK) ./...

# examples/ is excluded because its dependencies are the example apps', not
# this library's, and a stale test.exe left there would put whatever it was
# built against into the report. What consumers inherit is the tree ./go.mod
# declares, and nothing else.

# Double quotes: cmd.exe keeps single quotes as part of the argument.
## Generate a CycloneDX SBOM of the dependency tree
sbom: $(SYFT)
	@echo "Generating SBOM..."
	@$(SYFT) . --source-name $(MODULE_NAME) --exclude "./.gobin/**" --exclude "./examples/**" -o cyclonedx-json=$(SBOM_FILE)

## Scan the SBOM for known vulnerabilities with grype
vuln: $(GRYPE) sbom
	@echo "Scanning for vulnerabilities..."
	@$(GRYPE) sbom:$(SBOM_FILE) --fail-on medium
