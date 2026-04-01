# ==============================================================================
# Makefile — cpp-gen
# ==============================================================================
# Available targets:
#   make              → build (default)
#   make build        → compiles the binary
#   make install      → installs to ~/.local/bin (no sudo required)
#   make install-global → installs to /usr/local/bin (requires sudo)
#   make uninstall    → removes the installed binary
#   make clean        → removes local build artifacts
#   make release      → calculates next version and creates tag via Conventional Commits
#   make snapshot     → local multi-platform build with goreleaser (without publishing)
#   make changelog    → generates/updates CHANGELOG.md via git-cliff
#   make help         → displays this message
# ==============================================================================

# ── Metadata ──────────────────────────────────────────────────────────────────

BINARY     := cpp-gen
VERSION    := $(shell grep 'AppVersion' cmd/root.go 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "0.1.0")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# ── Directories ───────────────────────────────────────────────────────────────

# Default installation directory (no sudo, supported by most shells)
INSTALL_DIR        := $(HOME)/.local/bin

# Global installation directory (requires sudo)
INSTALL_DIR_GLOBAL := /usr/local/bin

# Local build output directory
BUILD_DIR  := ./dist

# ── Go compilation flags ──────────────────────────────────────────────────────

# Injects version, commit and date into the binary via ldflags for --version display
LDFLAGS := -s -w \
	-X 'cpp-gen/cmd.AppVersion=$(VERSION)' \
	-X 'cpp-gen/cmd.BuildDate=$(BUILD_DATE)' \
	-X 'cpp-gen/cmd.GitCommit=$(GIT_COMMIT)'

GO      := go
GOFLAGS := -trimpath

# ── Output colors ─────────────────────────────────────────────────────────────

RESET   := \033[0m
BOLD    := \033[1m
GREEN   := \033[32m
CYAN    := \033[36m
YELLOW  := \033[33m
PURPLE  := \033[35m
RED     := \033[31m
GRAY    := \033[90m

# ==============================================================================
# Targets
# ==============================================================================

.DEFAULT_GOAL := build

.PHONY: build install install-global uninstall clean release snapshot changelog help

# ── build ─────────────────────────────────────────────────────────────────────
## Compiles the binary to ./dist/cpp-gen
build:
	@printf "$(BOLD)$(CYAN)  Building$(RESET)  $(BINARY) v$(VERSION) ($(GIT_COMMIT))\n"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) .
	@printf "$(BOLD)$(GREEN)  ✓ Built$(RESET)    $(BUILD_DIR)/$(BINARY)\n"

# ── install ───────────────────────────────────────────────────────────────────
## Installs the binary to ~/.local/bin (no sudo required)
install: build
	@mkdir -p $(INSTALL_DIR)
	@ACTION=instalado; \
	if [ -f "$(INSTALL_DIR)/$(BINARY)" ]; then \
		ACTION=atualizado; \
		OLD_VER=$$($(INSTALL_DIR)/$(BINARY) version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
		if [ -z "$$OLD_VER" ]; then OLD_VER="?"; fi; \
		printf "\n$(BOLD)$(CYAN)  Updating$(RESET)   $(BINARY) $(GRAY)v$$OLD_VER$(RESET) → $(BOLD)v$(VERSION)$(RESET) em $(INSTALL_DIR)/$(BINARY)\n"; \
	else \
		printf "\n$(BOLD)$(CYAN)  Installing$(RESET) $(BINARY) v$(VERSION) → $(INSTALL_DIR)/$(BINARY)\n"; \
	fi; \
	cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY); \
	chmod +x $(INSTALL_DIR)/$(BINARY); \
	if [ "$$ACTION" = "atualizado" ]; then \
		printf "$(BOLD)$(GREEN)  ✓ Updated$(RESET)   $(INSTALL_DIR)/$(BINARY)\n"; \
	else \
		printf "$(BOLD)$(GREEN)  ✓ Installed$(RESET) $(INSTALL_DIR)/$(BINARY)\n"; \
	fi; \
	$(MAKE) --no-print-directory _post_install_log INSTALL_LOCATION=$(INSTALL_DIR) INSTALL_ACTION=$$ACTION

# ── install-global ────────────────────────────────────────────────────────────
## Installs the binary to /usr/local/bin (requires sudo)
install-global: build
	@ACTION=instalado; \
	if [ -f "$(INSTALL_DIR_GLOBAL)/$(BINARY)" ]; then \
		ACTION=atualizado; \
		OLD_VER=$$($(INSTALL_DIR_GLOBAL)/$(BINARY) version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
		if [ -z "$$OLD_VER" ]; then OLD_VER="?"; fi; \
		printf "\n$(BOLD)$(CYAN)  Updating$(RESET)   $(BINARY) $(GRAY)v$$OLD_VER$(RESET) → $(BOLD)v$(VERSION)$(RESET) em $(INSTALL_DIR_GLOBAL)/$(BINARY) $(GRAY)(requer sudo)$(RESET)\n"; \
	else \
		printf "\n$(BOLD)$(CYAN)  Installing$(RESET) $(BINARY) v$(VERSION) → $(INSTALL_DIR_GLOBAL)/$(BINARY) $(GRAY)(requer sudo)$(RESET)\n"; \
	fi; \
	sudo cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR_GLOBAL)/$(BINARY); \
	sudo chmod +x $(INSTALL_DIR_GLOBAL)/$(BINARY); \
	if [ "$$ACTION" = "atualizado" ]; then \
		printf "$(BOLD)$(GREEN)  ✓ Updated$(RESET)   $(INSTALL_DIR_GLOBAL)/$(BINARY)\n"; \
	else \
		printf "$(BOLD)$(GREEN)  ✓ Installed$(RESET) $(INSTALL_DIR_GLOBAL)/$(BINARY)\n"; \
	fi; \
	$(MAKE) --no-print-directory _post_install_log INSTALL_LOCATION=$(INSTALL_DIR_GLOBAL) INSTALL_ACTION=$$ACTION

# ── uninstall ─────────────────────────────────────────────────────────────────
## Removes the binary from ~/.local/bin and /usr/local/bin (if they exist)
uninstall:
	@printf "$(BOLD)$(CYAN)  Uninstalling$(RESET) $(BINARY)...\n"
	@removed=0; \
	if [ -f "$(INSTALL_DIR)/$(BINARY)" ]; then \
		rm -f "$(INSTALL_DIR)/$(BINARY)"; \
		printf "$(BOLD)$(GREEN)  ✓ Removed$(RESET)   $(INSTALL_DIR)/$(BINARY)\n"; \
		removed=1; \
	fi; \
	if [ -f "$(INSTALL_DIR_GLOBAL)/$(BINARY)" ]; then \
		sudo rm -f "$(INSTALL_DIR_GLOBAL)/$(BINARY)"; \
		printf "$(BOLD)$(GREEN)  ✓ Removed$(RESET)   $(INSTALL_DIR_GLOBAL)/$(BINARY)\n"; \
		removed=1; \
	fi; \
	if [ $$removed -eq 0 ]; then \
		printf "$(YELLOW)  ⚠ Nenhuma instalação encontrada.$(RESET)\n"; \
	fi

# ── release ───────────────────────────────────────────────────────────────────
## Calculates the next version via Conventional Commits and creates the git tag
release:
	@chmod +x scripts/release.sh
	@./scripts/release.sh

# ── release (dry-run) ─────────────────────────────────────────────────────────
## Displays the calculated next version without creating any tag
release-dry:
	@chmod +x scripts/release.sh
	@./scripts/release.sh --dry-run

# ── snapshot ──────────────────────────────────────────────────────────────────
## Local multi-platform build with goreleaser (without publishing to GitHub)
snapshot:
	@printf "$(BOLD)$(CYAN)  Snapshot$(RESET)   goreleaser build local...\n"
	@command -v goreleaser >/dev/null 2>&1 || { \
		printf "$(RED)  ✗ goreleaser não encontrado.$(RESET)\n"; \
		printf "  Instale com: $(CYAN)go install github.com/goreleaser/goreleaser/v2@latest$(RESET)\n"; \
		exit 1; \
	}
	@goreleaser release --snapshot --clean
	@printf "$(BOLD)$(GREEN)  ✓ Snapshot$(RESET)  Binários em ./dist/\n"

# ── changelog ────────────────────────────────────────────────────────────────
## Generates or updates CHANGELOG.md from commits (requires git-cliff)
changelog:
	@printf "$(BOLD)$(CYAN)  Changelog$(RESET)  gerando CHANGELOG.md...\n"
	@command -v git-cliff >/dev/null 2>&1 || { \
		printf "$(RED)  ✗ git-cliff não encontrado.$(RESET)\n"; \
		printf "  Instale com: $(CYAN)cargo install git-cliff$(RESET)\n"; \
		printf "  Ou via:      $(CYAN)brew install git-cliff$(RESET)\n"; \
		exit 1; \
	}
	@git-cliff --output CHANGELOG.md
	@printf "$(BOLD)$(GREEN)  ✓ Gerado$(RESET)   CHANGELOG.md\n"

# ── clean ─────────────────────────────────────────────────────────────────────
## Removes the ./dist directory with build artifacts
clean:
	@printf "$(BOLD)$(CYAN)  Cleaning$(RESET)   $(BUILD_DIR)/\n"
	@rm -rf $(BUILD_DIR)
	@printf "$(BOLD)$(GREEN)  ✓ Done$(RESET)\n"

# ── help ──────────────────────────────────────────────────────────────────────
## Displays all available targets
help:
	@printf "\n$(BOLD)$(PURPLE)⚡ cpp-gen$(RESET) — Makefile\n\n"
	@printf "$(BOLD)Uso:$(RESET)  make $(CYAN)<target>$(RESET)\n\n"
	@printf "$(BOLD)Targets:$(RESET)\n"
	@awk '/^## / { desc=substr($$0, 4) } \
	      /^[a-zA-Z][a-zA-Z_-]*:/ && desc != "" { \
	          target=$$1; sub(/:.*/, "", target); \
	          printf "  \033[36m%-20s\033[0m %s\n", target, desc; \
	          desc="" \
	      }' \
	     $(MAKEFILE_LIST)
	@printf "\n$(BOLD)Variáveis:$(RESET)\n"
	@printf "  $(CYAN)%-20s$(RESET) %s\n" "INSTALL_DIR"        "$(INSTALL_DIR)"
	@printf "  $(CYAN)%-20s$(RESET) %s\n" "INSTALL_DIR_GLOBAL" "$(INSTALL_DIR_GLOBAL)"
	@printf "  $(CYAN)%-20s$(RESET) %s\n" "VERSION"            "$(VERSION)"
	@printf "  $(CYAN)%-20s$(RESET) %s\n" "GIT_COMMIT"         "$(GIT_COMMIT)"
	@printf "\n$(BOLD)Exemplos:$(RESET)\n"
	@printf "  make install                $(GRAY)# instala em ~/.local/bin$(RESET)\n"
	@printf "  make install-global         $(GRAY)# instala em /usr/local/bin$(RESET)\n"
	@printf "  make INSTALL_DIR=~/bin install $(GRAY)# diretório customizado$(RESET)\n"
	@printf "  make release                $(GRAY)# cria a próxima tag de release$(RESET)\n"
	@printf "  make release-dry            $(GRAY)# simula a próxima versão (dry-run)$(RESET)\n"
	@printf "  make snapshot               $(GRAY)# build local multi-plataforma$(RESET)\n"
	@printf "  make changelog              $(GRAY)# atualiza CHANGELOG.md$(RESET)\n\n"

# ==============================================================================
# Internal target — Post-install log with PATH instructions
# ==============================================================================

INSTALL_LOCATION ?= $(INSTALL_DIR)
INSTALL_ACTION   ?= instalado

.PHONY: _post_install_log
_post_install_log:
	@printf "\n"
	@printf "$(BOLD)$(PURPLE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BOLD)$(PURPLE)║$(RESET)  $(BOLD)⚡ cpp-gen v$(VERSION) $(INSTALL_ACTION) com sucesso!$(RESET)              $(BOLD)$(PURPLE)║$(RESET)\n"
	@printf "$(BOLD)$(PURPLE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "\n"

# Checks if the installation directory is already in PATH
	@if echo "$$PATH" | tr ':' '\n' | grep -qx "$(INSTALL_LOCATION)"; then \
		printf "$(BOLD)$(GREEN)  ✓ $(INSTALL_LOCATION)$(RESET) já está no seu PATH. Tudo pronto!\n\n"; \
		printf "  Execute: $(BOLD)$(CYAN)$(BINARY) --help$(RESET)\n\n"; \
	else \
		printf "$(BOLD)$(YELLOW)  ⚠ Ação necessária:$(RESET) adicione $(BOLD)$(INSTALL_LOCATION)$(RESET) ao seu PATH.\n"; \
		printf "  Siga as instruções abaixo para o seu shell:\n\n"; \
		$(MAKE) --no-print-directory _log_bash_zsh INSTALL_LOCATION=$(INSTALL_LOCATION); \
		$(MAKE) --no-print-directory _log_fish      INSTALL_LOCATION=$(INSTALL_LOCATION); \
		$(MAKE) --no-print-directory _log_next_steps; \
	fi

# ── bash / zsh instructions ───────────────────────────────────────────────────
.PHONY: _log_bash_zsh
_log_bash_zsh:
	@printf "$(BOLD)$(CYAN)  ┌─ Bash / Zsh $(GRAY)─────────────────────────────────────────────$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  Adicione ao final do $(BOLD)~/.bashrc$(RESET) $(GRAY)(bash)$(RESET) ou $(BOLD)~/.zshrc$(RESET) $(GRAY)(zsh)$(RESET):\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(GRAY)# cpp-gen$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(GREEN)export PATH=\"$(INSTALL_LOCATION):\$$PATH\"$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  Ou execute o comando abaixo para adicionar automaticamente:\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(GRAY)# bash:$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(YELLOW)echo 'export PATH=\"$(INSTALL_LOCATION):\$$PATH\"' >> ~/.bashrc$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(YELLOW)source ~/.bashrc$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(GRAY)# zsh:$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(YELLOW)echo 'export PATH=\"$(INSTALL_LOCATION):\$$PATH\"' >> ~/.zshrc$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(YELLOW)source ~/.zshrc$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  └─────────────────────────────────────────────────────────────$(RESET)\n"
	@printf "\n"

# ── fish instructions ─────────────────────────────────────────────────────────
.PHONY: _log_fish
_log_fish:
	@printf "$(BOLD)$(CYAN)  ┌─ Fish Shell $(GRAY)────────────────────────────────────────────$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(BOLD)Opção 1$(RESET) — Comando único $(GRAY)(persistente, recomendado)$(RESET):\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(YELLOW)fish_add_path $(INSTALL_LOCATION)$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(BOLD)Opção 2$(RESET) — Adicione ao $(BOLD)~/.config/fish/config.fish$(RESET):\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(GRAY)# cpp-gen$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(GREEN)fish_add_path $(INSTALL_LOCATION)$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  Ou com set -x para compatibilidade:\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(GREEN)set -x PATH $(INSTALL_LOCATION) \$$PATH$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  Ou execute o comando abaixo para adicionar automaticamente:\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)  $(YELLOW)echo 'fish_add_path $(INSTALL_LOCATION)' >> ~/.config/fish/config.fish$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  │$(RESET)\n"
	@printf "$(BOLD)$(CYAN)  └─────────────────────────────────────────────────────────────$(RESET)\n"
	@printf "\n"

# ── Next steps ────────────────────────────────────────────────────────────────
.PHONY: _log_next_steps
_log_next_steps:
	@printf "$(BOLD)  Após configurar o PATH, recarregue o shell e execute:$(RESET)\n"
	@printf "\n"
	@printf "  $(BOLD)$(CYAN)$(BINARY) --help$(RESET)          $(GRAY)# exibe a ajuda$(RESET)\n"
	@printf "  $(BOLD)$(CYAN)$(BINARY) new meu-projeto$(RESET)  $(GRAY)# cria um projeto interativamente$(RESET)\n"
	@printf "  $(BOLD)$(CYAN)$(BINARY) version$(RESET)          $(GRAY)# confirma a versão instalada$(RESET)\n"
	@printf "\n"
	@printf "  $(GRAY)Para desinstalar: $(RESET)$(YELLOW)make uninstall$(RESET)\n"
	@printf "\n"
