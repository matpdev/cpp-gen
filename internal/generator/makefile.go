package generator

import "fmt"

// generateMakefile writes a Makefile to the project root.
// The generated file is a universal C++/CMake Makefile that wraps the most
// common build, test, lint and release workflows. It is intentionally
// project-agnostic: no cpp-gen-specific logic, only standard CMake + Go-style
// conventions that work for any C++ project.
func generateMakefile(root string, data *TemplateData, verbose bool) error {
	path := "Makefile"
	if err := writeTemplate(
		fmt.Sprintf("%s/%s", root, path),
		"Makefile",
		tmplMakefile,
		data,
		verbose,
	); err != nil {
		return fmt.Errorf("gerar Makefile: %w", err)
	}
	return nil
}

// tmplMakefile is the Go template for the universal C++/CMake Makefile.
//
// Template variables used (from TemplateData):
//
//	{{.Name}}         — project name in its original form (e.g. "my-app")
//	{{.NameSnake}}    — snake_case name           (e.g. "my_app")
//	{{.Version}}      — initial SemVer version     (e.g. "1.0.0")
//	{{.UseGit}}       — bool: whether git was initialised
//	{{.UseClangd}}    — bool: whether .clangd was generated
//	{{.UseClangFormat}}— bool: whether .clang-format was generated
const tmplMakefile = `# ==============================================================================
# Makefile — {{.Name}}
# ==============================================================================
#
# Targets
#   make / make build          Build (Debug preset)
#   make build-release         Build (Release preset)
#   make run                   Build + run the binary (Debug)
#   make test                  Build + run CTest suite
#   make test-verbose          Run CTest with verbose output
#   make configure             CMake configure only (Debug)
#   make configure-release     CMake configure only (Release)
#   make clean                 Remove the build/ directory
#   make lint                  Run clang-tidy on all sources
#   make format                Run clang-format (in-place) on all sources
#   make format-check          Check formatting without modifying files
#   make version               Show current project version
#   make version-bump-patch    Bump PATCH (x.y.Z) and commit + tag
#   make version-bump-minor    Bump MINOR (x.Y.0) and commit + tag
#   make version-bump-major    Bump MAJOR (X.0.0) and commit + tag
#   make changelog             Generate / update CHANGELOG.md (git-cliff)
#   make deps                  List all CMake targets and their dependencies
#   make info                  Print build environment summary
#   make help                  Show this message
# ==============================================================================

# ── Project metadata ──────────────────────────────────────────────────────────

PROJECT     := {{.Name}}
VERSION     := {{.Version}}
BUILD_DIR   := build

# ── CMake ─────────────────────────────────────────────────────────────────────

CMAKE       := cmake
CTEST       := ctest
PRESET_DBG  := debug
PRESET_REL  := release

# Fallback when CMakePresets.json is not present: manual flags
CMAKE_BUILD_TYPE_DBG := Debug
CMAKE_BUILD_TYPE_REL := Release
CMAKE_FLAGS          ?=

# ── Tools ─────────────────────────────────────────────────────────────────────

CLANG_TIDY   := clang-tidy
CLANG_FORMAT := clang-format

# Source files considered by lint / format targets
# Adjust the glob patterns to match your project layout.
SRC_FILES := $(shell find src include \
	-name '*.cpp' -o -name '*.hpp' -o -name '*.h' -o -name '*.cc' \
	2>/dev/null | sort)

# ── Output colors ─────────────────────────────────────────────────────────────

RESET  := \033[0m
BOLD   := \033[1m
GREEN  := \033[32m
CYAN   := \033[36m
YELLOW := \033[33m
PURPLE := \033[35m
RED    := \033[31m
GRAY   := \033[90m

# ── Version helpers ───────────────────────────────────────────────────────────
# Reads the current version from CMakeLists.txt (project(... VERSION x.y.z ...))
# and falls back to the hard-coded VERSION variable above.

CURRENT_VERSION := $(shell grep -m1 'project(' CMakeLists.txt 2>/dev/null \
	| grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "$(VERSION)")

_VER_MAJOR := $(word 1, $(subst ., ,$(CURRENT_VERSION)))
_VER_MINOR := $(word 2, $(subst ., ,$(CURRENT_VERSION)))
_VER_PATCH := $(word 3, $(subst ., ,$(CURRENT_VERSION)))

# ==============================================================================
# Default goal
# ==============================================================================

.DEFAULT_GOAL := build

.PHONY: build build-release run test test-verbose \
        configure configure-release \
        clean lint format format-check \
        version version-bump-patch version-bump-minor version-bump-major \
        changelog deps info help \
        _check_preset _bump _require_git_clean

# ==============================================================================
# Build
# ==============================================================================

## Build the project (Debug)
build: _check_preset
	@printf "$(BOLD)$(CYAN)  Building$(RESET)  $(PROJECT) v$(CURRENT_VERSION) [Debug]\n"
	@if [ -f CMakePresets.json ]; then \
		$(CMAKE) --build --preset $(PRESET_DBG); \
	else \
		$(MAKE) --no-print-directory configure; \
		$(CMAKE) --build $(BUILD_DIR)/debug -- -j$$(nproc 2>/dev/null || sysctl -n hw.logicalcpu 2>/dev/null || echo 4); \
	fi
	@printf "$(BOLD)$(GREEN)  ✓ Done$(RESET)\n"

## Build the project (Release, optimised)
build-release: _check_preset
	@printf "$(BOLD)$(CYAN)  Building$(RESET)  $(PROJECT) v$(CURRENT_VERSION) [Release]\n"
	@if [ -f CMakePresets.json ]; then \
		$(CMAKE) --build --preset $(PRESET_REL); \
	else \
		$(MAKE) --no-print-directory configure-release; \
		$(CMAKE) --build $(BUILD_DIR)/release -- -j$$(nproc 2>/dev/null || sysctl -n hw.logicalcpu 2>/dev/null || echo 4); \
	fi
	@printf "$(BOLD)$(GREEN)  ✓ Done$(RESET)\n"

# ── CMake configure (used as fallback when no presets are found) ───────────────

_check_preset:
	@if [ ! -f CMakePresets.json ] && [ ! -d "$(BUILD_DIR)/debug" ]; then \
		$(MAKE) --no-print-directory configure; \
	fi

## CMake configure — Debug
configure:
	@printf "$(BOLD)$(CYAN)  Configure$(RESET) Debug → $(BUILD_DIR)/debug\n"
	@$(CMAKE) -S . -B $(BUILD_DIR)/debug \
		-DCMAKE_BUILD_TYPE=$(CMAKE_BUILD_TYPE_DBG) \
		-DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
		$(CMAKE_FLAGS)
	@printf "$(BOLD)$(GREEN)  ✓ Configured$(RESET)\n"

## CMake configure — Release
configure-release:
	@printf "$(BOLD)$(CYAN)  Configure$(RESET) Release → $(BUILD_DIR)/release\n"
	@$(CMAKE) -S . -B $(BUILD_DIR)/release \
		-DCMAKE_BUILD_TYPE=$(CMAKE_BUILD_TYPE_REL) \
		$(CMAKE_FLAGS)
	@printf "$(BOLD)$(GREEN)  ✓ Configured$(RESET)\n"

# ==============================================================================
# Run
# ==============================================================================

## Build (Debug) and run the resulting binary
run: build
	@printf "$(BOLD)$(CYAN)  Running$(RESET)   $(PROJECT)\n"
	@if [ -f CMakePresets.json ]; then \
		BIN=$$(find $(BUILD_DIR) -maxdepth 4 -type f -name '$(PROJECT)' | head -1); \
	else \
		BIN=$$(find $(BUILD_DIR)/debug -maxdepth 4 -type f -name '$(PROJECT)' | head -1); \
	fi; \
	if [ -z "$$BIN" ]; then \
		printf "$(RED)  ✗ Binary not found.$(RESET) Build the project first.\n"; exit 1; \
	fi; \
	"$$BIN"

# ==============================================================================
# Test
# ==============================================================================

## Build and run the CTest suite (Debug)
test: build
	@printf "$(BOLD)$(CYAN)  Testing$(RESET)   $(PROJECT)\n"
	@if [ -f CMakePresets.json ]; then \
		$(CTEST) --preset $(PRESET_DBG) --output-on-failure; \
	else \
		cd $(BUILD_DIR)/debug && $(CTEST) --output-on-failure; \
	fi
	@printf "$(BOLD)$(GREEN)  ✓ Tests passed$(RESET)\n"

## Run CTest with verbose output
test-verbose: build
	@printf "$(BOLD)$(CYAN)  Testing$(RESET)   $(PROJECT) (verbose)\n"
	@if [ -f CMakePresets.json ]; then \
		$(CTEST) --preset $(PRESET_DBG) --output-on-failure -V; \
	else \
		cd $(BUILD_DIR)/debug && $(CTEST) --output-on-failure -V; \
	fi

# ==============================================================================
# Clean
# ==============================================================================

## Remove build artifacts
clean:
	@printf "$(BOLD)$(CYAN)  Cleaning$(RESET)  $(BUILD_DIR)/\n"
	@rm -rf $(BUILD_DIR)
	@printf "$(BOLD)$(GREEN)  ✓ Done$(RESET)\n"

# ==============================================================================
# Code quality
# ==============================================================================

## Run clang-tidy on all source files
lint:
	@printf "$(BOLD)$(CYAN)  Lint$(RESET)      clang-tidy\n"
	@command -v $(CLANG_TIDY) >/dev/null 2>&1 || { \
		printf "$(RED)  ✗ clang-tidy not found.$(RESET)\n"; \
		printf "  Install: $(CYAN)apt install clang-tidy$(RESET) / $(CYAN)brew install llvm$(RESET)\n"; \
		exit 1; \
	}
	@DB=$$(find $(BUILD_DIR) -name 'compile_commands.json' | head -1); \
	if [ -z "$$DB" ]; then \
		printf "$(YELLOW)  ⚠ compile_commands.json not found — run make configure first.$(RESET)\n"; \
		DB_FLAG=""; \
	else \
		DB_FLAG="-p $$DB"; \
	fi; \
	if [ -z "$(SRC_FILES)" ]; then \
		printf "$(YELLOW)  ⚠ No source files found in src/ or include/.$(RESET)\n"; exit 0; \
	fi; \
	$(CLANG_TIDY) $$DB_FLAG $(SRC_FILES)
	@printf "$(BOLD)$(GREEN)  ✓ Lint passed$(RESET)\n"

## Run clang-format in-place on all source files
format:
	@printf "$(BOLD)$(CYAN)  Format$(RESET)    clang-format (in-place)\n"
	@command -v $(CLANG_FORMAT) >/dev/null 2>&1 || { \
		printf "$(RED)  ✗ clang-format not found.$(RESET)\n"; exit 1; \
	}
	@if [ -z "$(SRC_FILES)" ]; then \
		printf "$(YELLOW)  ⚠ No source files found.$(RESET)\n"; exit 0; \
	fi
	@$(CLANG_FORMAT) -i $(SRC_FILES)
	@printf "$(BOLD)$(GREEN)  ✓ Formatted $(words $(SRC_FILES)) file(s)$(RESET)\n"

## Check formatting without modifying files (exits non-zero if changes needed)
format-check:
	@printf "$(BOLD)$(CYAN)  Format$(RESET)    clang-format (check only)\n"
	@command -v $(CLANG_FORMAT) >/dev/null 2>&1 || { \
		printf "$(RED)  ✗ clang-format not found.$(RESET)\n"; exit 1; \
	}
	@if [ -z "$(SRC_FILES)" ]; then \
		printf "$(YELLOW)  ⚠ No source files found.$(RESET)\n"; exit 0; \
	fi
	@DIFF=$$($(CLANG_FORMAT) --dry-run --Werror $(SRC_FILES) 2>&1); \
	if [ -n "$$DIFF" ]; then \
		printf "$(RED)  ✗ Formatting issues found.$(RESET) Run $(CYAN)make format$(RESET) to fix.\n"; \
		exit 1; \
	fi
	@printf "$(BOLD)$(GREEN)  ✓ All files correctly formatted$(RESET)\n"

# ==============================================================================
# Versioning
# ==============================================================================

## Show the current project version (read from CMakeLists.txt)
version:
	@printf "$(BOLD)$(PROJECT)$(RESET) v$(CURRENT_VERSION)\n"

## Bump PATCH version (x.y.Z+1), commit and create git tag
version-bump-patch:
	@$(MAKE) --no-print-directory _bump PART=patch

## Bump MINOR version (x.Y+1.0), commit and create git tag
version-bump-minor:
	@$(MAKE) --no-print-directory _bump PART=minor

## Bump MAJOR version (X+1.0.0), commit and create git tag
version-bump-major:
	@$(MAKE) --no-print-directory _bump PART=major

# ── Internal version bump ─────────────────────────────────────────────────────
# Usage: make _bump PART=patch|minor|major

PART ?= patch

_bump: _require_git_clean
	@case "$(PART)" in \
		patch) NEW_VER="$(_VER_MAJOR).$(_VER_MINOR).$$(( $(_VER_PATCH) + 1 ))" ;; \
		minor) NEW_VER="$(_VER_MAJOR).$$(( $(_VER_MINOR) + 1 )).0" ;; \
		major) NEW_VER="$$(( $(_VER_MAJOR) + 1 )).0.0" ;; \
		*) printf "$(RED)  ✗ Unknown PART=$(PART). Use patch|minor|major$(RESET)\n"; exit 1 ;; \
	esac; \
	printf "$(BOLD)$(CYAN)  Version$(RESET)   $(CURRENT_VERSION) → $$NEW_VER\n"; \
	sed -i.bak "s/\(project([^)]*VERSION \)[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*/\1$$NEW_VER/" CMakeLists.txt; \
	rm -f CMakeLists.txt.bak; \
	git add CMakeLists.txt; \
	git commit -m "chore: bump version to $$NEW_VER"; \
	git tag -a "v$$NEW_VER" -m "Release v$$NEW_VER"; \
	printf "$(BOLD)$(GREEN)  ✓ Tag v$$NEW_VER created.$(RESET)\n"; \
	printf "  Push with: $(CYAN)git push origin main --tags$(RESET)\n"

_require_git_clean:
	@if ! git diff --quiet HEAD 2>/dev/null; then \
		printf "$(RED)  ✗ Working tree is dirty.$(RESET) Commit or stash changes before bumping.\n"; \
		exit 1; \
	fi

# ==============================================================================
# Changelog
# ==============================================================================

## Generate / update CHANGELOG.md from Conventional Commits (requires git-cliff)
changelog:
	@printf "$(BOLD)$(CYAN)  Changelog$(RESET) generating CHANGELOG.md...\n"
	@command -v git-cliff >/dev/null 2>&1 || { \
		printf "$(RED)  ✗ git-cliff not found.$(RESET)\n"; \
		printf "  Install: $(CYAN)cargo install git-cliff$(RESET)"; \
		printf " / $(CYAN)brew install git-cliff$(RESET)\n"; \
		exit 1; \
	}
	@git-cliff --output CHANGELOG.md
	@printf "$(BOLD)$(GREEN)  ✓ Done$(RESET)    CHANGELOG.md updated\n"

# ==============================================================================
# Utilities
# ==============================================================================

## Print CMake targets and their file dependencies
deps:
	@printf "$(BOLD)$(CYAN)  Deps$(RESET)      cmake --graphviz\n"
	@$(CMAKE) --build $(BUILD_DIR)/debug --target help 2>/dev/null \
		|| printf "$(GRAY)  Run make configure first to list targets.$(RESET)\n"

## Print build environment summary
info:
	@printf "\n$(BOLD)$(PURPLE)  Project$(RESET)  $(PROJECT) v$(CURRENT_VERSION)\n"
	@printf "$(BOLD)  CMake$(RESET)    $$($(CMAKE) --version | head -1)\n"
	@printf "$(BOLD)  CTest$(RESET)    $$($(CTEST) --version | head -1)\n"
	@printf "$(BOLD)  CC$(RESET)       $${CC:-$$(command -v gcc || echo n/a)}\n"
	@printf "$(BOLD)  CXX$(RESET)      $${CXX:-$$(command -v g++ || echo n/a)}\n"
	@if command -v $(CLANG_TIDY) >/dev/null 2>&1; then \
		printf "$(BOLD)  Tidy$(RESET)     $$($(CLANG_TIDY) --version | head -1)\n"; \
	fi
	@if command -v $(CLANG_FORMAT) >/dev/null 2>&1; then \
		printf "$(BOLD)  Format$(RESET)   $$($(CLANG_FORMAT) --version)\n"; \
	fi
	@printf "$(BOLD)  Build$(RESET)    $(BUILD_DIR)/\n"
	@printf "\n"

# ==============================================================================
# Help
# ==============================================================================

## Show all available targets
help:
	@printf "\n$(BOLD)$(PURPLE)  $(PROJECT)$(RESET) — Makefile\n\n"
	@printf "$(BOLD)Usage:$(RESET)  make $(CYAN)<target>$(RESET)\n\n"
	@printf "$(BOLD)Targets:$(RESET)\n"
	@awk '/^## / { desc=substr($$0,4) } \
	      /^[a-zA-Z][a-zA-Z0-9_-]*:/ && desc != "" { \
	          t=$$1; sub(/:.*/, "", t); \
	          printf "  $(CYAN)%-26s$(RESET) %s\n", t, desc; \
	          desc="" \
	      }' $(MAKEFILE_LIST)
	@printf "\n$(BOLD)Variables$(RESET) (override with make VAR=value):\n"
	@printf "  $(CYAN)%-26s$(RESET) %s\n" "BUILD_DIR"   "$(BUILD_DIR)"
	@printf "  $(CYAN)%-26s$(RESET) %s\n" "PRESET_DBG"  "$(PRESET_DBG)"
	@printf "  $(CYAN)%-26s$(RESET) %s\n" "PRESET_REL"  "$(PRESET_REL)"
	@printf "  $(CYAN)%-26s$(RESET) %s\n" "CMAKE_FLAGS" "(empty)"
	@printf "\n$(BOLD)Examples:$(RESET)\n"
	@printf "  make build                   $(GRAY)# debug build$(RESET)\n"
	@printf "  make build-release           $(GRAY)# release build$(RESET)\n"
	@printf "  make test                    $(GRAY)# run test suite$(RESET)\n"
	@printf "  make lint                    $(GRAY)# clang-tidy$(RESET)\n"
	@printf "  make format                  $(GRAY)# clang-format (in-place)$(RESET)\n"
	@printf "  make format-check            $(GRAY)# CI formatting check$(RESET)\n"
	@printf "  make version-bump-patch      $(GRAY)# x.y.Z → x.y.(Z+1)$(RESET)\n"
	@printf "  make version-bump-minor      $(GRAY)# x.y.Z → x.(y+1).0$(RESET)\n"
	@printf "  make version-bump-major      $(GRAY)# x.y.Z → (x+1).0.0$(RESET)\n"
	@printf "  make changelog               $(GRAY)# update CHANGELOG.md$(RESET)\n"
	@printf "  make CMAKE_FLAGS=-DFOO=1 build $(GRAY)# extra CMake flags$(RESET)\n\n"
`
