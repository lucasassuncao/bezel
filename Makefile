.PHONY: help fmt lint test test-watch test-coverage security govulncheck sbom vuln deps docs tag all tools clean-tools clean

# Tool versions
GOLANGCI_LINT_VERSION := v2.13.2
GOMARKDOC_VERSION     := v1.1.0
GOTESTSUM_VERSION     := v1.13.0
GOSEC_VERSION         := v2.29.0
GOCOBERTURA_VERSION   := latest
SYFT_VERSION          := v1.51.1
GRYPE_VERSION         := v0.118.0
GOVULNCHECK_VERSION   := v1.7.0

# Tools are installed into ./.gobin (gitignored) rather than taken from PATH,
# so every target runs the version pinned above and never whatever happens to
# be installed system-wide. Each tool is a file target: the first target that
# needs it installs it, later runs reuse it. Bumping a version above does not
# invalidate an already installed binary - run `make clean-tools` first.
GOBIN_DIR := $(CURDIR)/.gobin

ifeq ($(OS),Windows_NT)
EXE := .exe
else
EXE :=
endif

GOLANGCI    := $(GOBIN_DIR)/golangci-lint$(EXE)
GOMARKDOC   := $(GOBIN_DIR)/gomarkdoc$(EXE)
GOTESTSUM   := $(GOBIN_DIR)/gotestsum$(EXE)
GOSEC       := $(GOBIN_DIR)/gosec$(EXE)
GOCOBERTURA := $(GOBIN_DIR)/gocover-cobertura$(EXE)
SYFT        := $(GOBIN_DIR)/syft$(EXE)
GRYPE       := $(GOBIN_DIR)/grype$(EXE)
GOVULNCHECK := $(GOBIN_DIR)/govulncheck$(EXE)

TOOLS := $(GOLANGCI) $(GOMARKDOC) $(GOTESTSUM) $(GOSEC) $(GOCOBERTURA) $(SYFT) $(GRYPE) $(GOVULNCHECK)

# `go install` takes its destination from GOBIN in the environment and has no
# flag for it, so the tool rules - and only those - export it.
$(TOOLS): export GOBIN = $(GOBIN_DIR)

$(GOLANGCI):
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(GOMARKDOC):
	@echo "Installing gomarkdoc $(GOMARKDOC_VERSION)..."
	@go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@$(GOMARKDOC_VERSION)

$(GOTESTSUM):
	@echo "Installing gotestsum $(GOTESTSUM_VERSION)..."
	@go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)

$(GOSEC):
	@echo "Installing gosec $(GOSEC_VERSION)..."
	@go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)

$(GOCOBERTURA):
	@echo "Installing gocover-cobertura $(GOCOBERTURA_VERSION)..."
	@go install github.com/t-yuki/gocover-cobertura@$(GOCOBERTURA_VERSION)

$(SYFT):
	@echo "Installing syft $(SYFT_VERSION)..."
	@go install github.com/anchore/syft/cmd/syft@$(SYFT_VERSION)

$(GRYPE):
	@echo "Installing grype $(GRYPE_VERSION)..."
	@go install github.com/anchore/grype/cmd/grype@$(GRYPE_VERSION)

$(GOVULNCHECK):
	@echo "Installing govulncheck $(GOVULNCHECK_VERSION)..."
	@go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

# Coverage
COVERAGE_OUT  := coverage.out
COVERAGE_HTML := coverage.html
COVERAGE_XML  := coverage.xml

# Project variables
MODULE_NAME := bezel
SBOM_FILE   := sbom.json

# Recipes run through whatever shell make finds, and on Windows that is sh.exe
# only when one is on PATH - cmd.exe otherwise. A recipe line has to be valid
# in both, which is why the help text is branched rather than written once with
# grep and awk.
help: ## Show this help message
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command "Select-String -Path $(MAKEFILE_LIST) -Pattern '^([a-zA-Z_-]+):.*## (.+)' | Sort-Object { $$_.Matches[0].Groups[1].Value } | ForEach-Object { Write-Host -NoNewline -ForegroundColor Cyan ('{0,-20}' -f $$_.Matches[0].Groups[1].Value); Write-Host (' ' + $$_.Matches[0].Groups[2].Value) }"
else
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
endif

fmt: ## Format code
	@go fmt ./...

lint: $(GOLANGCI) ## Run linter checks
	@$(GOLANGCI) -v run ./...

test: $(GOTESTSUM) ## Run tests with gotestsum (testdox format)
	@$(GOTESTSUM) --format testdox -- -race ./...

test-watch: $(GOTESTSUM) ## Run tests in watch mode (reruns on file changes)
	@$(GOTESTSUM) --format testdox --watch -- -race ./...

test-coverage: $(GOTESTSUM) $(GOCOBERTURA) ## Run tests with coverage (HTML + Cobertura XML)
	@$(GOTESTSUM) --format testdox -- -race -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./...
	@go tool cover -func=$(COVERAGE_OUT) | tail -1
	@go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@$(GOCOBERTURA) < $(COVERAGE_OUT) > $(COVERAGE_XML)
	@echo "Reports: $(COVERAGE_HTML) | $(COVERAGE_XML)"

security: $(GOSEC) ## Run security analysis with gosec
	@$(GOSEC) -stdout -severity medium ./...

# gosec reads this repository's code; govulncheck reads the dependency tree the
# way the compiler does, so it only reports advisories whose vulnerable symbol
# is actually reachable from this module. For a library that matters twice
# over: every consumer inherits this dependency tree.
govulncheck: $(GOVULNCHECK) ## Report vulnerabilities reachable from this code
	@echo "Checking for reachable vulnerabilities..."
	@$(GOVULNCHECK) ./...

# examples/ is excluded because its dependencies are the example apps', not
# this library's, and a stale test.exe left there would put whatever it was
# built against into the report. What consumers inherit is the tree ./go.mod
# declares, and nothing else.
sbom: $(SYFT) ## Generate a CycloneDX SBOM of the dependency tree
	@echo "Generating SBOM..."
	@$(SYFT) . --source-name $(MODULE_NAME) --exclude './.gobin/**' --exclude './examples/**' -o cyclonedx-json=$(SBOM_FILE)

vuln: $(GRYPE) sbom ## Scan the SBOM for known vulnerabilities with grype
	@echo "Scanning for vulnerabilities..."
	@$(GRYPE) sbom:$(SBOM_FILE) --fail-on medium

deps: ## Download and tidy dependencies
	@go mod download
	@go mod tidy

docs: $(GOMARKDOC) ## Generate documentation with gomarkdoc
	@$(GOMARKDOC) -e \
		--repository.url https://github.com/lucasassuncao/bezel \
		--repository.default-branch main \
		--repository.path / \
		-o '{{.Dir}}/README.md' \
		./animation/... \
		./browser/... \
		./draw/... \
		./form/... \
		./icon/... \
		./inline/... \
		./layout/... \
		./legend/... \
		./list/... \
		./overlay/... \
		./palette/... \
		./progress/... \
		./shell/... \
		./table/... \
		./textbox/... \
		./theme/... \
		./themebrowser/... \
		./tree/...

tag: ## Create and push an annotated git tag (usage: make tag VERSION=v1.2.3)
ifndef VERSION
	$(error Usage: make tag VERSION=v1.2.3)
endif
	git diff --exit-code --quiet
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Local, deterministic checks first, so a failure points at the code. The
# supply-chain scans go last because they are the only steps that need the
# network: govulncheck and grype both fetch a vulnerability database, and a
# hiccup there should not mask a lint or test failure. govulncheck runs before
# grype so the reachability answer is printed even when grype fails the build
# on an advisory nothing here can execute.
all: deps fmt docs lint security test-coverage govulncheck sbom vuln ## Run every check

tools: $(TOOLS) ## Install every pinned tool into ./.gobin
	@echo "Tools installed in $(GOBIN_DIR)"

clean-tools: ## Remove ./.gobin so the next target reinstalls the pinned tools
	@echo "Removing $(GOBIN_DIR)..."
	@rm -rf $(GOBIN_DIR)

clean: clean-tools ## Remove coverage artifacts, installed tools and cache
	@rm -rf $(COVERAGE_OUT) $(COVERAGE_HTML) $(COVERAGE_XML) $(SBOM_FILE)
	@go clean -cache -testcache
