##@ Release

.PHONY: notes tag

## Print the git-cliff notes that `make tag` would write for the untagged commits
notes: $(GITCLIFF)
	@$(GITCLIFF) --unreleased --strip all

## Create and push an annotated git tag with git-cliff notes (usage: make tag VERSION=v1.2.3)
tag: $(GITCLIFF)
ifndef VERSION
	$(error Usage: make tag VERSION=v1.2.3)
endif
	git diff --exit-code --quiet
	@$(call script,tag) $(GITCLIFF) $(VERSION)
	git push origin $(VERSION)
