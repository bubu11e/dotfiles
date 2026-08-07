#!/usr/bin/env bash
# Re-vendor the community Claude skills from upstream at pinned commits.
#
# Each pinned ref below is bumped by Renovate (git-refs digest datasource).
# A Renovate postUpgradeTask then re-runs this script so the same PR carries
# the refreshed skill content, not just a new SHA string.
#
# The file set for each skill is derived from the upstream tree at the pinned
# ref, not hardcoded here: only the skill names are curated. A file added or
# removed upstream is mirrored on the next run without editing this script.
#
# Adding a skill: append a "repo|subpath|skill" row to SKILLS. If it comes from
# a repo not yet listed, also add a renovate-commented <NAME>_REF for that repo
# and a case arm in ref_for_repo.
#
# Requires: curl, jq.
# Run manually:  bash scripts/update-vendored-skills.sh
set -euo pipefail

# renovate: datasource=git-refs depName=mattpocock/skills packageName=https://github.com/mattpocock/skills branch=main
MATTPOCOCK_SKILLS_REF="84fdeffd12f2ee307994d1eb6feb48173b6e0502"

# renovate: datasource=git-refs depName=cloudflare/security-audit-skill packageName=https://github.com/cloudflare/security-audit-skill branch=main
CLOUDFLARE_SECURITY_AUDIT_REF="8bac42001ddd90a4dcd8d5a5045199283a8eba75"

# renovate: datasource=git-refs depName=JuliusBrussee/caveman packageName=https://github.com/JuliusBrussee/caveman branch=main
CAVEMAN_REF="ec83e5bace4c20484d704dea21e12fc4eb94e9aa"

# Managed files live under home/ (see .chezmoiroot), so the skills tree is
# home/dot_claude/skills, not dot_claude/skills at the repo root.
SKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/home/dot_claude/skills"
RETRIEVED="$(date -u +%Y-%m-%d)"

# The curated set, as "repo|subpath|skill". The subpath is where the skill dirs
# live upstream; it differs per repo and even per skill within a repo. Each
# skill's own file list is read from upstream, not stated here.
SKILLS=(
  "mattpocock/skills|skills/engineering|grill-with-docs"
  "mattpocock/skills|skills/engineering|improve-codebase-architecture"
  "mattpocock/skills|skills/productivity|grill-me"
  "cloudflare/security-audit-skill|skills|security-audit"
  "JuliusBrussee/caveman|skills|caveman-compress"
)

# Per-repo responses are cached here so a repo used by several skills is
# fetched once. Keyed by the repo slug with "/" flattened to "_".
CACHE_DIR="$(mktemp -d)"
trap 'rm -rf "$CACHE_DIR"' EXIT

ref_for_repo() {
  # $1 = repo slug -> echoes the pinned ref Renovate maintains for it.
  case "$1" in
    mattpocock/skills) echo "$MATTPOCOCK_SKILLS_REF" ;;
    cloudflare/security-audit-skill) echo "$CLOUDFLARE_SECURITY_AUDIT_REF" ;;
    JuliusBrussee/caveman) echo "$CAVEMAN_REF" ;;
    *) echo "no pinned ref for repo: $1" >&2; return 1 ;;
  esac
}

cache_key() {
  # $1 = repo slug -> echoes a filename-safe key.
  echo "${1//\//_}"
}

tree_json() {
  # $1 = repo slug, $2 = ref -> echoes the path to the cached tree JSON.
  local cache="${CACHE_DIR}/$(cache_key "$1").tree.json"
  if [[ ! -f "$cache" ]]; then
    curl -fsSL "https://api.github.com/repos/$1/git/trees/$2?recursive=1" -o "$cache"
  fi
  echo "$cache"
}

license_text() {
  # $1 = repo slug, $2 = ref -> echoes the path to the cached LICENSE.
  local cache="${CACHE_DIR}/$(cache_key "$1").LICENSE"
  if [[ ! -f "$cache" ]]; then
    curl -fsSL "https://raw.githubusercontent.com/$1/$2/LICENSE" -o "$cache"
  fi
  echo "$cache"
}

write_upstream_md() {
  # $1 = repo, $2 = subpath, $3 = skill, $4 = ref, $5 = destination dir
  # The licence is embedded verbatim from upstream so the copyright line always
  # matches the repo the files actually came from.
  {
    cat <<EOF
# Upstream provenance

Vendored from [$1](https://github.com/$1) — \`$2/$3\`.

- Pinned commit: \`$4\`
- Retrieved: ${RETRIEVED}
- Refreshed by: \`scripts/update-vendored-skills.sh\` (bumped by Renovate)

## License

Verbatim copy of \`LICENSE\` from the upstream repository at the pinned commit:

\`\`\`
EOF
    cat "$(license_text "$1" "$4")"
    echo '```'
  } > "$5/UPSTREAM.md"
}

list_skill_files() {
  # $1 = repo, $2 = subpath, $3 = skill, $4 = ref
  # -> echoes each file path relative to the skill's own dir.
  jq -r --arg p "$2/$3/" \
    '.tree[] | select(.type == "blob") | select(.path | startswith($p)) | .path[($p | length):]' \
    "$(tree_json "$1" "$4")"
}

vendor_skill() {
  # $1 = repo, $2 = subpath, $3 = skill
  local repo="$1" subpath="$2" skill="$3" ref dest f
  ref="$(ref_for_repo "$repo")"
  dest="${SKILLS_DIR}/${skill}"
  # Wipe and refetch so a file removed upstream does not linger. UPSTREAM.md is
  # regenerated below, so nothing hand-written is lost in the wipe.
  rm -rf "$dest"
  mkdir -p "$dest"
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    echo "fetching ${repo}/${subpath}/${skill}/${f}"
    mkdir -p "${dest}/$(dirname "$f")"
    curl -fsSL "https://raw.githubusercontent.com/${repo}/${ref}/${subpath}/${skill}/${f}" \
      -o "${dest}/${f}"
  done < <(list_skill_files "$repo" "$subpath" "$skill" "$ref")
  write_upstream_md "$repo" "$subpath" "$skill" "$ref" "$dest"
}

for entry in "${SKILLS[@]}"; do
  IFS='|' read -r repo subpath skill <<<"$entry"
  vendor_skill "$repo" "$subpath" "$skill"
done

echo "done: vendored ${#SKILLS[@]} skills"
