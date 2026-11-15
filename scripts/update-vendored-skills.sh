#!/usr/bin/env bash
# Re-vendor the community Claude skills from upstream at pinned commits.
#
# Each pinned ref below is bumped by Renovate (git-refs digest datasource).
# A Renovate postUpgradeTask then re-runs this script so the same PR carries
# the refreshed skill content, not just a new SHA string.
#
# The file set for each skill is derived from the upstream tree at the pinned
# ref, not hardcoded here. A row naming `*` as its skill takes every skill under
# that subpath, so a skill added or removed upstream is mirrored too (ADR-0007).
# A vendored directory no longer produced by any row is pruned.
#
# Adding a skill: append a "repo|subpath|skill" row to SKILLS. If it comes from
# a repo not yet listed, also add a renovate-commented <NAME>_REF for that repo
# and a case arm in ref_for_repo. A locally-authored skill instead goes in the
# `vendored` exclude's lookahead in .pre-commit-config.yaml.
#
# Requires: curl, jq.
# Run manually:  bash scripts/update-vendored-skills.sh
set -euo pipefail

# renovate: datasource=git-refs depName=mattpocock/skills packageName=https://github.com/mattpocock/skills branch=main
MATTPOCOCK_SKILLS_REF="4ae2317ddb22b046e26455e4f9c9b3aaaf62212d"

# renovate: datasource=git-refs depName=cloudflare/security-audit-skill packageName=https://github.com/cloudflare/security-audit-skill branch=main
CLOUDFLARE_SECURITY_AUDIT_REF="c1c8a8c1471069fb0e188eeaff69b8e8db6564a8"

# renovate: datasource=git-refs depName=JuliusBrussee/caveman packageName=https://github.com/JuliusBrussee/caveman branch=main
CAVEMAN_REF="4bf8898b31c93326a5e5d90999241a1c15b52ed2"

# renovate: datasource=git-refs depName=AminBlg/SimpleEnglish packageName=https://github.com/AminBlg/SimpleEnglish branch=main
SIMPLE_ENGLISH_REF="a6fcb4fde098b33617cc1578d151ebf58774b883"

# renovate: datasource=git-refs depName=DietrichGebert/ponytail packageName=https://github.com/DietrichGebert/ponytail branch=main
PONYTAIL_REF="96608e078b2d5ea607c5c3b8b5a54d1c1fba6846"

# Managed files live under home/ (see .chezmoiroot), so the skills tree is
# home/dot_claude/skills, not dot_claude/skills at the repo root.
#
# These, and SKILLS_ROWS below, are overridable so
# scripts/test-update-vendored-skills.sh can run offline against file:// fixtures.
SKILLS_DIR="${SKILLS_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/home/dot_claude/skills}"
GITHUB_API="${GITHUB_API:-https://api.github.com}"
GITHUB_RAW="${GITHUB_RAW:-https://raw.githubusercontent.com}"
RETRIEVED="$(date -u +%Y-%m-%d)"

# The vendored set, as "repo|subpath|skill", where skill `*` means every skill
# under subpath. The subpath is where the skill dirs live upstream; it differs
# per repo and even per skill within a repo. mattpocock's in-progress/ and
# deprecated/ are left out on purpose: upstream marks them as not ready.
SKILLS=(
  "mattpocock/skills|skills/engineering|*"
  "mattpocock/skills|skills/misc|*"
  "mattpocock/skills|skills/productivity|*"
  "cloudflare/security-audit-skill|skills|security-audit"
  "JuliusBrussee/caveman|skills|caveman-compress"
  "AminBlg/SimpleEnglish|skills|simple-english"
  "DietrichGebert/ponytail|skills|*"
)
if [[ -n "${SKILLS_ROWS:-}" ]]; then
  IFS=$'\n' read -r -d '' -a SKILLS <<<"$SKILLS_ROWS" || true
fi

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
    AminBlg/SimpleEnglish) echo "$SIMPLE_ENGLISH_REF" ;;
    DietrichGebert/ponytail) echo "$PONYTAIL_REF" ;;
    *) echo "no pinned ref for repo: $1" >&2; return 1 ;;
  esac
}

cache_key() {
  # $1 = repo slug -> echoes a filename-safe key.
  echo "${1//\//_}"
}

# Fetch $1 to $2, materialising the file only on success. curl -f still creates
# (and leaves behind) a truncated or empty output file when the request fails,
# which the `-f` cache tests below would then take for a valid hit.
#
# Every caller propagates failure with an explicit `|| return 1` rather than
# leaning on `set -e`: these run inside nested command substitutions, where errexit
# does not reliably abort at the failing command, and an unnoticed fetch failure is
# exactly the bug this script is being hardened against.
fetch_to_cache() {
  if ! curl -fsSL "$1" -o "$2.part"; then
    rm -f "$2.part"
    echo "fetch failed: $1" >&2
    return 1
  fi
  mv "$2.part" "$2"
}

tree_json() {
  # $1 = repo slug, $2 = ref -> echoes the path to the cached tree JSON.
  local cache
  cache="${CACHE_DIR}/$(cache_key "$1").tree.json"
  if [[ ! -f "$cache" ]]; then
    # shellcheck disable=SC2310
    fetch_to_cache "${GITHUB_API}/repos/$1/git/trees/$2?recursive=1" "$cache" || return 1
  fi
  echo "$cache"
}

license_text() {
  # $1 = repo slug, $2 = ref -> echoes the path to the cached LICENSE.
  local cache
  cache="${CACHE_DIR}/$(cache_key "$1").LICENSE"
  if [[ ! -f "$cache" ]]; then
    # shellcheck disable=SC2310
    fetch_to_cache "${GITHUB_RAW}/$1/$2/LICENSE" "$cache" || return 1
  fi
  echo "$cache"
}

write_upstream_md() {
  # $1 = repo, $2 = subpath, $3 = skill, $4 = ref, $5 = destination dir,
  # $6 = retrieval date
  # The licence is embedded verbatim from upstream so the copyright line always
  # matches the repo the files actually came from.
  {
    cat <<EOF
# Upstream provenance

Vendored from [$1](https://github.com/$1) — \`$2/$3\`.

- Pinned commit: \`$4\`
- Retrieved: $6
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
  local tree
  # shellcheck disable=SC2310
  tree="$(tree_json "$1" "$4")" || return 1
  jq -r --arg p "$2/$3/" \
    '.tree[] | select(.type == "blob") | select(.path | startswith($p)) | .path[($p | length):]' \
    "$tree"
}

expand_skills() {
  # $1 = repo, $2 = subpath, $3 = skill or `*` -> echoes one skill name per line.
  # A named skill is checked against the tree too, so every fetch that can fail
  # does so before anything is wiped.
  local tree ref names
  # shellcheck disable=SC2310
  ref="$(ref_for_repo "$1")" || return 1
  # shellcheck disable=SC2310
  tree="$(tree_json "$1" "$ref")" || return 1
  names="$(jq -r --arg p "$2/" --arg want "$3" \
    '.tree[] | select(.type == "blob") | .path | select(startswith($p)) | .[($p | length):]
     | select(test("^[^/]+/SKILL\\.md$")) | split("/")[0]
     | select($want == "*" or . == $want)' \
    "$tree")" || return 1
  if [[ -z "$names" ]]; then
    echo "no skill matching '$3' under $1/$2 at ${ref}" >&2
    return 1
  fi
  echo "$names"
}

vendor_skill() {
  # $1 = repo, $2 = subpath, $3 = skill
  local repo="$1" subpath="$2" skill="$3" ref dest listing f retrieved staging
  ref="$(ref_for_repo "$repo")"
  dest="${SKILLS_DIR}/${skill}"

  # Resolve the file list *before* the wipe below, and via a command substitution
  # rather than `done < <(...)`: a process substitution's exit status is invisible
  # to `set -e`, so a failed tree fetch used to yield zero lines, skip the loop
  # entirely, and leave the script reporting success over a directory it had
  # already emptied. Renovate runs this unattended and commits the result, so that
  # failure mode is an auto-merged PR that deletes a skill. An assignment to a
  # pre-declared variable does propagate the failure (declaring and assigning on
  # one line would not: `local` would supply its own exit status).
  listing="$(list_skill_files "$repo" "$subpath" "$skill" "$ref")"
  if [[ -z "$listing" ]]; then
    echo "no files under ${repo}/${subpath}/${skill} at ${ref}" >&2
    return 1
  fi

  # An unchanged pin keeps its date: otherwise every Renovate bump of one repo
  # would rewrite the UPSTREAM.md of every skill from every other repo too.
  retrieved=""
  if [[ -f "$dest/UPSTREAM.md" ]] && grep -qF -- "- Pinned commit: \`${ref}\`" "$dest/UPSTREAM.md"; then
    retrieved="$(sed -n 's/^- Retrieved: //p' "$dest/UPSTREAM.md")"
  fi
  retrieved="${retrieved:-$RETRIEVED}"

  # Fetch into a staging dir and swap it in whole, so a file removed upstream does
  # not linger and a failed download leaves the previous copy in place. The
  # staging dir sits outside SKILLS_DIR so the prune pass can never see it.
  staging="${CACHE_DIR}/staging/${skill}"
  rm -rf "$staging"
  mkdir -p "$staging"
  while IFS= read -r f; do
    echo "fetching ${repo}/${subpath}/${skill}/${f}"
    mkdir -p "${staging}/$(dirname "$f")"
    curl -fsSL "${GITHUB_RAW}/${repo}/${ref}/${subpath}/${skill}/${f}" \
      -o "${staging}/${f}"
  done <<<"$listing"
  write_upstream_md "$repo" "$subpath" "$skill" "$ref" "$staging" "$retrieved"
  rm -rf "$dest"
  mv "$staging" "$dest"
}

# Resolve every row before touching the tree, so a failed fetch or a name clash
# aborts with nothing wiped.
resolved=()
for entry in "${SKILLS[@]}"; do
  IFS='|' read -r repo subpath skill <<<"$entry"
  # shellcheck disable=SC2310
  names="$(expand_skills "$repo" "$subpath" "$skill")" || exit 1
  while IFS= read -r name; do
    # The wipe in vendor_skill must never reach a locally-authored skill: a
    # wildcard row can surface an upstream skill that shares its name.
    if [[ -d "${SKILLS_DIR}/${name}" && ! -f "${SKILLS_DIR}/${name}/UPSTREAM.md" ]]; then
      echo "refusing to overwrite locally-authored skill: ${name}" >&2
      exit 1
    fi
    resolved+=("${repo}|${subpath}|${name}")
  done <<<"$names"
done

dupes="$(printf '%s\n' "${resolved[@]}" | cut -d'|' -f3 | sort | uniq -d)"
if [[ -n "$dupes" ]]; then
  echo "skill name vendored from more than one place: ${dupes}" >&2
  exit 1
fi

for entry in "${resolved[@]}"; do
  IFS='|' read -r repo subpath skill <<<"$entry"
  vendor_skill "$repo" "$subpath" "$skill"
done

# A skill dropped upstream, or from SKILLS, would otherwise linger forever.
# UPSTREAM.md is what marks a directory as vendored rather than locally authored.
for marker in "${SKILLS_DIR}"/*/UPSTREAM.md; do
  [[ -f "$marker" ]] || continue
  dir="$(dirname "$marker")"
  if ! printf '%s\n' "${resolved[@]}" | cut -d'|' -f3 | grep -qFx "$(basename "$dir")"; then
    echo "pruning $(basename "$dir"): no longer vendored"
    rm -rf "$dir"
  fi
done

echo "done: vendored ${#resolved[@]} skills"
