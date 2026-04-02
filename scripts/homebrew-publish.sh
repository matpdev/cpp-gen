#!/usr/bin/env bash
# ==============================================================================
# scripts/homebrew-publish.sh — cpp-gen
# ==============================================================================
# Publishes the cpp-gen Homebrew formula to the matpdev/homebrew-tap GitHub
# repository for a given release version.
#
# Fetches real SHA256 checksums from the GitHub release, generates the formula
# from the homebrew/cpp-gen.rb template, clones the tap repo, commits and
# pushes — all locally, without relying on CI/CD.
#
# Prerequisites:
#   - The matpdev/homebrew-tap repository must exist on GitHub.
#     Create it at: https://github.com/new  (name: homebrew-tap)
#   - A GitHub Personal Access Token with write access to homebrew-tap must be
#     set in the HOMEBREW_TAP_GITHUB_TOKEN environment variable.
#
# Usage:
#   ./scripts/homebrew-publish.sh                    # uses latest git tag
#   ./scripts/homebrew-publish.sh --version 1.2.3   # explicit version
#   ./scripts/homebrew-publish.sh --dry-run          # no push, just preview
#   ./scripts/homebrew-publish.sh --yes              # no confirmation prompt
#
# Environment variables (alternative to flags):
#   HOMEBREW_TAP_GITHUB_TOKEN   GitHub PAT with write access to homebrew-tap
#                               (required — script will abort without it)
#   HOMEBREW_VERSION            version to publish (alternative to --version)
#
# Make this script executable:
#   chmod +x scripts/homebrew-publish.sh
# ==============================================================================

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────────────────

VERSION="${HOMEBREW_VERSION:-}"
DRY_RUN=false
AUTO_YES=false

GITHUB_REPO="matpdev/cpp-gen"
TAP_REPO="matpdev/homebrew-tap"
TAP_CLONE_DIR="/tmp/homebrew-tap-cpp-gen"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FORMULA_TPL="$REPO_ROOT/homebrew/cpp-gen.rb"

# ── Argument parsing ──────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)        DRY_RUN=true ;;
    --yes|-y)         AUTO_YES=true ;;
    --version|-v)     VERSION="$2"; shift ;;
    --version=*)      VERSION="${1#*=}" ;;
    --help|-h)
      grep '^#' "$0" | head -35 | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      printf "Unknown option: %s\n" "$1" >&2
      exit 1
      ;;
  esac
  shift
done

# ── Colors ────────────────────────────────────────────────────────────────────

RESET="\033[0m"
BOLD="\033[1m"
GREEN="\033[32m"
CYAN="\033[36m"
YELLOW="\033[33m"
PURPLE="\033[35m"
RED="\033[31m"
GRAY="\033[90m"
DIM="\033[2m"

# ── Helpers ───────────────────────────────────────────────────────────────────

info()    { printf "  ${BOLD}${CYAN}%-14s${RESET} %s\n" "$1" "$2"; }
success() { printf "  ${BOLD}${GREEN}✓ %-12s${RESET} %s\n" "$1" "$2"; }
warn()    { printf "  ${BOLD}${YELLOW}⚠ %-12s${RESET} %s\n" "$1" "$2"; }
error()   { printf "  ${BOLD}${RED}✗ %-12s${RESET} %s\n" "$1" "$2" >&2; }
section() { printf "\n${BOLD}${PURPLE}── %s ${RESET}\n" "$1"; }
dim()     { printf "  ${DIM}${GRAY}%s${RESET}\n" "$1"; }

die() {
  error "Error" "$1"
  exit 1
}

# ── Banner ────────────────────────────────────────────────────────────────────

printf "\n"
printf "${BOLD}${PURPLE}╔══════════════════════════════════════════════════════════════╗${RESET}\n"
printf "${BOLD}${PURPLE}║${RESET}  ${BOLD}⚡ cpp-gen — Homebrew Publish Script${RESET}                       ${BOLD}${PURPLE}║${RESET}\n"
printf "${BOLD}${PURPLE}╚══════════════════════════════════════════════════════════════╝${RESET}\n"

# ── Environment checks ────────────────────────────────────────────────────────

section "Checking prerequisites"

for cmd in git curl; do
  if command -v "$cmd" >/dev/null 2>&1; then
    success "found" "$cmd"
  else
    die "'$cmd' not found in PATH."
  fi
done

git rev-parse --git-dir >/dev/null 2>&1 || die "Not a git repository."
success "found" "git repository"

if [[ -z "${HOMEBREW_TAP_GITHUB_TOKEN:-}" ]]; then
  printf "\n"
  error "Token" "HOMEBREW_TAP_GITHUB_TOKEN is not set."
  printf "\n"
  printf "  Set it before running this script:\n"
  printf "  ${CYAN}export HOMEBREW_TAP_GITHUB_TOKEN=ghp_your_token_here${RESET}\n"
  printf "\n"
  printf "  Generate a token at:\n"
  printf "  ${CYAN}https://github.com/settings/tokens${RESET}\n"
  printf "  Required scopes: ${BOLD}repo${RESET} (or at minimum: ${BOLD}public_repo${RESET})\n"
  printf "\n"
  exit 1
fi
success "found" "HOMEBREW_TAP_GITHUB_TOKEN"

[[ -f "$FORMULA_TPL" ]] || die "Formula template not found: $FORMULA_TPL"
success "found" "homebrew/cpp-gen.rb template"

# ── Version resolution ────────────────────────────────────────────────────────

section "Version"

if [[ -z "$VERSION" ]]; then
  info "Detecting" "latest git tag..."

  LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

  if [[ -z "$LATEST_TAG" ]]; then
    die "No git tags found. Use --version <x.y.z> to specify the version explicitly."
  fi

  VERSION="${LATEST_TAG#v}"
  info "Detected" "v${VERSION} (latest git tag: ${LATEST_TAG})"
else
  VERSION="${VERSION#v}"
  info "Specified" "v${VERSION}"
fi

if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  die "Version '$VERSION' is not in MAJOR.MINOR.PATCH format."
fi

# ── Fetch checksums from GitHub release ──────────────────────────────────────

section "Checksums"

CHECKSUMS_URL="https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/checksums.txt"

info "Fetching" "$CHECKSUMS_URL"

CHECKSUMS_TMP="$(mktemp)"
trap 'rm -f "$CHECKSUMS_TMP"' EXIT

HTTP_CODE=$(curl -sL -w "%{http_code}" -o "$CHECKSUMS_TMP" "$CHECKSUMS_URL")

if [[ "$HTTP_CODE" != "200" ]]; then
  printf "\n"
  error "Download" "HTTP ${HTTP_CODE} — checksums.txt not found."
  printf "\n"
  printf "  Make sure the release v${VERSION} exists at:\n"
  printf "  ${CYAN}https://github.com/%s/releases/tag/v%s${RESET}\n" "$GITHUB_REPO" "$VERSION"
  printf "\n"
  exit 1
fi

SHA_DARWIN_AMD64=$(grep "darwin_amd64\.tar\.gz"  "$CHECKSUMS_TMP" | awk '{print $1}' || echo "")
SHA_DARWIN_ARM64=$(grep "darwin_arm64\.tar\.gz"  "$CHECKSUMS_TMP" | awk '{print $1}' || echo "")
SHA_LINUX_AMD64=$(grep  "linux_amd64\.tar\.gz"   "$CHECKSUMS_TMP" | awk '{print $1}' || echo "")
SHA_LINUX_ARM64=$(grep  "linux_arm64\.tar\.gz"   "$CHECKSUMS_TMP" | awk '{print $1}' || echo "")

[[ -n "$SHA_DARWIN_AMD64" ]] || die "Checksum for darwin_amd64 not found in checksums.txt"
[[ -n "$SHA_LINUX_AMD64"  ]] || die "Checksum for linux_amd64 not found in checksums.txt"

success "darwin_amd64" "$SHA_DARWIN_AMD64"

if [[ -n "$SHA_DARWIN_ARM64" ]]; then
  success "darwin_arm64" "$SHA_DARWIN_ARM64"
else
  warn "darwin_arm64" "Not found in checksums.txt — will keep PLACEHOLDER_SHA256_DARWIN_ARM64"
  SHA_DARWIN_ARM64="PLACEHOLDER_SHA256_DARWIN_ARM64"
fi

success "linux_amd64"  "$SHA_LINUX_AMD64"

if [[ -n "$SHA_LINUX_ARM64" ]]; then
  success "linux_arm64" "$SHA_LINUX_ARM64"
else
  warn "linux_arm64" "Not found in checksums.txt — will keep PLACEHOLDER_SHA256_LINUX_ARM64"
  SHA_LINUX_ARM64="PLACEHOLDER_SHA256_LINUX_ARM64"
fi

# ── Generate formula ──────────────────────────────────────────────────────────

section "Formula"

FORMULA_OUT="$(mktemp /tmp/cpp-gen-formula-XXXXXX.rb)"
trap 'rm -f "$CHECKSUMS_TMP" "$FORMULA_OUT"' EXIT

sed \
  -e "s/version \"0\.1\.0\"/version \"${VERSION}\"/" \
  -e "s/PLACEHOLDER_SHA256_DARWIN_AMD64/${SHA_DARWIN_AMD64}/" \
  -e "s/PLACEHOLDER_SHA256_DARWIN_ARM64/${SHA_DARWIN_ARM64}/" \
  -e "s/PLACEHOLDER_SHA256_LINUX_AMD64/${SHA_LINUX_AMD64}/"   \
  -e "s/PLACEHOLDER_SHA256_LINUX_ARM64/${SHA_LINUX_ARM64}/"   \
  "$FORMULA_TPL" > "$FORMULA_OUT"

success "Generated" "$FORMULA_OUT"

# ── Summary panel ─────────────────────────────────────────────────────────────

printf "\n"
printf "${BOLD}${PURPLE}╔══════════════════════════════════════════════════════════════╗${RESET}\n"
printf "${BOLD}${PURPLE}║${RESET}  ${BOLD}Publish Summary${RESET}                                            ${BOLD}${PURPLE}║${RESET}\n"
printf "${BOLD}${PURPLE}╚══════════════════════════════════════════════════════════════╝${RESET}\n"

info "Version"       "v${VERSION}"
info "darwin_amd64"  "$SHA_DARWIN_AMD64"
info "darwin_arm64"  "$SHA_DARWIN_ARM64"
info "linux_amd64"   "$SHA_LINUX_AMD64"
info "linux_arm64"   "$SHA_LINUX_ARM64"
info "Tap repo"      "github.com/${TAP_REPO}"

# ── Dry-run ───────────────────────────────────────────────────────────────────

if [[ "$DRY_RUN" == true ]]; then
  printf "\n"
  section "Formula preview (dry-run)"
  printf "\n"
  cat "$FORMULA_OUT"
  printf "\n"
  warn "dry-run" "No push was made."
  dim "Remove --dry-run to publish to the Homebrew tap."
  printf "\n"
  exit 0
fi

# ── Confirmation ──────────────────────────────────────────────────────────────

printf "\n"

if [[ "$AUTO_YES" == false ]]; then
  printf "  Publish ${BOLD}${GREEN}v${VERSION}${RESET} to ${CYAN}${TAP_REPO}${RESET}? [y/N] "
  read -r CONFIRM
  if [[ "${CONFIRM,,}" != "y" ]]; then
    printf "\n"
    dim "Aborted. Nothing was pushed."
    printf "\n"
    exit 0
  fi
fi

# ── Clone tap repo ────────────────────────────────────────────────────────────

section "Cloning tap repository"

# Clean up any leftover clone from a previous run
rm -rf "$TAP_CLONE_DIR"

info "Cloning" "github.com/${TAP_REPO}..."

if ! git clone \
  "https://x-access-token:${HOMEBREW_TAP_GITHUB_TOKEN}@github.com/${TAP_REPO}.git" \
  "$TAP_CLONE_DIR" 2>&1 | while IFS= read -r line; do dim "$line"; done; then
  printf "\n"
  error "Clone" "Failed to clone github.com/${TAP_REPO}."
  printf "\n"
  printf "  Make sure the repository exists at:\n"
  printf "  ${CYAN}https://github.com/${TAP_REPO}${RESET}\n"
  printf "\n"
  printf "  If it doesn't exist yet, create it at:\n"
  printf "  ${CYAN}https://github.com/new${RESET}  (repository name: ${BOLD}homebrew-tap${RESET})\n"
  printf "\n"
  printf "  Also verify that HOMEBREW_TAP_GITHUB_TOKEN has write access to that repo.\n"
  printf "\n"
  exit 1
fi

success "Cloned" "$TAP_CLONE_DIR"

# ── Copy formula into tap ─────────────────────────────────────────────────────

section "Updating formula"

mkdir -p "$TAP_CLONE_DIR/Formula"
cp "$FORMULA_OUT" "$TAP_CLONE_DIR/Formula/cpp-gen.rb"

success "Copied" "Formula/cpp-gen.rb"

# ── Commit and push ───────────────────────────────────────────────────────────

section "Commit & push"

cd "$TAP_CLONE_DIR"

git config user.name  "$(git -C "$REPO_ROOT" config user.name  2>/dev/null || echo 'matpdev')"
git config user.email "$(git -C "$REPO_ROOT" config user.email 2>/dev/null || echo 'matheus2ep@gmail.com')"

git add Formula/cpp-gen.rb

if git diff --staged --quiet; then
  warn "Commit" "No changes detected — the tap is already at v${VERSION}."
  printf "\n"
  dim "Nothing to push. The formula is already up to date."
  printf "\n"
  exit 0
fi

git diff --staged --stat | while IFS= read -r line; do
  dim "$line"
done

COMMIT_MSG="Brew formula update for cpp-gen version v${VERSION}"
git commit -m "$COMMIT_MSG"
success "Commit" "$COMMIT_MSG"

info "Pushing" "origin → github.com/${TAP_REPO}..."

git push origin HEAD 2>&1 | while IFS= read -r line; do
  dim "$line"
done

# ── Done ──────────────────────────────────────────────────────────────────────

printf "\n"
printf "${BOLD}${PURPLE}╔══════════════════════════════════════════════════════════════╗${RESET}\n"
printf "${BOLD}${PURPLE}║${RESET}  ${BOLD}${GREEN}✓ cpp-gen v%s published to Homebrew tap!${RESET}\n" "$VERSION"
printf "${BOLD}${PURPLE}╚══════════════════════════════════════════════════════════════╝${RESET}\n"
printf "\n"
printf "  ${GRAY}Tap repo:${RESET}  ${CYAN}https://github.com/${TAP_REPO}${RESET}\n"
printf "\n"
printf "  ${DIM}Install with:${RESET}\n"
printf "  ${CYAN}  brew tap matpdev/tap${RESET}\n"
printf "  ${CYAN}  brew install cpp-gen${RESET}\n"
printf "\n"
