# ==============================================================================
# Makefile — cpp-gen
# ==============================================================================
# Targets disponíveis:
#   make              → build (padrão)
#   make build        → compila o binário
#   make install      → instala em ~/.local/bin (sem sudo)
#   make install-global → instala em /usr/local/bin (requer sudo)
#   make uninstall    → remove o binário instalado
#   make clean        → remove artefatos de build locais
#   make help         → exibe esta mensagem
# ==============================================================================

# ── Metadados ─────────────────────────────────────────────────────────────────

BINARY     := cpp-gen
VERSION    := $(shell grep 'AppVersion' cmd/root.go 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "0.1.0")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# ── Diretórios ────────────────────────────────────────────────────────────────

# Diretório de instalação padrão (sem sudo, já suportado pela maioria dos shells)
INSTALL_DIR        := $(HOME)/.local/bin

# Diretório de instalação global (requer sudo)
INSTALL_DIR_GLOBAL := /usr/local/bin

# Diretório de saída do build local
BUILD_DIR  := ./dist

# ── Flags de compilação Go ────────────────────────────────────────────────────

# Injeta versão, commit e data no binário via ldflags para exibição no --version
LDFLAGS := -s -w \
	-X 'cpp-gen/cmd.AppVersion=$(VERSION)' \
	-X 'cpp-gen/cmd.BuildDate=$(BUILD_DATE)' \
	-X 'cpp-gen/cmd.GitCommit=$(GIT_COMMIT)'

GO      := go
GOFLAGS := -trimpath

# ── Cores para output ─────────────────────────────────────────────────────────

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

.PHONY: build install install-global uninstall clean help

# ── build ─────────────────────────────────────────────────────────────────────
## Compila o binário em ./dist/cpp-gen
build:
	@printf "$(BOLD)$(CYAN)  Building$(RESET)  $(BINARY) v$(VERSION) ($(GIT_COMMIT))\n"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) .
	@printf "$(BOLD)$(GREEN)  ✓ Built$(RESET)    $(BUILD_DIR)/$(BINARY)\n"

# ── install ───────────────────────────────────────────────────────────────────
## Instala o binário em ~/.local/bin (sem necessidade de sudo)
install: build
	@printf "\n$(BOLD)$(CYAN)  Installing$(RESET) $(BINARY) → $(INSTALL_DIR)/$(BINARY)\n"
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@chmod +x $(INSTALL_DIR)/$(BINARY)
	@printf "$(BOLD)$(GREEN)  ✓ Installed$(RESET) $(INSTALL_DIR)/$(BINARY)\n"
	@$(MAKE) --no-print-directory _post_install_log INSTALL_LOCATION=$(INSTALL_DIR)

# ── install-global ────────────────────────────────────────────────────────────
## Instala o binário em /usr/local/bin (requer sudo)
install-global: build
	@printf "\n$(BOLD)$(CYAN)  Installing$(RESET) $(BINARY) → $(INSTALL_DIR_GLOBAL)/$(BINARY)$(RESET) $(GRAY)(requer sudo)$(RESET)\n"
	@sudo cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR_GLOBAL)/$(BINARY)
	@sudo chmod +x $(INSTALL_DIR_GLOBAL)/$(BINARY)
	@printf "$(BOLD)$(GREEN)  ✓ Installed$(RESET) $(INSTALL_DIR_GLOBAL)/$(BINARY)\n"
	@$(MAKE) --no-print-directory _post_install_log INSTALL_LOCATION=$(INSTALL_DIR_GLOBAL)

# ── uninstall ─────────────────────────────────────────────────────────────────
## Remove o binário de ~/.local/bin e /usr/local/bin (se existirem)
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

# ── clean ─────────────────────────────────────────────────────────────────────
## Remove o diretório ./dist com os artefatos de build
clean:
	@printf "$(BOLD)$(CYAN)  Cleaning$(RESET)   $(BUILD_DIR)/\n"
	@rm -rf $(BUILD_DIR)
	@printf "$(BOLD)$(GREEN)  ✓ Done$(RESET)\n"

# ── help ──────────────────────────────────────────────────────────────────────
## Exibe todos os targets disponíveis
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
	@printf "  make INSTALL_DIR=~/bin install $(GRAY)# diretório customizado$(RESET)\n\n"

# ==============================================================================
# Target interno — Log pós-instalação com instruções de PATH
# ==============================================================================

INSTALL_LOCATION ?= $(INSTALL_DIR)

.PHONY: _post_install_log
_post_install_log:
	@printf "\n"
	@printf "$(BOLD)$(PURPLE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BOLD)$(PURPLE)║$(RESET)  $(BOLD)⚡ cpp-gen v$(VERSION) instalado com sucesso!$(RESET)              $(BOLD)$(PURPLE)║$(RESET)\n"
	@printf "$(BOLD)$(PURPLE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "\n"

# Verifica se o diretório de instalação já está no PATH
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

# ── Instruções bash / zsh ─────────────────────────────────────────────────────
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

# ── Instruções fish ───────────────────────────────────────────────────────────
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

# ── Próximos passos ───────────────────────────────────────────────────────────
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
