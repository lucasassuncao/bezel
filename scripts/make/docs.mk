##@ Docs

.PHONY: docs

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
	./palette/... \
	./shell/... \
	./theme/...

## Generate documentation with gomarkdoc
docs: $(GOMARKDOC)
	@$(call script,package-docs) $(GOMARKDOC) https://github.com/lucasassuncao/bezel main $(DOC_PACKAGES)
