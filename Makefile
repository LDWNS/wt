# wt — worktree manager. Run `make help` for available targets.
# Typical setup on a new machine: make install

BINARY := wt
INSTALL_DIR := $(HOME)/go/bin
GOBIN := $(shell go env GOBIN 2>/dev/null)
ifneq ($(GOBIN),)
INSTALL_DIR := $(GOBIN)
endif

.PHONY: all build install uninstall deps clean help

.DEFAULT_GOAL := help

all: build

build: ## compile the wt binary
	go build -o $(BINARY) .

install: deps build ## build, check deps, and install binary + shell hooks
	mkdir -p $(INSTALL_DIR)
	install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo ""
	@echo "Installed $(BINARY) to $(INSTALL_DIR)"
	@echo ""
	@if ! echo "$$PATH" | tr ':' '\n' | grep -qx "$(INSTALL_DIR)"; then \
		echo "WARNING: $(INSTALL_DIR) not on PATH. Add to ~/.zshrc:"; \
		echo "  export PATH=\"$(INSTALL_DIR):\$$PATH\""; \
		echo ""; \
	fi
	@echo "Next steps (see README.md):"
	@echo "  1. Add the wt() shell wrapper to ~/.zshrc (cd support)"
	@echo "  2. Add: source <(wt completion zsh)"
	@echo "  3. source ~/.zshrc"

uninstall: ## remove installed binary
	rm -f $(INSTALL_DIR)/$(BINARY)

deps: ## verify required dependencies are present
	@command -v git >/dev/null 2>&1 || { echo "Missing dependency: git"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "Missing dependency: go (https://go.dev/doc/install)"; exit 1; }
	@command -v fzf >/dev/null 2>&1 || { echo "Missing dependency: fzf (brew install fzf)"; exit 1; }
	@echo "All dependencies present."

clean: ## remove built binary
	rm -f $(BINARY)

help: ## list available targets
	@echo "wt Makefile"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
