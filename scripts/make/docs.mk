##@ Docs

.PHONY: docs stage-docs serve-docs

# Every package but the root, whose README.md is written by hand, bezeltest and
# the examples.
DOC_PACKAGES := \
	./animation/... \
	./component/... \
	./draw/... \
	./effect/... \
	./key/... \
	./layout/... \
	./legend/... \
	./overlay/... \
	./shell \
	./theme/...

## Generate documentation with gomarkdoc
docs: $(GOMARKDOC)
	@$(call script,package-docs) $(GOMARKDOC) https://github.com/lucasassuncao/bezel main $(DOC_PACKAGES)

# The Pages job in CI runs this too, so the local preview and the published site
# are built from the same layout.
## Assemble the docs site into ./site (docs/pages, the guides and the recordings)
stage-docs:
	@$(call script,stage-docs) docs $(SITE_DIR)

# docsify fetches the .md files at runtime, which browsers block on file://, so
# the site needs a server.
## Stage the docs site and serve it on http://localhost:8080 (Ctrl+C to stop)
serve-docs: $(CADDY) stage-docs
	@echo "Landing page: http://localhost:8080/"
	@echo "Guides:       http://localhost:8080/guide.html"
	@$(CADDY) file-server --root $(SITE_DIR) --listen :8080
