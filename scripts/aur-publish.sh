#!/usr/bin/env bash
# ==============================================================================
# scripts/aur-publish.sh — cpp-gen
# ==============================================================================
# Publishes the cpp-gen-bin package to the AUR for a given release version.
#
# Fetches real SHA256 checksums from the GitHub release, generates PKGBUILD
# and .SRCINFO, clones the AUR git repo, commits and pushes — all locally,
# without relying on CI/CD.
#
# Usage:
#   ./scripts/aur-publish.sh                        # uses latest git tag
#   ./scripts/aur-publish.sh --version 1.2.3        # explicit version
#   ./scripts/aur-publish.sh --key ~/.ssh/aur        # explicit SSH key
#   ./scripts/aur-publish.sh --dry-run              # no push, just preview
#   ./scripts/aur-publish.sh --yes                  # no confirmation prompt
#
# Environment variables (alternative to flags):
#   AUR_SSH_KEY   path to AUR SSH private key (default: ~/.ssh/aur)
#   AUR_VERSION   version to publish          (default: latest git tag)
#
# Prerequisites:
#   - git, curl, ssh, ssh-keygen, ssh-agent, ssh-add
#   - AUR account with SSH public key registered
#   - AUR SSH private key with NO passphrase
# ==============================================================================

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────────────────

VERSION="${AUR_VERSION:-}"
SSH_KEY="${AUR_SSH_KEY:-$HOME/.ssh/aur}"
DRY_RUN=false
AUTO_YES=false

AUR_REPO="ssh://aur@aur.archlinux.org/cpp-gen-bin.git"
GITHUB_REPO="matpdev/cpp-gen"
PKGNAME="cpp-gen-bin"
PKGDESC="Modern C++ project generator with CMake, package managers, IDE configurations and development tools"

WORK_DIR="$(mktemp -d)"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# ── Argument parsing ──────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)        DRY_RUN=true ;;
    --yes|-y)         AUTO_YES=true ;;
    --version|-v)     VERSION="$2"; shift ;;
    --version=*)      VERSION="${1#*=}" ;;
    --key|-k)         SSH_KEY="$2"; shift ;;
    --key=*)          SSH_KEY="${1#*=}" ;;
    --help|-h)
      grep '^#' "$0" | head -30 | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      printf "Opção desconhecida: %s\n" "$1" >&2
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
  error "Erro" "$1"
  rm -rf "$WORK_DIR"
  exit 1
}

cleanup() {
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

# ── Banner ────────────────────────────────────────────────────────────────────

printf "\n"
printf "${BOLD}${PURPLE}╔══════════════════════════════════════════════════════════════╗${RESET}\n"
printf "${BOLD}${PURPLE}║${RESET}  ${BOLD}⚡ cpp-gen — AUR Publish Script${RESET}                            ${BOLD}${PURPLE}║${RESET}\n"
printf "${BOLD}${PURPLE}╚══════════════════════════════════════════════════════════════╝${RESET}\n"

# ── Prerequisite checks ───────────────────────────────────────────────────────

section "Verificando pré-requisitos"

for cmd in git curl ssh ssh-keygen ssh-agent ssh-add; do
  if command -v "$cmd" >/dev/null 2>&1; then
    success "found" "$cmd"
  else
    die "'$cmd' não encontrado no PATH."
  fi
done

# ── Version resolution ────────────────────────────────────────────────────────

section "Versão"

if [[ -z "$VERSION" ]]; then
  cd "$REPO_ROOT"
  VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || true)
  if [[ -z "$VERSION" ]]; then
    die "Nenhuma tag encontrada. Use --version <x.y.z> para especificar a versão."
  fi
  info "Detectada" "v${VERSION} (última tag git)"
else
  VERSION="${VERSION#v}"
  info "Especificada" "v${VERSION}"
fi

if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  die "Versão '$VERSION' não está no formato MAJOR.MINOR.PATCH."
fi

# ── SSH key validation ────────────────────────────────────────────────────────

section "Chave SSH"

info "Chave" "$SSH_KEY"

[[ -f "$SSH_KEY" ]] || die "Arquivo de chave SSH não encontrado: $SSH_KEY"

chmod 600 "$SSH_KEY"

# Verify the key can produce a public key (requires valid private key, no passphrase)
if ! ssh-keygen -y -f "$SSH_KEY" >/dev/null 2>&1; then
  printf "\n"
  error "Chave" "Não foi possível ler '$SSH_KEY'."
  printf "\n"
  printf "  Possíveis causas:\n"
  printf "  ${YELLOW}1.${RESET} A chave tem passphrase — remova com:\n"
  printf "     ${CYAN}ssh-keygen -p -N \"\" -f %s${RESET}\n" "$SSH_KEY"
  printf "  ${YELLOW}2.${RESET} O arquivo está corrompido — gere uma nova chave:\n"
  printf "     ${CYAN}ssh-keygen -t ed25519 -f %s${RESET}\n" "$SSH_KEY"
  printf "  ${YELLOW}3.${RESET} Não se esqueça de cadastrar a chave pública no AUR:\n"
  printf "     ${CYAN}https://aur.archlinux.org/account/${RESET}\n"
  printf "\n"
  exit 1
fi

success "Válida" "$(ssh-keygen -l -f "$SSH_KEY" | awk '{print $2, $4}')"

# ── Load key into SSH agent ───────────────────────────────────────────────────

section "SSH Agent"

# Start a fresh agent scoped to this script
eval "$(ssh-agent -s)" >/dev/null
SSH_AGENT_STARTED=true

# Ensure agent is killed on exit
trap 'ssh-agent -k >/dev/null 2>&1 || true; cleanup' EXIT

ssh-add "$SSH_KEY" 2>&1 | while IFS= read -r line; do
  dim "$line"
done

if ! ssh-add -l >/dev/null 2>&1; then
  die "ssh-add falhou. Verifique se a chave não tem passphrase."
fi

success "Carregada" "chave adicionada ao agente (PID $SSH_AGENT_PID)"

# Test SSH connectivity to AUR
info "Testando" "conexão com aur.archlinux.org..."
ssh-keyscan -H aur.archlinux.org >> ~/.ssh/known_hosts 2>/dev/null

if ssh -o BatchMode=yes -o ConnectTimeout=10 aur@aur.archlinux.org help 2>&1 | grep -q "Interactive"; then
  success "Conexão" "aur.archlinux.org acessível"
else
  # ssh to AUR always exits non-zero, but we can check the output
  SSH_TEST=$(ssh -o BatchMode=yes -o ConnectTimeout=10 aur@aur.archlinux.org help 2>&1 || true)
  if echo "$SSH_TEST" | grep -qi "permission denied\|publickey"; then
    printf "\n"
    error "SSH" "Autenticação negada pelo AUR."
    printf "\n"
    printf "  Verifique se a chave pública está cadastrada em:\n"
    printf "  ${CYAN}https://aur.archlinux.org/account/${RESET}\n"
    printf "\n"
    printf "  Chave pública desta sessão:\n"
    ssh-keygen -y -f "$SSH_KEY" | while IFS= read -r line; do
      printf "  ${GRAY}%s${RESET}\n" "$line"
    done
    printf "\n"
    exit 1
  fi
  success "Conexão" "aur.archlinux.org acessível"
fi

# ── Fetch checksums from GitHub release ──────────────────────────────────────

section "Checksums"

CHECKSUMS_URL="https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/checksums.txt"

info "Buscando" "$CHECKSUMS_URL"

CHECKSUMS_FILE="$WORK_DIR/checksums.txt"
HTTP_CODE=$(curl -sL -w "%{http_code}" -o "$CHECKSUMS_FILE" "$CHECKSUMS_URL")

if [[ "$HTTP_CODE" != "200" ]]; then
  printf "\n"
  error "Download" "HTTP $HTTP_CODE — checksums.txt não encontrado."
  printf "\n"
  printf "  Verifique se a release v${VERSION} existe em:\n"
  printf "  ${CYAN}https://github.com/%s/releases/tag/v%s${RESET}\n" "$GITHUB_REPO" "$VERSION"
  printf "\n"
  exit 1
fi

SHA_X86_64=$(grep "linux_amd64\.tar\.gz" "$CHECKSUMS_FILE" | awk '{print $1}')
SHA_I686=$(grep "linux_386\.tar\.gz"     "$CHECKSUMS_FILE" | awk '{print $1}')

[[ -n "$SHA_X86_64" ]] || die "Checksum para linux_amd64 não encontrado em checksums.txt"
[[ -n "$SHA_I686"   ]] || die "Checksum para linux_386 não encontrado em checksums.txt"

success "x86_64" "$SHA_X86_64"
success "i686"   "$SHA_I686"

# ── Generate PKGBUILD ─────────────────────────────────────────────────────────

section "PKGBUILD"

PKGBUILD_TPL="$REPO_ROOT/aur/PKGBUILD"
[[ -f "$PKGBUILD_TPL" ]] || die "Template não encontrado: $PKGBUILD_TPL"

PKGBUILD_OUT="$WORK_DIR/PKGBUILD"

sed \
  -e "s/^pkgver=.*/pkgver=${VERSION}/" \
  -e "s/^pkgrel=.*/pkgrel=1/" \
  -e "s/sha256sums_x86_64=('SKIP')/sha256sums_x86_64=('${SHA_X86_64}')/" \
  -e "s/sha256sums_i686=('SKIP')/sha256sums_i686=('${SHA_I686}')/" \
  "$PKGBUILD_TPL" > "$PKGBUILD_OUT"

success "Gerado" "$PKGBUILD_OUT"

# ── Generate .SRCINFO ─────────────────────────────────────────────────────────

section ".SRCINFO"

SRCINFO_OUT="$WORK_DIR/.SRCINFO"

if command -v makepkg >/dev/null 2>&1; then
  # Preferred: use makepkg for 100% accurate .SRCINFO
  cp "$PKGBUILD_OUT" "$WORK_DIR/PKGBUILD"
  (cd "$WORK_DIR" && makepkg --printsrcinfo > "$SRCINFO_OUT")
  success "Gerado" ".SRCINFO via makepkg --printsrcinfo"
else
  # Fallback: generate manually (safe for this known package structure)
  {
    printf 'pkgbase = %s\n'                          "$PKGNAME"
    printf '\tpkgdesc = %s\n'                        "$PKGDESC"
    printf '\tpkgver = %s\n'                         "$VERSION"
    printf '\tpkgrel = 1\n'
    printf '\turl = https://github.com/%s\n'         "$GITHUB_REPO"
    printf '\tarch = x86_64\n'
    printf '\tarch = i686\n'
    printf '\tlicense = MIT\n'
    printf '\tprovides = cpp-gen\n'
    printf '\tconflicts = cpp-gen\n'
    printf '\toptions = !strip\n'
    printf '\tsource_x86_64 = %s-%s-x86_64.tar.gz::https://github.com/%s/releases/download/v%s/cpp-gen_%s_linux_amd64.tar.gz\n' \
      "$PKGNAME" "$VERSION" "$GITHUB_REPO" "$VERSION" "$VERSION"
    printf '\tsha256sums_x86_64 = %s\n'              "$SHA_X86_64"
    printf '\tsource_i686 = %s-%s-i686.tar.gz::https://github.com/%s/releases/download/v%s/cpp-gen_%s_linux_386.tar.gz\n' \
      "$PKGNAME" "$VERSION" "$GITHUB_REPO" "$VERSION" "$VERSION"
    printf '\tsha256sums_i686 = %s\n'                "$SHA_I686"
    printf '\n'
    printf 'pkgname = %s\n'                          "$PKGNAME"
  } > "$SRCINFO_OUT"
  warn "Gerado" ".SRCINFO manual (instale makepkg para maior precisão)"
fi

# ── Copy LICENSE ──────────────────────────────────────────────────────────────

LICENSE_SRC="$REPO_ROOT/aur/LICENSE"
[[ -f "$LICENSE_SRC" ]] || die "aur/LICENSE não encontrado."
cp "$LICENSE_SRC" "$WORK_DIR/LICENSE"

# ── Preview ───────────────────────────────────────────────────────────────────

section "Arquivos que serão enviados ao AUR"

for f in PKGBUILD .SRCINFO LICENSE; do
  printf "\n  ${BOLD}${CYAN}%s${RESET}\n" "$f"
  sed 's/^/    /' "$WORK_DIR/$f"
done

# ── Dry-run ───────────────────────────────────────────────────────────────────

if [[ "$DRY_RUN" == true ]]; then
  printf "\n"
  warn "dry-run" "Nenhum push foi feito."
  dim "Remova --dry-run para publicar no AUR."
  printf "\n"
  exit 0
fi

# ── Confirmation ──────────────────────────────────────────────────────────────

printf "\n"

if [[ "$AUTO_YES" == false ]]; then
  printf "  Publicar ${BOLD}${GREEN}v${VERSION}${RESET} no AUR (${CYAN}${PKGNAME}${RESET})? [s/N] "
  read -r CONFIRM
  if [[ "${CONFIRM,,}" != "s" ]]; then
    printf "\n"
    dim "Abortado. Nenhum push foi feito."
    printf "\n"
    exit 0
  fi
fi

# ── Clone AUR repo ────────────────────────────────────────────────────────────

section "Clonando repositório AUR"

AUR_CLONE="$WORK_DIR/aur-repo"

info "Clonando" "$AUR_REPO"

git -c init.defaultBranch=master clone "$AUR_REPO" "$AUR_CLONE" 2>&1 | while IFS= read -r line; do
  dim "$line"
done

# AUR always starts on master — rename if needed
cd "$AUR_CLONE"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "master")
if [[ "$CURRENT_BRANCH" != "master" ]]; then
  git branch -m "$CURRENT_BRANCH" master
fi

success "Clonado" "$AUR_CLONE"

# ── Copy files into the AUR repo ──────────────────────────────────────────────

section "Atualizando arquivos"

cp "$WORK_DIR/PKGBUILD"  "$AUR_CLONE/PKGBUILD"
cp "$WORK_DIR/.SRCINFO"  "$AUR_CLONE/.SRCINFO"
cp "$WORK_DIR/LICENSE"   "$AUR_CLONE/LICENSE"

success "Copiado" "PKGBUILD, .SRCINFO, LICENSE"

# ── Commit ────────────────────────────────────────────────────────────────────

section "Commit"

cd "$AUR_CLONE"

git config user.name  "$(git -C "$REPO_ROOT" config user.name  2>/dev/null || echo 'matpdev')"
git config user.email "$(git -C "$REPO_ROOT" config user.email 2>/dev/null || echo 'matheus2ep@gmail.com')"

git add PKGBUILD .SRCINFO LICENSE

if git diff --staged --quiet; then
  warn "Commit" "Nenhuma alteração detectada — o AUR já está na v${VERSION}."
  exit 0
fi

git diff --staged --stat | while IFS= read -r line; do
  dim "$line"
done

COMMIT_MSG="Update to v${VERSION}"
git commit -m "$COMMIT_MSG"
success "Commit" "$COMMIT_MSG"

# ── Push ──────────────────────────────────────────────────────────────────────

section "Push para o AUR"

info "Enviando" "origin master → $AUR_REPO"

git push origin master 2>&1 | while IFS= read -r line; do
  dim "$line"
done

# ── Done ──────────────────────────────────────────────────────────────────────

printf "\n"
printf "${BOLD}${PURPLE}╔══════════════════════════════════════════════════════════════╗${RESET}\n"
printf "${BOLD}${PURPLE}║${RESET}  ${BOLD}${GREEN}✓ cpp-gen-bin v%s publicado no AUR!${RESET}\n" "$VERSION"
printf "${BOLD}${PURPLE}╚══════════════════════════════════════════════════════════════╝${RESET}\n"
printf "\n"
printf "  ${GRAY}Confira em:${RESET} ${CYAN}https://aur.archlinux.org/packages/cpp-gen-bin${RESET}\n"
printf "\n"
printf "  ${DIM}Instalar com:${RESET}  ${CYAN}yay -S cpp-gen-bin${RESET}\n"
printf "\n"
