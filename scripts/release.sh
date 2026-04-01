#!/usr/bin/env bash
# ==============================================================================
# scripts/release.sh — cpp-gen
# ==============================================================================
# Calculates the next semantic version based on conventional commits since
# the last git tag, displays a summary and creates + pushes the release tag.
#
# Conventional Commits → SemVer:
#   feat!: / BREAKING CHANGE  →  MAJOR bump  (x.0.0)
#   feat:                     →  MINOR bump  (0.x.0)
#   fix: / perf: / refactor:  →  PATCH bump  (0.0.x)
#
# Usage:
#   ./scripts/release.sh              # interactive mode
#   ./scripts/release.sh --dry-run    # only displays the calculated version
#   ./scripts/release.sh --yes        # no confirmation (CI)
# ==============================================================================

set -euo pipefail

# ── Options ───────────────────────────────────────────────────────────────────

DRY_RUN=false
AUTO_YES=false

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --yes|-y)  AUTO_YES=true ;;
    --help|-h)
      grep '^#' "$0" | head -20 | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "Opção desconhecida: $arg" >&2
      exit 1
      ;;
  esac
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

info()    { printf "  ${BOLD}${CYAN}%-12s${RESET} %s\n" "$1" "$2"; }
success() { printf "  ${BOLD}${GREEN}✓ %-10s${RESET} %s\n" "$1" "$2"; }
warn()    { printf "  ${BOLD}${YELLOW}⚠ %-10s${RESET} %s\n" "$1" "$2"; }
error()   { printf "  ${BOLD}${RED}✗ %-10s${RESET} %s\n" "$1" "$2" >&2; }
section() { printf "\n${BOLD}${PURPLE}%s${RESET}\n" "$1"; }
dim()     { printf "  ${DIM}${GRAY}%s${RESET}\n" "$1"; }

die() {
  error "Erro" "$1"
  exit 1
}

# ── Environment checks ────────────────────────────────────────────────────────

command -v git >/dev/null 2>&1 || die "git não encontrado no PATH."

git rev-parse --git-dir >/dev/null 2>&1 || die "Não é um repositório git."

# Check if there are commits
git rev-parse HEAD >/dev/null 2>&1 || die "Nenhum commit encontrado no repositório."

# ── Last tag ──────────────────────────────────────────────────────────────────

LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -z "$LAST_TAG" ]; then
  warn "Aviso" "Nenhuma tag encontrada — iniciando a partir de v0.0.0"
  LAST_TAG="v0.0.0"
  COMMITS=$(git log --pretty=format:"%s" 2>/dev/null)
else
  COMMITS=$(git log "${LAST_TAG}..HEAD" --pretty=format:"%s" 2>/dev/null)
fi

# ── Current version ───────────────────────────────────────────────────────────

CURRENT_VERSION="${LAST_TAG#v}"

# Extracts major.minor.patch (tolerates tags without the 'v' prefix)
if ! echo "$CURRENT_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+'; then
  die "Tag '$LAST_TAG' não está no formato vMAJOR.MINOR.PATCH."
fi

MAJOR=$(echo "$CURRENT_VERSION" | cut -d. -f1)
MINOR=$(echo "$CURRENT_VERSION" | cut -d. -f2)
PATCH=$(echo "$CURRENT_VERSION" | cut -d. -f3)

# ── Commit analysis ───────────────────────────────────────────────────────────

BUMP="patch"
COUNT_BREAKING=0
COUNT_FEAT=0
COUNT_FIX=0
COUNT_PERF=0
COUNT_REFACTOR=0
COUNT_OTHER=0

while IFS= read -r commit; do
  [ -z "$commit" ] && continue

  # Breaking change: type with ! or BREAKING CHANGE in body
  if echo "$commit" | grep -qE '^[a-z]+(\(.+\))?!:' || \
     echo "$commit" | grep -qiE '^BREAKING[[:space:]]CHANGE'; then
    BUMP="major"
    COUNT_BREAKING=$((COUNT_BREAKING + 1))
    continue
  fi

  # Feature: new feature → at minimum minor
  if echo "$commit" | grep -qE '^feat(\(.+\))?:'; then
    [ "$BUMP" = "patch" ] && BUMP="minor"
    COUNT_FEAT=$((COUNT_FEAT + 1))
    continue
  fi

  # Fix
  if echo "$commit" | grep -qE '^fix(\(.+\))?:'; then
    COUNT_FIX=$((COUNT_FIX + 1))
    continue
  fi

  # Performance
  if echo "$commit" | grep -qE '^perf(\(.+\))?:'; then
    COUNT_PERF=$((COUNT_PERF + 1))
    continue
  fi

  # Refactor
  if echo "$commit" | grep -qE '^refactor(\(.+\))?:'; then
    COUNT_REFACTOR=$((COUNT_REFACTOR + 1))
    continue
  fi

  COUNT_OTHER=$((COUNT_OTHER + 1))
done <<EOF
$COMMITS
EOF

# ── Calculate new version ─────────────────────────────────────────────────────

case "$BUMP" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

# ── Total commit count ────────────────────────────────────────────────────────

TOTAL_COMMITS=0
if [ -n "$COMMITS" ]; then
  TOTAL_COMMITS=$(echo "$COMMITS" | grep -c . || true)
fi

# ── Display report ────────────────────────────────────────────────────────────

printf "\n"
printf "${BOLD}${PURPLE}╔══════════════════════════════════════════════════════════════╗${RESET}\n"
printf "${BOLD}${PURPLE}║${RESET}  ${BOLD}⚡ cpp-gen — Release Script${RESET}                               ${BOLD}${PURPLE}║${RESET}\n"
printf "${BOLD}${PURPLE}╚══════════════════════════════════════════════════════════════╝${RESET}\n"

section "Versão"
info "Última tag"    "$LAST_TAG"
info "Bump tipo"     "$BUMP"
info "Nova versão"   "${BOLD}${GREEN}${NEW_VERSION}${RESET}"

section "Commits desde $LAST_TAG  ${DIM}(${TOTAL_COMMITS} total)${RESET}"

if [ "$TOTAL_COMMITS" -eq 0 ]; then
  warn "Aviso" "Nenhum commit novo desde $LAST_TAG"
  if [ "$DRY_RUN" = false ] && [ "$AUTO_YES" = false ]; then
    printf "\n  Continuar mesmo assim? [s/N] "
    read -r CONFIRM_EMPTY
    [[ "${CONFIRM_EMPTY,,}" != "s" ]] && { echo; dim "Abortado."; echo; exit 0; }
  fi
else
  # Breaking commits
  if [ "$COUNT_BREAKING" -gt 0 ]; then
    printf "  ${BOLD}${RED}Breaking Changes  (%d)${RESET}\n" "$COUNT_BREAKING"
    while IFS= read -r c; do
      echo "$c" | grep -qE '^[a-z]+(\(.+\))?!:|^BREAKING' && \
        printf "  ${RED}  ✗ %s${RESET}\n" "$c"
    done <<< "$COMMITS"
  fi

  # Feature commits
  if [ "$COUNT_FEAT" -gt 0 ]; then
    printf "  ${BOLD}${CYAN}Novas Funcionalidades  (%d)${RESET}\n" "$COUNT_FEAT"
    while IFS= read -r c; do
      echo "$c" | grep -qE '^feat(\(.+\))?:' && \
        printf "  ${CYAN}  ◆ %s${RESET}\n" "$c"
    done <<< "$COMMITS"
  fi

  # Fix commits
  if [ "$COUNT_FIX" -gt 0 ]; then
    printf "  ${BOLD}${GREEN}Correções  (%d)${RESET}\n" "$COUNT_FIX"
    while IFS= read -r c; do
      echo "$c" | grep -qE '^fix(\(.+\))?:' && \
        printf "  ${GREEN}  ✓ %s${RESET}\n" "$c"
    done <<< "$COMMITS"
  fi

  # Performance commits
  if [ "$COUNT_PERF" -gt 0 ]; then
    printf "  ${BOLD}${YELLOW}Performance  (%d)${RESET}\n" "$COUNT_PERF"
    while IFS= read -r c; do
      echo "$c" | grep -qE '^perf(\(.+\))?:' && \
        printf "  ${YELLOW}  ⚡ %s${RESET}\n" "$c"
    done <<< "$COMMITS"
  fi

  # Refactor commits
  if [ "$COUNT_REFACTOR" -gt 0 ]; then
    printf "  ${BOLD}${GRAY}Refatorações  (%d)${RESET}\n" "$COUNT_REFACTOR"
    while IFS= read -r c; do
      echo "$c" | grep -qE '^refactor(\(.+\))?:' && \
        printf "  ${GRAY}  ~ %s${RESET}\n" "$c"
    done <<< "$COMMITS"
  fi

  # Others
  if [ "$COUNT_OTHER" -gt 0 ]; then
    printf "  ${DIM}${GRAY}Outros  (%d)${RESET}\n" "$COUNT_OTHER"
  fi
fi

# ── Dry-run: only displays and exits ─────────────────────────────────────────

if [ "$DRY_RUN" = true ]; then
  printf "\n"
  warn "dry-run" "Nenhuma tag foi criada."
  printf "  ${DIM}Próxima versão seria: ${BOLD}${NEW_VERSION}${RESET}\n\n"
  exit 0
fi

# ── Confirmation ──────────────────────────────────────────────────────────────

printf "\n"

if [ "$AUTO_YES" = false ]; then
  printf "  Criar e enviar a tag ${BOLD}${GREEN}${NEW_VERSION}${RESET}? [s/N] "
  read -r CONFIRM
  if [[ "${CONFIRM,,}" != "s" ]]; then
    printf "\n"
    dim "Abortado. Nenhuma tag foi criada."
    printf "\n"
    exit 0
  fi
fi

# ── Create the annotated tag ─────────────────────────────────────────────────

printf "\n"
info "Criando" "tag anotada ${NEW_VERSION}..."

# Tag message lists the commit types found
TAG_MSG="Release ${NEW_VERSION}"
if [ "$TOTAL_COMMITS" -gt 0 ]; then
  TAG_MSG="${TAG_MSG}

Commits desde ${LAST_TAG}:"
  [ "$COUNT_BREAKING" -gt 0 ] && TAG_MSG="${TAG_MSG}
- ${COUNT_BREAKING} breaking change(s)"
  [ "$COUNT_FEAT"     -gt 0 ] && TAG_MSG="${TAG_MSG}
- ${COUNT_FEAT} nova(s) funcionalidade(s)"
  [ "$COUNT_FIX"      -gt 0 ] && TAG_MSG="${TAG_MSG}
- ${COUNT_FIX} correção(ões)"
  [ "$COUNT_PERF"     -gt 0 ] && TAG_MSG="${TAG_MSG}
- ${COUNT_PERF} melhoria(s) de performance"
  [ "$COUNT_REFACTOR" -gt 0 ] && TAG_MSG="${TAG_MSG}
- ${COUNT_REFACTOR} refatoração(ões)"
fi

git tag -a "$NEW_VERSION" -m "$TAG_MSG"
success "Criada" "tag local $NEW_VERSION"

# ── Push to remote ────────────────────────────────────────────────────────────

REMOTE=$(git remote 2>/dev/null | head -1)

if [ -z "$REMOTE" ]; then
  warn "Aviso" "Nenhum remote configurado — tag criada apenas localmente."
  warn "Push"  "Execute: git push origin ${NEW_VERSION}"
else
  info "Enviando" "tag ${NEW_VERSION} → ${REMOTE}..."
  git push "$REMOTE" "$NEW_VERSION"
  success "Enviada" "${NEW_VERSION} → ${REMOTE}"
fi

# ── Conclusion ────────────────────────────────────────────────────────────────

printf "\n"
printf "${BOLD}${PURPLE}╔══════════════════════════════════════════════════════════════╗${RESET}\n"
printf "${BOLD}${PURPLE}║${RESET}  ${BOLD}${GREEN}✓ Tag ${NEW_VERSION} criada com sucesso!${RESET}                        ${BOLD}${PURPLE}║${RESET}\n"
printf "${BOLD}${PURPLE}╚══════════════════════════════════════════════════════════════╝${RESET}\n"
printf "\n"
printf "  ${DIM}O GitHub Actions irá detectar a tag e disparar o goreleaser${RESET}\n"
printf "  ${DIM}automaticamente para criar a release com os binários.${RESET}\n"
printf "\n"
printf "  ${GRAY}Acompanhe em:${RESET} ${CYAN}https://github.com/SEU_USER/cpp-gen/actions${RESET}\n"
printf "\n"
