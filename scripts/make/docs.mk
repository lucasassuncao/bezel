##@ Docs

.PHONY: docs

# Every package but the root, whose README.md is written by hand, bezeltest and
# the examples.
DOC_PACKAGES := \
	./animation/... \
	./browser/... \
	./chart/... \
	./cursor/... \
	./draw/... \
	./form/... \
	./icon/... \
	./inline/... \
	./input/... \
	./key/... \
	./layout/... \
	./legend/... \
	./list/... \
	./overlay/... \
	./paginator/... \
	./palette/... \
	./progress/... \
	./shell/... \
	./spinner/... \
	./stopwatch/... \
	./table/... \
	./textbox/... \
	./theme/... \
	./themebrowser/... \
	./timer/... \
	./tree/...

## Generate documentation with gomarkdoc
docs: $(GOMARKDOC)
	@$(call script,package-docs) $(GOMARKDOC) https://github.com/lucasassuncao/bezel main $(DOC_PACKAGES)
