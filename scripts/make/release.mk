##@ Release

.PHONY: notes release-notes tag

# The tag at HEAD, so it matches what `make tag` wrote into the tag message.
## Write the git-cliff notes of the tag at HEAD into release-notes.md
release-notes: $(GITCLIFF)
	@$(GITCLIFF) --current --strip all -o $(RELEASE_NOTES)

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
