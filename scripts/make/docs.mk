##@ Docs

.PHONY: docs

# Every package but the root, whose README.md is written by hand, bezeltest and
# the examples.
DOC_PACKAGES := \
	./animation/... \
	./browser/... \
	./chart/... \
	./draw/... \
	./form/... \
	./icon/... \
	./inline/... \
	./key/... \
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

## Generate documentation with gomarkdoc
docs: $(GOMARKDOC)
	@$(call script,package-docs) $(GOMARKDOC) https://github.com/lucasassuncao/bezel main $(DOC_PACKAGES)
